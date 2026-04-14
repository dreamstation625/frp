package net

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"sync"
)

// ExtraPaddingBytesFunc returns how many extra random bytes should be appended
// to the current payload frame.
type ExtraPaddingBytesFunc func(payloadLen int) int

// NewAsymmetricStream wraps rwc with a framed stream.
//
// Wire format for each frame:
//   - 4 bytes payload length (big endian)
//   - 4 bytes extra padding length (big endian)
//   - payload bytes
//   - padding bytes (random, discarded by reader)
//
// The reader always returns payload bytes only, so padding bytes will not be
// returned to upper-layer user connections.
func NewAsymmetricStream(rwc io.ReadWriteCloser, extraPaddingFn ExtraPaddingBytesFunc) io.ReadWriteCloser {
	if extraPaddingFn == nil {
		extraPaddingFn = func(int) int { return 0 }
	}
	return &asymmetricStream{
		rwc:            rwc,
		extraPaddingFn: extraPaddingFn,
	}
}

type asymmetricStream struct {
	rwc            io.ReadWriteCloser
	extraPaddingFn ExtraPaddingBytesFunc

	writeMu sync.Mutex
	readMu  sync.Mutex

	frameRemain int
	padRemain   int
}

func (s *asymmetricStream) Read(p []byte) (int, error) {
	s.readMu.Lock()
	defer s.readMu.Unlock()

	if len(p) == 0 {
		return 0, nil
	}

	if s.frameRemain == 0 {
		header := make([]byte, 8)
		if _, err := io.ReadFull(s.rwc, header); err != nil {
			return 0, err
		}
		s.frameRemain = int(binary.BigEndian.Uint32(header[:4]))
		s.padRemain = int(binary.BigEndian.Uint32(header[4:]))
	}

	if s.frameRemain == 0 {
		return s.discardPadding()
	}

	n := len(p)
	if n > s.frameRemain {
		n = s.frameRemain
	}

	readN, err := io.ReadFull(s.rwc, p[:n])
	s.frameRemain -= readN
	if err != nil {
		return readN, err
	}
	if s.frameRemain == 0 {
		if _, err := s.discardPadding(); err != nil {
			return readN, err
		}
	}
	return readN, nil
}

func (s *asymmetricStream) discardPadding() (int, error) {
	discarded := 0
	for s.padRemain > 0 {
		toRead := s.padRemain
		if toRead > 4096 {
			toRead = 4096
		}
		buf := make([]byte, toRead)
		n, err := io.ReadFull(s.rwc, buf)
		discarded += n
		s.padRemain -= n
		if err != nil {
			return discarded, err
		}
	}
	return discarded, nil
}

func (s *asymmetricStream) Write(p []byte) (int, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	paddingLen := s.extraPaddingFn(len(p))
	if paddingLen < 0 {
		paddingLen = 0
	}

	if len(p) > int(^uint32(0)) || paddingLen > int(^uint32(0)) {
		return 0, fmt.Errorf("payload too large")
	}

	header := make([]byte, 8)
	binary.BigEndian.PutUint32(header[:4], uint32(len(p)))
	binary.BigEndian.PutUint32(header[4:], uint32(paddingLen))

	if _, err := s.rwc.Write(header); err != nil {
		return 0, err
	}
	if len(p) > 0 {
		if _, err := s.rwc.Write(p); err != nil {
			return 0, err
		}
	}

	if paddingLen > 0 {
		padding := make([]byte, paddingLen)
		if _, err := cryptorand.Read(padding); err != nil {
			return 0, err
		}
		if _, err := s.rwc.Write(padding); err != nil {
			return 0, err
		}
	}
	return len(p), nil
}

func (s *asymmetricStream) Close() error {
	return s.rwc.Close()
}
