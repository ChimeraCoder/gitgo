package gitgo

import (
	"fmt"
	"io"
)

// scanner functions similarly to bufio.Scanner,
// except that it never reads more input than necessary,
// which allow predictable consumption (and reuse) of readers
type scanner struct {
	r    io.Reader
	data []byte
	err  error
}

func (s *scanner) scan() bool {
	if s.err != nil {
		return false
	}
	s.data = s.read()
	return s.err == nil
}

func (s *scanner) Err() error {
	if s.err == io.EOF {
		return nil
	}
	return s.err
}

func (s *scanner) read() []byte {
	if s.err != nil {
		return nil
	}
	result := make([]byte, 1)
	n, err := s.r.Read(result)
	if err != nil {
		s.err = err
		return nil
	}
	if n == 0 {
		s.err = fmt.Errorf("read zero bytes")
	}
	return result
}
