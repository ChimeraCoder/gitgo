package gitgo

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// patchDelta will apply a delta to a base.
func patchDelta(start io.ReadSeeker, delta io.Reader) (io.Reader, error) {
	base := errReadSeeker{start, nil}
	deltar := newErrReader(delta)

	// First, read the source and target lengths (varints)
	// we can ignore err as long as we check deltar.err at the end
	// we have no need for the targetLength right now, so we discard it
	sourceLength, _ := parseVarInt(deltar)
	_, _ = parseVarInt(deltar)
	if deltar.err != nil {
		return nil, deltar.err
	}

	result := bytes.NewBuffer(nil)
	//result = io.MultiWriter(result, os.Stderr)

	// Now, the rest of the bytes are either copy or insert instructions
	// If the MSB is set, it is a copy

	for {
		bs := make([]byte, 1)
		n := deltar.read(bs)
		if n == 0 {
			break
		}

		// b represents our command itself (opcode)
		b := bs[0]
		switch b & 128 {
		case 128:
			// b is a copy instruction

			// the last four bits represent the offset from the base (source)
			var baseOffset int
			if (b & 1) > 0 {
				baseOffset = baseOffset | int(uint(deltar.readByte()))
			}

			if (b & 2) > 0 {
				baseOffset = baseOffset | int((uint(deltar.readByte()) << 8))
			}

			if (b & 4) > 0 {
				baseOffset = baseOffset | int((uint(deltar.readByte()) << 16))
			}

			if (b & 8) > 0 {
				baseOffset = baseOffset | int((uint(deltar.readByte()) << 24))
			}

			// read the number of bytes to copy from source→target
			// if the fifth bit from the right is set, read the next byte
			// The number of copy bytes must fit into

			var numBytes uint

			if (b & 16) > 0 {
				numBytes = numBytes | uint(uint(deltar.readByte(true)))
			}
			if (b & 32) > 0 {
				numBytes = numBytes | uint((uint(deltar.readByte(true)) << 8))
			}

			if (b & 64) > 0 {
				numBytes = numBytes | uint((uint(deltar.readByte(true)) << 16))
			}

			// Default to 0x10000 due to overflow
			if numBytes == 0 {
				numBytes = 65536
			}

			// read numBytes from source, starting at baseOffset
			// and write that to the target
			base.Seek(int64(baseOffset), os.SEEK_SET)
			buf := make([]byte, numBytes)
			base.read(buf)

			_, err := result.Write(buf)
			if err != nil {
				return result, err
			}

		case 0:
			if b == 0 {
				// cmd == 0 is reserved for future encoding extensions
				return nil, fmt.Errorf("cannot process delta opcode 0")
			}

			// insert instruction
			// this means we write data directly from delta to the target

			// b itself tells us the number of bytes to write to the target
			// the MSB is not set, so the maximum number to insert is 127 bytes

			numBytes := int(b)
			buf := make([]byte, numBytes)
			deltar.read(buf)
			_, err := result.Write(buf)
			if err != nil {
				return nil, err
			}

		default:
			return nil, fmt.Errorf("invalid opcode %08b", b)
		}
	}

	n, err := base.Seek(0, os.SEEK_END)
	if err != nil {
		return nil, err
	}
	if n != int64(sourceLength) {
		return nil, fmt.Errorf("expected to read %d bytes and read %d", sourceLength, n)
	}

	if deltar.err == io.EOF {
		return result, nil
	}
	return result, deltar.err
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
	n   int
	err error
}

// Turn an io.Reader into an errReader.
// If r is already an io.Reader, this is a no-op
func newErrReader(r io.Reader) errReader {
	er, ok := r.(errReader)
	if !ok {
		er = errReader{r, 0, nil}
	}
	return er
}

// Read, but only if no errors have been encountered
// in a previous read (including io.EOF)
func (er *errReader) read(buf []byte) int {
	var n int
	if er.err != nil {
		return 0
	}
	n, er.err = io.ReadFull(er.r, buf)
	er.n += n
	return n
}

// Like read(), but expect a single byte
func (er *errReader) readByte(p ...bool) byte {
	b := make([]byte, 1)
	n := er.read(b)
	if n != 1 && er.err != nil {
		if er.err == io.EOF {
			er.err = io.EOF
			return b[0]
		}
		er.err = fmt.Errorf("expected to read single byte and read none")
	}
	return b[0]
}

// Should not actually be called
// Defined only to ensure that errReader is itself an ioReader
func (er errReader) Read(buf []byte) (int, error) {
	n := er.read(buf)
	return n, er.err
}
