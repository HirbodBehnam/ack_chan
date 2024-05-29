package ack_chan

import (
	"context"
	"io"
	"sync/atomic"
)

// AckableChan enables you to receive the value which is in the chan but
// must choose to not to "read" the data.
//
// However, this channel is multiple producer single consumer. The consumer must not
// be shared between multiple readers.
type AckableChan[T any] struct {
	// The channel which we send and receive data from it
	c chan T
	// Last data which is read but not acked
	lastData T
	// Do we have unacked data in the channel (read but not acked)
	unackedData bool
	// Is this channel closed?
	closed atomic.Bool
}

// NewAckableChan creates a new ackable channel
func NewAckableChan[T any]() *AckableChan[T] {
	return &AckableChan[T]{
		c: make(chan T),
	}
}

// Receive a data from channel. This data must be acked.
//
// The error can have three different values.
//
// 1. io.EOF means that the channel is closed.
// 2. context can return the inner error as well.
func (c *AckableChan[T]) Receive(ctx context.Context) (T, error) {
	// Check if we have unacked data
	if c.unackedData {
		return c.lastData, nil
	}
	// Otherwise, get the data from channel
	var defaultValue T
	select {
	case data, ok := <-c.c: // got new data
		if ok {
			c.unackedData = true
			c.lastData = data
			return data, nil
		} else { // channel closed
			return defaultValue, io.EOF
		}
	case <-ctx.Done():
		return defaultValue, ctx.Err()
	}
}

// Send will send a data in the channel with cancellation.
//
// Will either return io.EOF as error if the channel is closed or
// the error of canceled context.
func (c *AckableChan[T]) Send(ctx context.Context, data T) error {
	// Check if channel is closed
	if c.closed.Load() {
		return io.EOF
	}
	// Send data or wait
	select {
	case c.c <- data:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Ack will ack a received packet.
// Ack will panic if there is no data to ack.
func (c *AckableChan[T]) Ack() {
	if c.unackedData {
		c.unackedData = false
		var defaultValue T
		c.lastData = defaultValue
	} else {
		panic("unexpected ack")
	}
}

// Close will close the channel.
//
// WARNING: Cancel all the contexts of senders before calling this method to avoid panics.
func (c *AckableChan[T]) Close() error {
	if c.closed.CompareAndSwap(false, true) {
		close(c.c)
	}
	return nil
}
