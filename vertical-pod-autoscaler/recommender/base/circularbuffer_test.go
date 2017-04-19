package base

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCircularBuffer(t *testing.T) {
	b := NewCircularBuffer(3)
	overflow, discarded := b.Push(1.0)
	assert.False(t, overflow)
	assert.Equal(t, 1.0, *b.Head())
	assert.Equal(t, []float64{1.0}, b.Contents())

	overflow, discarded = b.Push(2.0)
	assert.False(t, overflow)
	assert.Equal(t, 2.0, *b.Head())
	assert.Equal(t, []float64{1.0, 2.0}, b.Contents())

	overflow, discarded = b.Push(3.0)
	assert.False(t, overflow)
	assert.Equal(t, 3.0, *b.Head())
	assert.Equal(t, []float64{1.0, 2.0, 3.0}, b.Contents())

	overflow, discarded = b.Push(4.0)
	assert.True(t, overflow)
	assert.Equal(t, discarded, 1.0)
	assert.Equal(t, 4.0, *b.Head())
	assert.Equal(t, []float64{2.0, 3.0, 4.0}, b.Contents())
}

