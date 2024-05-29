package ack_chan

import (
	"context"
	"io"
)

// AckableByteWriter is a wrapper around AckableChan[[]byte] which
// implements the io.Writer interface. It is convenient for scenarios which
// a data must be passed to another goroutine and processed which can fail.
type AckableByteWriter struct {
	AckableChan[[]byte]
}

// NewAckableByteWriter will create a new ackable channel which implements io.Writer interface
func NewAckableByteWriter() *AckableByteWriter {
	return &AckableByteWriter{*NewAckableChan[[]byte]()}
}

// Write will write data to the channel.
//
// Will return io.EOF if the channel is closed.
func (c *AckableByteWriter) Write(data []byte) (n int, err error) {
	// Note: I have not found a better way to do this.
	// The io.Writer interface does not have a context. This means that
	// we have to pass context.background to the channel.
	// However, if the reader side closes the channel mid write, we will get a panic.
	//
	// In this case, we simply defer a recover here.
	// This will decrease the performance, but I really don't know any better way
	// except implementing a channel which closes when the AckableChan closes.
	// With that, we can select between sending data and reading from closed channel.
	defer func() {
		if recover() != nil {
			n = 0
			err = io.EOF
		}
	}()
	err = c.Send(context.Background(), data)
	if err == nil {
		n = len(data)
	}
	return n, err
}
