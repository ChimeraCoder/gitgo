package gitgo

import (
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"reflect"
)

type errReadSeeker struct {
	r   io.ReadSeeker
	err error
}

// Read, but only if no errors have been encountered
// in a previous read (including io.EOF)
func (er *errReadSeeker) read(buf []byte) int {
	var n int
	if er.err != nil {
		return 0
	}
	n, er.err = io.ReadFull(er.r, buf)
	return n
}

func (er *errReadSeeker) Seek(offset int64, whence int) (int64, error) {
	return er.r.Seek(offset, whence)
}

// VerifyPack returns the pack objects contained in the packfile and
// corresponding index file.
func VerifyPack(pack io.ReadSeeker, idx io.Reader) ([]*packObject, error) {

	objectsMap := map[SHA]*packObject{}
	objects, err := parsePack(errReadSeeker{pack, nil}, idx)
	for _, object := range objects {
		objectsMap[object.Name] = object
	}

	for _, object := range objectsMap {
		if object.err != nil {
			continue
		}
		if object._type == OBJ_OFS_DELTA {
			// TODO improve this
			// linear search to find the right offset
			var base *packObject
			for _, o := range objects {
				if o.Offset == object.Offset-object.negativeOffset {
					base = o
					break
				}
			}
			if base == nil {
				object.err = fmt.Errorf("could not find object with negative offset %d - %d for %s", object.Offset, object.negativeOffset, object.Name)
				continue
			}
			object.BaseObjectName = base.Name
		}
	}

	for _, object := range objectsMap {

		object.err = object.Patch(objectsMap)
	}
	return objects, err
}

func parsePack(pack errReadSeeker, idx io.Reader) (objects []*packObject, err error) {
	signature := make([]byte, 4)
	pack.read(signature)
	if string(signature) != "PACK" {
		return nil, fmt.Errorf("Received invalid signature: %s", string(signature))
	}
	if err != nil {
		return nil, err
	}
	version := make([]byte, 4)
	pack.read(version)

	// TODO use encoding/binary here
	v := version[3]
	switch v {
	case 2:
		// Parse version 2 packfile
		objects, err = parseIdx(idx, 2)
		if err != nil {
			return
		}
		objects, err = parsePackV2(pack, objects)
		return

	default:
		return nil, fmt.Errorf("cannot parse packfile with version %d", v)
	}
}

func bytesToNum(b []byte) uint {
	var n uint
	for i := 0; i < len(b); i++ {
		n = n | (uint(b[len(b)-i-1]) << uint(i*8))
	}
	return n
}

// parsePackV2 parses a packfile that uses
// version 2 of the format
func parsePackV2(r errReadSeeker, objects []*packObject) ([]*packObject, error) {

	numObjectsBts := make([]byte, 4)
	r.read(numObjectsBts)
	if int(bytesToNum(numObjectsBts)) != len(objects) {
		return nil, fmt.Errorf("Expected %d objects and found %d", len(objects), numObjectsBts)
	}

	for _, object := range objects {

		var btsread int
		r.Seek(int64(object.Offset), os.SEEK_SET)
		_bytes := make([]byte, 1)
		btsread += r.read(_bytes)
		_byte := _bytes[0]

		// This will extract the last three bits of
		// the first nibble in the byte
		// which tells us the object type
		object._type = packObjectType(((_byte >> 4) & 7))

		// determine the (decompressed) object size
		// and then deflate the following bytes

		// The most-significant byte (MSB)
		// tells us whether we need to read more bytes
		// to get the encoded object size
		MSB := (_byte & 128) // will be either 128 or 0

		// This will extract the last four bits of the byte
		var objectSize = int((uint(_byte) & 15))

		// shift the first size by 0
		// and the rest by 4 + (i-1) * 7
		var shift uint = 4

		// If the most-significant bit is 0, this is the last byte
		// for the object size
		for MSB > 0 {
			// Keep reading the size until the MSB is 0
			_bytes := make([]byte, 1)
			btsread += r.read(_bytes)
			_byte := _bytes[0]

			MSB = (_byte & 128)

			objectSize += int((uint(_byte) & 127) << shift)
			shift += 7
		}

		object.Size = objectSize
		switch {
		case object._type < 5:
			// the object is a commit, tree, blob, or tag

			// (objectSize) is the size, in bytes, of this object *when expanded*
			// the IDX file tells us how many *compressed* bytes the object will take
			// (in other words, how much space to allocate for the result)
			object.Data = make([]byte, objectSize)

			zr, err := zlib.NewReader(r.r)
			if err != nil {
				return nil, err
			}

			n, err := zr.Read(object.Data)
			if err != nil {
				if err == io.EOF {
					err = nil
				} else {
					return nil, err
				}
			}
			zr.Close()
			object.Data = object.Data[:n]

			// TODO figure out why sometimes n < objectSize

		case object._type == OBJ_OFS_DELTA:
			// read the n-byte offset
			// from the git docs:
			// "n bytes with MSB set in all but the last one.
			// The offset is then the number constructed by
			// concatenating the lower 7 bit of each byte, and
			// for n >= 2 adding 2^7 + 2^14 + ... + 2^(7*(n-1))
			// to the result."

			var offset int

			// number of bytes read in variable length encoding
			var nbytes uint

			MSB := 128
			for (MSB & 128) > 0 {
				nbytes++

				// Keep reading the size until the MSB is 0
				_bytes := make([]byte, 1)
				r.read(_bytes)
				_byte := _bytes[0]

				sevenBytes := uint(_byte) & 127

				offset = (offset << 7) + int(sevenBytes)

				MSB = int(_byte & 128)
				if MSB == 0 {
					break
				}
			}

			if nbytes >= 2 {
				offset += (1 << (7 * (nbytes - 1)))
			}

			object.negativeOffset = offset
			object.baseOffset = object.Offset - object.negativeOffset
			object.Data = make([]byte, objectSize)

			zr, err := zlib.NewReader(r.r)
			if err != nil {
				object.err = err
				continue
			}
			n, err := zr.Read(object.Data)
			if err != nil && err != io.EOF {
				object.err = err
				continue
			}
			object.Data = object.Data[:n]
			zr.Close()
			if len(object.Data) != objectSize {
				object.err = fmt.Errorf("received wrong object size: %d (expected %d)", object.Data, objectSize)
			}

		case object._type == OBJ_REF_DELTA:
			r.Seek(int64(object.Offset), os.SEEK_SET)
			// Read the 20-byte base object name
			baseObjName := make([]byte, 20)

			r.read(baseObjName)
			object.Data = make([]byte, objectSize)

			zr, err := zlib.NewReader(r.r)
			if err != nil {
				object.err = err
				continue
			}
			n, err := zr.Read(object.Data)
			if err != nil && err != io.EOF {
				object.err = err
				continue
			}
			object.Data = object.Data[:n]
			zr.Close()
		}
	}

	return objects, r.err
}

