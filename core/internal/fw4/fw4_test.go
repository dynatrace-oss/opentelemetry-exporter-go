package fw4

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFw4TagFromContext_IsNil(t *testing.T) {
	ctx := context.Background()
	tag := Fw4TagFromContext(ctx)
	assert.Nil(t, tag)
}

func TestGetFw4TagFromContext_IsNilWhenNilIsAssigned(t *testing.T) {
	ctx := ContextWithFw4Tag(context.Background(), nil)
	tag := Fw4TagFromContext(ctx)
	assert.Nil(t, tag)
}

func TestGetFw4TagFromContext_IsNotNil(t *testing.T) {
	tag := EmptyTag()
	ctx := ContextWithFw4Tag(context.Background(), &tag)

	tagFromContext := Fw4TagFromContext(ctx)
	assert.NotNil(t, tagFromContext)
	assert.Equal(t, &tag, tagFromContext)
}
