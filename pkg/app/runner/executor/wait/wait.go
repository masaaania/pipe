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

package wait

import (
	"context"
	"fmt"
	"time"

	"github.com/kapetaniosci/pipe/pkg/app/runner/executor"
	"github.com/kapetaniosci/pipe/pkg/model"
)

var defaultDuration = time.Minute

type Executor struct {
	executor.Input
}

func init() {
	var (
		f = func(in executor.Input) executor.Executor {
			return &Executor{
				Input: in,
			}
		}
		r = executor.DefaultRegistry()
	)
	r.Register(model.StageWait, f)
}

func (e *Executor) Execute(ctx context.Context) (model.StageStatus, error) {
	duration := defaultDuration
	timer := time.NewTimer(duration)
	defer timer.Stop()

	e.LogPersister.Append(fmt.Sprintf("Waiting for %v...", duration))
	select {
	case <-timer.C:
	case <-ctx.Done():
		return model.StageStatus_STAGE_CANCELLED, fmt.Errorf("context cancelled")
	}
	e.LogPersister.Append(fmt.Sprintf("Waited for %v", duration))

	return model.StageStatus_STAGE_SUCCESS, nil
}