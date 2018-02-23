// Copyright 2018 Istio Authors.
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

package proto

import (
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"
	"istio.io/istio/pkg/log"
)

func MessageToStruct(p proto.Message) (s *types.Struct, err error) {
	bytes, err := proto.Marshal(p)
	if err != nil {
		return nil, err
	}
	err = s.Unmarshal(bytes)
	return
}

func ToDuration(d *duration.Duration) time.Duration {
	if d == nil {
		return 0
	}

	dur, err := ptypes.Duration(d)
	if err != nil {
		log.Warnf("error converting duration %#v, using 0: %v", d, err)
	}
	return dur
}

func DurationToMs(d *duration.Duration) int64 {
	return int64(ToDuration(d) / time.Millisecond)
}
