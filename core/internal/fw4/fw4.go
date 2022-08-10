package fw4

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel/trace"
)

const maxBlobLen = 0x4000

const (
	emptyEntryAgentId int32 = -1
	emptyEntryTagId   int32 = -1
)

const (
	linkIdMask        uint32 = ^uint32(0b11111 << 27)
	linkIdIgnoredMask uint32 = 1 << 31
)

type tlvId int

const (
	tlvIdCustomBlob tlvId = iota + 1
	tlvIdTagDepth
	tlvIdEntryAgentId
	tlvIdEntryTagId
	tlvIdPayloadBitset
	tlvIdTraceId
	tlvIdSpanId
)

type Fw4Tag struct {
	agentID       int32
	tagID         int32
	encodedLinkID uint32
	ServerID      int32
	ClusterID     int32
	TenantID      int32
	PathInfo      uint32

	// Extension fields
	CustomBlob    string
	tagDepth      int32
	entryAgentID  int32
	entryTagID    int32 // Problem: -1 is the "zero value" for this.
	payloadBitset int32
	TraceID       trace.TraceID
	SpanID        trace.SpanID
}

func EmptyTag() Fw4Tag {
	return Fw4Tag{entryAgentID: emptyEntryAgentId, entryTagID: emptyEntryTagId}
}

func (fw4 Fw4Tag) HasEntryAgentID() bool {
	return fw4.entryAgentID != -1
}

func (fw4 Fw4Tag) HasEntryTagID() bool {
	return fw4.entryTagID >= 0
}

func (fw4 Fw4Tag) HasTagDepth() bool {
	return fw4.tagDepth > 0
}

func (fw4 Fw4Tag) linkID() int32 {
	return int32(fw4.encodedLinkID & linkIdMask)
}

func (fw4 Fw4Tag) IsIgnored() bool {
	return (fw4.encodedLinkID & linkIdIgnoredMask) != 0
}

func (fw4 Fw4Tag) samplingExponent() int32 {
	return int32((fw4.encodedLinkID >> 27) & 0b1111)
}

func (fw4 Fw4Tag) String() string {
	var sb strings.Builder

	fmt.Fprintf(
		&sb,
		"FW4{a=%v;t=%v;L=%v;s=%v;c=%v;t=%v;p=%v",
		fw4.agentID,
		fw4.tagID,
		fw4.encodedLinkID,
		fw4.ServerID,
		fw4.ClusterID,
		fw4.TenantID,
		fw4.PathInfo)

	if len(fw4.CustomBlob) > 0 {
		sb.WriteString("+b=")
		hex.NewEncoder(&sb).Write([]byte(fw4.CustomBlob)) //nolint:errcheck
	}

	if fw4.tagDepth != 0 {
		fmt.Fprint(&sb, "+d=", fw4.tagDepth)
	}

	if fw4.HasEntryAgentID() {
		fmt.Fprint(&sb, "+A=", fw4.entryAgentID)
	}

	if fw4.entryTagID != -1 { // Don't use HasEntryTagID to display non-canonical values
		fmt.Fprint(&sb, "+e=", fw4.entryTagID)
	}

	if fw4.payloadBitset != 0 {
		fmt.Fprint(&sb, "+P=", fw4.payloadBitset)
	}

	if fw4.TraceID.IsValid() {
		fmt.Fprint(&sb, "+T=", fw4.TraceID)
	}

	if fw4.SpanID.IsValid() {
		fmt.Fprint(&sb, "+S=", fw4.SpanID)
	}
	sb.WriteByte('}')
	return sb.String()
}

// FW4Tag instantiation

// Used to generate FwTag.PathInfo in a thread-safe way.
type pathInfoGenerator struct {
	randomNumberGenerator *rand.Rand
	mutex                 sync.Mutex
}

func (p *pathInfoGenerator) generatePathInfo() uint32 {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	// PathInfo must be an unsigned 32 bit integer
	// whose lowest 8 bits are a pseudo-random number in the range [0, 255]
	return uint32(p.randomNumberGenerator.Intn(256))
}

var pathInfoGeneratorInstance = pathInfoGenerator{
	randomNumberGenerator: rand.New(rand.NewSource(time.Now().UnixNano())),
	mutex:                 sync.Mutex{},
}

func NewFw4Tag(clusterId, tenantId int32, spanContext trace.SpanContext) *Fw4Tag {
	tag := EmptyTag()
	tag.ClusterID = clusterId
	tag.TenantID = tenantId
	tag.PathInfo = pathInfoGeneratorInstance.generatePathInfo()
	tag.TraceID = spanContext.TraceID()
	tag.SpanID = spanContext.SpanID()
	return &tag
}

// Context-related FW4Tag utilities

type fw4TagKeyType int

const fw4TagKey fw4TagKeyType = iota

func ContextWithFw4Tag(ctx context.Context, fw4Tag *Fw4Tag) context.Context {
	return context.WithValue(ctx, fw4TagKey, fw4Tag)
}

func Fw4TagFromContext(ctx context.Context) *Fw4Tag {
	if fw4Tag, ok := ctx.Value(fw4TagKey).(*Fw4Tag); ok {
		return fw4Tag
	}
	return nil
}
