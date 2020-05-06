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

package executor

import (
	"context"

	"go.uber.org/zap"

	"github.com/kapetaniosci/pipe/pkg/config"
	"github.com/kapetaniosci/pipe/pkg/model"
)

type Executor interface {
	Execute(ctx context.Context) (model.StageStatus, error)
}

type Factory func(in Input) Executor

type Input struct {
	Stage             *config.PipelineStage
	AppConfig         *config.Config
	WorkingDir        string
	Deployment        *model.Deployment
	LogPersister      LogPersister
	MetadataPersister MetadataPersister
	Logger            *zap.Logger
}

type LogPersister interface {
	Append(log string)
}

type MetadataPersister interface {
	Save(metadata []byte) error
}