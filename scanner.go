// Package reverse provides utilities for reading data in a reverse manner.
package reverse

import (
	"bytes"
	"errors"
	"io"
)

// ErrBufferTooSmall indicates that the scanned payload did not fit
// in the provided buffer.
var ErrBufferTooSmall = errors.New("buffer size too small")

// Scanner allows for scanning the contents of the underlying reader in reverse.
type Scanner struct {
	rd         io.ReaderAt // The underlying reader
	delimiter  byte        // Scanned parts separator
	chunkSize  int64       // Internal chunker size
	maxBufSize int64       // Max size of the returned buffer

	buf    []byte // Output buffer
	offset int64  // Last offset position
	eof    bool   // io.EOF
}

// NewScanner returns a new Scanner that reads from the tail up.
//
// The provided offset states at what position should the scanner start
// reading from.
func NewScanner(rd io.ReaderAt, offset int64) *Scanner {
	return &Scanner{
		rd:         rd,
		delimiter:  '\n',
		chunkSize:  1024,
		maxBufSize: 1 << 20,

		offset: offset,
	}
}

// SetMaxBufferSize sets max size for the buffer.
func (s *Scanner) SetMaxBufferSize(max int64) {
	s.maxBufSize = max
}

// SetChunkSize sets the size of the internal chunker.
func (s *Scanner) SetChunkSize(size int64) {
	s.chunkSize = size
}

// SetDelimiter set the given delimiter.
func (s *Scanner) SetDelimiter(delimiter byte) {
	s.delimiter = delimiter
}

// Scan the payload between delimiters.
//
// Returns an io.EOF after reading all the content.
func (s *Scanner) Scan() (out []byte, err error) {
	if s.eof {
		return nil, io.EOF
	}

	for {
		if idx := bytes.LastIndexByte(s.buf, s.delimiter); idx >= 0 {
			out = s.buf[idx+1:]
			s.buf = s.buf[:idx]
			return out, nil
		}

		if err := s.read(); err != nil {
			if errors.Is(err, io.EOF) {
				s.eof = true
				if len(s.buf) > 0 {
					return s.buf, nil
				}
			}

			return nil, err
		}
	}
}

// proceedChunker proceeds the offset and returns the current size.
func (s *Scanner) proceedChunker() int64 {
	size := s.chunkSize
	if size > s.offset {
		size = s.offset
	}
	s.offset -= size
	return size
}

// read performs a ReadAt operation on the underlying reader.
func (s *Scanner) read() error {
	if s.offset == 0 {
		return io.EOF
	}

	size := s.proceedChunker()
	bufSize := size + int64(len(s.buf))
	if bufSize > s.maxBufSize {
		return ErrBufferTooSmall
	}

	buf := make([]byte, size, bufSize)
	_, err := s.rd.ReadAt(buf, s.offset)
	if err == nil {
		s.buf = append(buf, s.buf...)
	}
	return err
}
