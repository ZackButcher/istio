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

package v2

import (
	"crypto/sha1"
	"fmt"
)

const (
	// MaxClusterNameLength is the maximum cluster name length
	MaxClusterNameLength = 189
)

func truncateClusterName(name string, ctx context) string {
	maxLen := MaxClusterNameLength
	if ctx.Mesh.DefaultConfig.StatNameLength > 0 {
		maxLen = int(ctx.Mesh.DefaultConfig.StatNameLength)
	}

	if len(name) > maxLen {
		prefix := name[:maxLen-sha1.Size*2]
		sum := sha1.Sum([]byte(name))
		return fmt.Sprintf("%s%x", prefix, sum)
	}
	return name
}
