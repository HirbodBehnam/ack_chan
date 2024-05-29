package ack_chan

import (
	"context"
	"github.com/stretchr/testify/assert"
	"io"
	"sync"
	"testing"
	"time"
)

func TestAckableChan_SendReceive(test *testing.T) {
	t := assert.New(test)
	const N = 1234
	c := NewAckableChan[int]()
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := range N {
			t.NoError(c.Send(context.Background(), i))
		}
	}()
	go func() {
		defer wg.Done()
		for i := range N {
			n, err := c.Receive(context.Background())
			t.NoError(err)
			t.Equal(i, n)
			c.Ack()
		}
	}()
	wg.Wait()
}

func TestAckableChan_Ack(test *testing.T) {
	t := assert.New(test)
	const NumberToSend = 1234567
	const NumberOfTries = 1234
	c := NewAckableChan[int]()
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range NumberOfTries {
			n, err := c.Receive(context.Background())
			t.NoError(err)
			t.Equal(NumberToSend, n)
		}
	}()
	t.NoError(c.Send(context.Background(), NumberToSend))
	wg.Wait()
}

func TestAckableChan_AckUnexpected(t *testing.T) {
	c := NewAckableChan[int]()
	assert.PanicsWithValue(t, "unexpected ack", c.Ack)
}

func TestAckableChan_Context(test *testing.T) {
	t := assert.New(test)
	const WaitTime = time.Millisecond * 10
	c := NewAckableChan[int]()
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), WaitTime)
		defer cancel()
		t.ErrorIs(c.Send(ctx, 0), context.DeadlineExceeded)
	}()
	go func() {
		defer wg.Done()
		time.Sleep(WaitTime * 2)
		ctx, cancel := context.WithTimeout(context.Background(), WaitTime)
		defer cancel()
		_, err := c.Receive(ctx)
		t.ErrorIs(err, context.DeadlineExceeded)
	}()
	wg.Wait()
}

func TestAckableChan_EarlyClose(test *testing.T) {
	t := assert.New(test)
	c := NewAckableChan[int]()
	t.NoError(c.Close())
	t.NoError(c.Close())
	t.ErrorIs(c.Send(context.Background(), 0), io.EOF)
	_, err := c.Receive(context.Background())
	t.ErrorIs(err, io.EOF)
}

func TestAckableChan_MidRecvClose(test *testing.T) {
	t := assert.New(test)
	c := NewAckableChan[int]()
	const WaitTime = time.Millisecond * 10
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := c.Receive(context.Background())
		t.ErrorIs(err, io.EOF)
	}()
	time.Sleep(WaitTime)
	t.NoError(c.Close())
	wg.Wait()
}
