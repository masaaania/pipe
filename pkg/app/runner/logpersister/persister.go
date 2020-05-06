// Copyright 2020 The PipeCD Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logpersister

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/kapetaniosci/pipe/pkg/app/api/service/runnerservice"
	"github.com/kapetaniosci/pipe/pkg/model"
)

type apiClient interface {
	ReportStageLog(ctx context.Context, in *runnerservice.ReportStageLogRequest, opts ...grpc.CallOption) (*runnerservice.ReportStageLogResponse, error)
}

type Persister interface {
	Run(ctx context.Context) error
	StageLogPersister(deploymentID, stageID string) StageLogPersister
}

type StageLogPersister interface {
	Append(log string)
	Complete(ctx context.Context) error
}

type key struct {
	DeploymentID string
	StageID      string
}

type persister struct {
	apiClient       apiClient
	stagePersisters sync.Map
	flushInterval   time.Duration
	stalePeriod     time.Duration
	gracePeriod     time.Duration
	logger          *zap.Logger
}

// NewPersister creates a new persister instance for saving the stage logs into server's storage.
// This controls how many concurent api calls should be executed and when to flush the logs.
func NewPersister(apiClient apiClient, logger *zap.Logger) Persister {
	return &persister{
		apiClient:     apiClient,
		flushInterval: 10 * time.Second,
		stalePeriod:   time.Minute,
		gracePeriod:   30 * time.Second,
		logger:        logger.Named("logger-persister"),
	}
}

// Run starts running workers to flush logs to server.
func (p *persister) Run(ctx context.Context) error {
	p.logger.Info("start running log persister")
	ticker := time.NewTicker(p.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.flush(ctx)
		case <-ctx.Done():
			break
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), p.gracePeriod)
	defer cancel()
	p.flush(ctx)

	p.logger.Info("log persister has been stopped")
	return nil
}

func (p *persister) flush(ctx context.Context) {
	completedKeys := make([]key, 0)

	// Check for new log entries and flush if needed.
	p.stagePersisters.Range(func(_, v interface{}) bool {
		sp := v.(*stageLogPersister)
		if completed := p.flushStage(ctx, sp); completed {
			completedKeys = append(completedKeys, sp.key)
		}
		return false
	})

	// Clean up all completed stages.
	for _, k := range completedKeys {
		p.stagePersisters.Delete(k)
	}
}

func (p *persister) flushStage(ctx context.Context, sp *stageLogPersister) bool {
	sp.mu.Lock()
	// Stage was completed and no more retries.
	if sp.completed && sp.retries <= 0 {
		deletable := time.Since(sp.completedAt) > p.stalePeriod
		sp.mu.Unlock()
		return deletable
	}

	sp.mu.Lock()
	var (
		blocks     = sp.blocks
		blockCount = sp.index
		completed  = sp.completed
	)
	sp.mu.Unlock()
	// Flush all current blocks at the local.
	if err := p.reportStageLog(ctx, sp.key, blocks, completed, blockCount); err != nil {
		p.logger.Error("failed to report stage log",
			zap.Any("key", sp.key),
			zap.Error(err),
		)
		return false
	}

	// Remove all sent blocks.
	sp.mu.Lock()
	sp.blocks = sp.blocks[len(blocks):]
	sp.mu.Unlock()

	return false
}

func (p *persister) reportStageLog(ctx context.Context, k key, blocks []*model.LogBlock, completed bool, blockCount int) error {
	req := &runnerservice.ReportStageLogRequest{
		DeploymentId:    k.DeploymentID,
		StageId:         k.StageID,
		Blocks:          blocks,
		TotalBlockCount: int64(blockCount),
		Completed:       completed,
	}
	_, err := p.apiClient.ReportStageLog(ctx, req)
	return err
}

// StageLogPersister creates a child persister instance for a specific stage.
func (p *persister) StageLogPersister(deploymentID, stageID string) StageLogPersister {
	var (
		k = key{
			DeploymentID: deploymentID,
			StageID:      stageID,
		}
		sp = &stageLogPersister{
			key:       k,
			persister: p,
		}
	)
	p.stagePersisters.Store(k, sp)
	return sp
}

// stageLogPersister represents a log persister for a specific stage.
type stageLogPersister struct {
	key         key
	index       int
	blocks      []*model.LogBlock
	completed   bool
	completedAt time.Time
	retries     int
	mu          sync.Mutex
	persister   *persister
}

// Append appends a new log block.
func (sp *stageLogPersister) Append(log string) {
	now := time.Now()
	sp.mu.Lock()
	defer sp.mu.Unlock()

	sp.blocks = append(sp.blocks, &model.LogBlock{
		Index:     int64(sp.index),
		Log:       log,
		CreatedAt: now.Unix(),
	})
	sp.index++
}

// Complete marks the completion of logging for this stage.
// This means no more log for this stage will be added into this persister.
func (sp *stageLogPersister) Complete(ctx context.Context) error {
	sp.mu.Lock()
	var (
		blocks     = sp.blocks
		blockCount = sp.index
	)
	sp.completed = true
	sp.completedAt = time.Now()
	sp.mu.Unlock()

	// Immediately send the log to the server.
	err := sp.persister.reportStageLog(ctx, sp.key, blocks, true, blockCount)
	if err == nil {
		return nil
	}

	// If the log was not sent to the server successfully,
	// we should retry them later.
	sp.mu.Lock()
	sp.retries = 3
	sp.mu.Unlock()

	return err
}