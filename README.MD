# Ack Channel

[![Run Tests](https://github.com/HirbodBehnam/ack-chan/actions/workflows/test.yml/badge.svg)](https://github.com/HirbodBehnam/ack-chan/actions/workflows/test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/HirbodBehnam/ack-chan.svg)](https://pkg.go.dev/github.com/HirbodBehnam/ack-chan)

Generic channels in Golang which the receiver must acknowledge every received data.

## Features

* Generic Channels
* Made for Multiple Senders, Single Receiver
* Almost as fast as using raw channels

## Documentation

The documentation is available at [GoDoc](https://pkg.go.dev/github.com/HirbodBehnam/ack-chan).

## Example Usage

A very simple usage can be seen in the tests which looks like this:

```go
c := NewAckableChan[int]()
go func() {
	for i := range N {
		c.Send(context.Background(), i)
	}
}()
go func() {
	for i := range N {
		n, _ := c.Receive(context.Background())
		fmt.Println("Got", n)
		c.Ack()
	}
}()
```