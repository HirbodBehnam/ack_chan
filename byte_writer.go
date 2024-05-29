package ack_chan

import (
	"context"
	"io"
)

type AckableByteWriter struct {
	AckableChan[[]byte]
}

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