func parseIdx(idx io.Reader, version int) (objects []*packObject, err error) {
	if version != 2 {
		return nil, fmt.Errorf("cannot parse IDX with version %d", version)
	}
	// parse version 2 idxfile

	// Version 2 starts with a 4-byte magic number
	header := make([]byte, 4)
	n, err := idx.Read(header)
	if err != nil {
		return nil, err
	}

	if !reflect.DeepEqual([]byte{255, 116, 79, 99}, header) {
		return nil, fmt.Errorf("invalid IDX header: %q", string(header))
	}

	// Then the version number in four bytes
	versionBts := make([]byte, 4)
	_, err = idx.Read(versionBts)
	if err != nil {
		return nil, err
	}
	// We already know the version, so we can ignore it

	// Then the fanout table
	// The fanout table has 256 entries, each 4 bytes long
	fanoutTableFlat := make([]byte, 256*4)
	n, err = idx.Read(fanoutTableFlat)
	if err == nil && n != len(fanoutTableFlat) {
		err = fmt.Errorf("read incomplete fanout table: %d", n)
	}
	if err != nil {
		return nil, err
	}

	// Initialize the flat fanout table
	fanoutTable := make([][]byte, 256)
	for i := 0; i < len(fanoutTableFlat); i += 4 {
		entry := fanoutTableFlat[i : i+4]
		fanoutTable[(i+1)/4] = entry
	}

	numObjects := int(bytesToNum(fanoutTable[len(fanoutTable)-1]))
	objects = make([]*packObject, numObjects)

	objectNames := make([]SHA, numObjects)

	for i := 0; i < numObjects; i++ {
		sha := make([]byte, 20)
		n, err = idx.Read(sha)
		if err != nil {
			return nil, err
		}

		objectNames[i] = SHA(fmt.Sprintf("%x", sha[:n]))
		objects[i] = &packObject{Name: SHA(fmt.Sprintf("%x", sha[:n]))}
	}

	// Then come 4-byte CRC32 values
	crc32Table := make([]byte, numObjects*4)
	_, err = idx.Read(crc32Table)
	if err != nil {
		return nil, err
	}

	// Next come 4-byte offset values
	// If the MSB is set, there is an index into the next table
	// otherwise, these are 31 bits each
	offsetsFlat := make([]byte, numObjects*4)
	_, err = idx.Read(offsetsFlat)
	if err != nil {
		return nil, err
	}

	offsets := make([]int, numObjects)
	for i := 0; i < len(offsets); i++ {
		offset := int(bytesToNum(offsetsFlat[i*4 : (i+1)*4]))
		// check if the MSB is 1
		if offset&2147483648 > 0 {
			return nil, fmt.Errorf("packfile is too large to parse")
		}
		offsets[i] = offset
		objects[i].Offset = offset
	}

	// If the pack file is more than 2 GB, there will be a table of 8-byte offset entries here
	// TODO implement this

	// This is the same as the checksum at the end of the corresponding packfile
	packfileChecksum := make([]byte, 20)
	_, err = idx.Read(packfileChecksum)
	if err != nil {
		return
	}

	// This is the checksum of all of the above data
	// We're not checking it now, but if we can't read it properly
	// that means an error has occurred earlier in parsing
	idxChecksum := make([]byte, 20)
	_, err = idx.Read(idxChecksum)
	if err != nil {
		return
	}

	// TODO check that there isn't any data left

	return objects, err
}
