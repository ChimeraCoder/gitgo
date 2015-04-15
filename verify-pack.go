package gitgo

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"log"
	"path"
	"path/filepath"
	"reflect"
)

type packObject struct {
	Name   SHA
	Offset int
}

const (
	OBJ_COMMIT    uint8 = 1
	OBJ_TREE            = 2
	OBJ_BLOB            = 3
	OBJ_TAG             = 4
	OBJ_OFS_DELTA       = 6
	OBJ_REF_DELTA       = 7
)

func GetIdxPath(dotGitRootPath string) (idxFilePath string, err error) {
	files, err := filepath.Glob(path.Join(dotGitRootPath, "objects/pack", "*.idx"))
	idxFilePath = files[0]
	return
}

func VerifyPack(pack io.Reader, idx io.Reader) error {
	versionChan := make(chan int)

	_, err := parsePack(pack, idx, versionChan)
	return err
}

func parsePack(pack io.Reader, idx io.Reader, versionChan chan<- int) (objects []packObject, err error) {
	signature := make([]byte, 4)
	n, err := pack.Read(signature)
	if err == nil && n != 4 {
		return nil, fmt.Errorf("expected to read 4 bytes, read %d", n)
	}
	if string(signature) != "PACK" {
		return nil, fmt.Errorf("Received invalid signature: %s", string(signature))
	}
	if err != nil {
		return nil, err
	}
	log.Printf("signature %+v", signature)

	version := make([]byte, 4)
	_, err = pack.Read(version)
	if err != nil {
		return nil, err
	}
	// TODO use encoding/binary here
	log.Printf("version is %+v", version)
	v := version[3]
	switch v {
	case 2:
		// Parse version 2 packfile
		objects, err = parseIdx(idx, 2)
		if err != nil {
			return
		}
		return nil, parsePackV2(pack, objects)

	default:
		return nil, fmt.Errorf("cannot parse packfile with version %d", v)
	}

	return nil, nil
}

func Clone(r io.Reader) (*bufio.Reader, *bufio.Reader) {
	var b1 bytes.Buffer
	var b2 bytes.Buffer
	w := io.MultiWriter(&b1, &b2)
	io.Copy(w, r)
	return bufio.NewReader(&b1), bufio.NewReader(&b2)
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
func parsePackV2(pack io.Reader, objects []packObject) error {

	r := bufio.NewReader(pack)
	numObjectsBts := make([]byte, 4)
	_, err := r.Read(numObjectsBts)
	if err != nil {
		return err
	}

	var numObjects uint32
	for i := 0; i < len(numObjectsBts); i++ {
		numObjects = numObjects | (uint32(numObjectsBts[len(numObjectsBts)-i-1]) << uint(i*8))
	}

	for i := 0; i < int(numObjects); i++ {
		_byte, err := r.ReadByte()
		if err != nil {
			return err
		}

		// This will extract the last three bits of
		// the first nibble in the byte
		// which tells us the object type
		objectType := ((_byte >> 4) & 7)
		if objectType < 5 {
			// the object is a commit, tree, blob, or tag
		}
		switch {
		case objectType < 5:
			// the object is a commit, tree, blob, or tag
			log.Printf("Object type %d", objectType)

			// determine the (decompressed) object size
			// and then deflate the following bytes

			// The most-significant byte (MSB)
			// tells us whether we need to read more bytes
			// to get the encoded object size
			MSB := (_byte & 128) // will be either 128 or 0

			// This will extract the last four bits of the byte
			var objectSize int = int((uint(_byte) & 15))

			// shift the first size by 0
			// and the rest by 4 + (i-1) * 7
			var shift uint = 4

			// If the most-significant bit is 0, this is the last byte
			// for the object size
			for MSB > 0 {
				// Keep reading the size until the MSB is 0
				_byte, err = r.ReadByte()
				if err != nil {
					return err
				}
				MSB = (_byte & 128)

				objectSize += int((uint(_byte) & 127) << shift)
				shift += 7
			}

			// (objectSize) is the size, in bytes, of this object *when expanded*
			// the IDX file tells us how many *compressed* bytes the object will take
			// (in other words, how much space to allocate for the result)
			object := make([]byte, objectSize)

			zr, err := zlib.NewReader(r)
			if err != nil {
				return err
			}
			n, err := zr.Read(object)
			if err != nil {
				return err
			}
			zr.Close()

			if n != objectSize {
				return fmt.Errorf("expected to read %d bytes, read %d", objectSize, n)
			}
			log.Printf("read %+v", string(object))

		case objectType == OBJ_OFS_DELTA:
			// read the n-byte offset
			log.Printf("encountered ofs delta")

		case objectType == OBJ_REF_DELTA:
			// Read the 20-byte base object name
			log.Printf("encountered ref delta")
		}
	}

	return nil
}

func parseIdx(idx io.Reader, version int) (objects []packObject, err error) {
	if version != 2 {
		return nil, fmt.Errorf("cannot parse IDX with version %d")
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

	for i, row := range fanoutTable {
		log.Printf("row %d: %+v", i, row)
	}

	numObjects := int(bytesToNum(fanoutTable[len(fanoutTable)-1]))
	objects = make([]packObject, numObjects)

	objectNames := make([]SHA, numObjects)

	for i := 0; i < numObjects; i++ {
		sha := make([]byte, 20)
		n, err = idx.Read(sha)
		if err != nil {
			return nil, err
		}
		log.Printf("%x", sha[:n])

		objectNames[i] = SHA(fmt.Sprintf("%x", sha[:n]))
		objects[i].Name = SHA(fmt.Sprintf("%x", sha[:n]))
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
