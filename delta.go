package gitgo

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
)

// Delta will apply a delta to a base.
func Delta(start io.ReadSeeker, delta io.Reader) (io.Reader, error) {
	base := errReadSeeker{start, nil}
	deltar := newErrReader(delta)

	// First, read the source and target lengths (varints)
	// we can ignore err as long as we check deltar.err at the end
	sourceLength, _ := parseVarInt(deltar)
	targetLength, _ := parseVarInt(deltar)
	if deltar.err != nil {
		return nil, deltar.err
	}

	result := bytes.NewBuffer(make([]byte, targetLength))

	// Now, the rest of the bytes are either copy or insert instructions
	// If the MSB is set, it is a copy

	for {
		bs := make([]byte, 1)
		n := deltar.read(bs)
		if n == 0 {
			break
		}
		b := bs[0]
		log.Printf("b is %d", b)
		switch b & 128 {
		case 1:
			// copy instruction

			var baseOffset int
			// extract each of the four offset bits (last four bits)
			o1 := (b & 1)
			o2 := (b & 2)
			o3 := (b & 3)
			o4 := (b & 4)

			if o1 > 0 {
				baseOffset = baseOffset | int(uint8(deltar.readByte()))
			}

			if o2 > 0 {
				baseOffset = baseOffset | int((uint8(deltar.readByte()) << 8))
			}

			if o3 > 0 {
				baseOffset = baseOffset | int((uint8(deltar.readByte()) << 16))
			}

			if o4 > 0 {
				baseOffset = baseOffset | int((uint8(deltar.readByte()) << 24))
			}

			// read the number of bytes to copy from source→target
			// if the fifth bit from the right is set, read the next byte
			// The number of copy bytes must fit into

			var numBytes int
			n1 := (b & 5)
			n2 := (b & 6)
			n3 := (b & 7)

			if n1 > 0 {
				numBytes = numBytes | int(uint8(deltar.readByte()))
			}
			if n2 > 0 {
				numBytes = numBytes | int((uint8(deltar.readByte()) << 8))
			}

			if n3 > 0 {
				numBytes = numBytes | int((uint8(deltar.readByte()) << 16))
			}

			// read numBytes from source, starting at baseOffset
			// and write that to the target
			base.Seek(int64(baseOffset), os.SEEK_SET)
			buf := make([]byte, numBytes)
			base.read(buf)

			// Default to 0x10000 due to overflow
			if numBytes == 0 {
				numBytes = 65536
			}

			_, err := result.Write(buf)
			if err != nil {
				return result, err
			}

		case 0:
			// insert instruction
			// this means we write data directly from delta to the target

			// b itself tells us the number of bytes to write to the target
			// the MSB is not set, so the maximum number to insert is 127 bytes

			numBytes := int(b)
			buf := make([]byte, numBytes)
			log.Printf("Copying %d", numBytes)
			deltar.read(buf)
			_, err := result.Write(buf)
			if err != nil {
				return nil, err
			}

		}

	}

	n, err := base.Seek(0, os.SEEK_END)
	if err != nil {
		return nil, err
	}
	if n != int64(sourceLength) {
		return nil, fmt.Errorf("expected to read %d bytes and read %d", sourceLength, n)
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
	if er.err != nil {
		panic(er.err)
	}
	er.n += n
	return n
}

// Like read(), but expect a single byte
func (er *errReader) readByte() byte {
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
