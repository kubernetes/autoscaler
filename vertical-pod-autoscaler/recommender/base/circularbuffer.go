package base

// A queue with a fixed maximum capacity. Elements are inserted/removed in the
// FIFO order. Elements are removed from the buffer when it runs out of capacity
// and a new element is inserted.
type CircularBuffer interface {
	// Add a value to the end of the queue. On overflow returns true and the
	// oldest value, which is also removed from the buffer. Otherwise returns
	// (false, _).
	Push(value float64) (bool, float64)

	// Returns the elements in the buffer, ordered by time of insertion (oldest
	// first).
	Contents() []float64

	// Returns a pointer to the most recently added element. The pointer can be
	// used to modify the last element. It is only valid until the next call to
	// Push().
	Head() *float64
}

type CircularBufferImpl struct {
	buffer []float64
	head int	// Index of the most recently added element.
	capacity int	// Max number of elements.
	isFull bool	// The number of elements in the buffer equals capacity.
}

func NewCircularBuffer(size int) *CircularBufferImpl {
	if (size < 1) {
		panic("Buffer size must be at least 1")
	}
	return &CircularBufferImpl{make([]float64, 0), -1, size, false}
}

func (b *CircularBufferImpl) Head() *float64 {
	if (b.head == -1) {
		panic("Called Head() on an empty buffer")
	}
	return &b.buffer[b.head]
}

func (b *CircularBufferImpl) Contents() []float64 {
	return append(b.buffer[b.head+1:], b.buffer[:b.head+1]...)
}

func (b *CircularBufferImpl) Push(value float64) (bool, float64) {
	b.head++
	if (b.head == b.capacity) {
		b.head = 0
		b.isFull = true
	}
	if (!b.isFull) {
		b.buffer = append(b.buffer, value)
		return false, 0.0
	} else {
		prevValue := b.buffer[b.head]
		b.buffer[b.head] = value
		return true, prevValue
	}
}

