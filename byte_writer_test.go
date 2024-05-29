package ack_chan

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"io"
	"sync"
	"testing"
	"time"
)

func TestAckableByteWriter_Write(test *testing.T) {
	t := assert.New(test)
	const RepeatTimes = 1024 * 32
	bufferReader := new(bytes.Buffer)
	bufferWriter := new(bytes.Buffer)
	for range RepeatTimes {
		for i := range 256 {
			bufferReader.WriteByte(byte(i))
		}
	}
	expectedBytes := bufferReader.Bytes()
	expectedWrittenBytes := bufferReader.Len()
	bufferWriter.Grow(expectedWrittenBytes)
	w := NewAckableByteWriter()
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		defer wg.Done()
		defer w.Close()
		n, err := io.Copy(w, bufferReader)
		t.NoError(err)
		t.EqualValues(expectedWrittenBytes, n)
	}()
	go func() {
		defer wg.Done()
		for {
			b, err := w.Receive(context.Background())
			if err == io.EOF {
				return
			}
			t.NoError(err)
			w.Ack()
			bufferWriter.Write(b)
		}
	}()
	wg.Wait()
	t.Equal(expectedBytes, bufferWriter.Bytes())
}

func TestAckableByteWriter_CloseOnWrite(test *testing.T) {
	t := assert.New(test)
	w := NewAckableByteWriter()
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(time.Millisecond * 10)
		t.NoError(w.Close())
	}()
	n, err := w.Write([]byte{1, 2, 3})
	wg.Wait()
	t.EqualValues(0, n)
	t.ErrorIs(err, io.EOF)
}

func TestAckableByteWriter_CloseAfterWrite(test *testing.T) {
	t := assert.New(test)
	w := NewAckableByteWriter()
	t.NoError(w.Close())
	n, err := w.Write([]byte{1, 2, 3})
	t.EqualValues(0, n)
	t.ErrorIs(err, io.EOF)
}
