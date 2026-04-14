package net

import (
	"bytes"
	"io"
	stdnet "net"
	"testing"
)

type nopReadWriteCloser struct {
	io.Reader
	io.Writer
}

func (n *nopReadWriteCloser) Close() error { return nil }

func TestAsymmetricStreamReadWrite(t *testing.T) {
	c1, c2 := stdnet.Pipe()
	defer c1.Close()
	defer c2.Close()

	clientSide := NewAsymmetricStream(c1, func(payloadLen int) int {
		return payloadLen / 10
	})
	serverSide := NewAsymmetricStream(c2, nil)

	payload := bytes.Repeat([]byte("x"), 4096)
	go func() {
		_, _ = clientSide.Write(payload)
	}()

	got := make([]byte, len(payload))
	if _, err := io.ReadFull(serverSide, got); err != nil {
		t.Fatalf("read payload error: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("payload mismatch")
	}
}

func TestAsymmetricStreamDiscardPadding(t *testing.T) {
	var wire bytes.Buffer
	writer := NewAsymmetricStream(&nopReadWriteCloser{Reader: &wire, Writer: &wire}, func(payloadLen int) int {
		return 32
	})
	reader := NewAsymmetricStream(&nopReadWriteCloser{Reader: &wire, Writer: io.Discard}, nil)

	payload := []byte("hello-frp")
	if _, err := writer.Write(payload); err != nil {
		t.Fatalf("write error: %v", err)
	}

	got := make([]byte, len(payload))
	if _, err := io.ReadFull(reader, got); err != nil {
		t.Fatalf("read error: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("payload mismatch")
	}

	// The wire buffer should be fully consumed, including discarded padding.
	rest, err := io.ReadAll(&wire)
	if err != nil {
		t.Fatalf("read rest error: %v", err)
	}
	if len(rest) != 0 {
		t.Fatalf("unexpected remaining bytes: %d", len(rest))
	}
}
