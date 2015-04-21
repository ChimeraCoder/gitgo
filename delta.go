package gitgo

import "io"

func Delta(start, delta io.Reader) io.Reader {

	// First, read the source and target lengths (varints)

	return nil
}

func parseVarInt(r io.Reader) (int, error) {
	// The MSB of the first byte indicates whether to read
	// the next byte

	_bytes := make([]byte, 1)
	_, err := r.Read(_bytes)
	if err != nil {
		return 0, err
	}
	_byte := _bytes[0]

	// The most-significant byte (MSB)
	// tells us whether we need to read more bytes
	// to get the encoded object size
	MSB := (_byte & 128) // will be either 128 or 0

	// This will extract the last seven bits of the byte
	var objectSize int = int((uint(_byte) & 127))

	// shift the first size by 0
	// and the rest by (i-1) * 7
	var shift uint = 0

	// If the most-significant bit is 0, this is the last byte
	// for the object size
	for MSB > 0 {
		shift += 7
		// Keep reading the size until the MSB is 0
		_bytes := make([]byte, 1)
		_, err := r.Read(_bytes)
		if err != nil {
			return 0, err
		}
		_byte := _bytes[0]

		MSB = (_byte & 128)

		objectSize += int((uint(_byte) & 127) << shift)
	}
	return objectSize, nil
}

type errReader struct {
	r   io.Reader
	err error
}

// Read, but only if no errors have been encountered
// in a previous read (including io.EOF)
func (er *errReader) read(buf []byte) int {
	var n int
	if er.err != nil {
		return 0
	}
	n, er.err = io.ReadFull(er.r, buf)
	return n
}
