package gitgo

import (
	"bytes"
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

// ScanNullLines is like bufio.ScanLines, except it uses the null character as the delimiter
// instead of a newline
func ScanNullLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\x00'); i >= 0 {
		// We have a full null-terminated line.
		return i + 1, data[0:i], nil
	}
	// If we're at EOF, we have a final, non-terminated "line". Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

// ScanLinesNoTrim is exactly like bufio.ScanLines, except it does not trim the newline
func ScanLinesNoTrim(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0 : i+1], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}
