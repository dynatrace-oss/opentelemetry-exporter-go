package trace

import "context"

type serverIdKeyType int
const serverIdKey serverIdKeyType = iota

// This will be called from the Dynatrace propagator
//nolint:deadcode,unused
func setServerIdOnContext(ctx context.Context, serverId int32) context.Context {
	return context.WithValue(ctx, serverIdKey, serverId)
}

func getServerIdFromContext(ctx context.Context) int32 {
	if serverId, ok := ctx.Value(serverIdKey).(int32); ok {
		return serverId
	}
	return 0
}
