// Copyright 2022 Dynatrace LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import "context"

type serverIdKeyType int

const serverIdKey serverIdKeyType = iota

// This will be called from the Dynatrace propagator
//nolint:deadcode,unused
func SetServerIdOnContext(ctx context.Context, serverId int32) context.Context {
	return context.WithValue(ctx, serverIdKey, serverId)
}

func GetServerIdFromContext(ctx context.Context) int32 {
	if serverId, ok := ctx.Value(serverIdKey).(int32); ok {
		return serverId
	}
	return 0
}
