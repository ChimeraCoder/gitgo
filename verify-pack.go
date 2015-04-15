package gitgo

import (
	"bufio"
	"compress/zlib"
	"fmt"
	"io"
	"log"
	"path"
	"path/filepath"
	"reflect"
)

func GetIdxPath(dotGitRootPath string) (idxFilePath string, err error) {
	files, err := filepath.Glob(path.Join(dotGitRootPath, "objects/pack", "*.idx"))
	idxFilePath = files[0]
	return
}

func VerifyPack(pack io.Reader, idx io.Reader) error {
	versionChan := make(chan int)
	err := parsePack(pack, versionChan)
	if err != nil {
		return err
	}

	go parseIdx(idx, versionChan)
	return err
}

func parsePack(pack io.Reader, versionChan chan<- int) (err error) {
	signature := make([]byte, 4)
	n, err := pack.Read(signature)
	if err == nil && n != 4 {
		return fmt.Errorf("expected to read 4 bytes, read %d", n)
	}
	if string(signature) != "PACK" {
		return fmt.Errorf("Received invalid signature: %s", string(signature))
	}
	if err != nil {
		return err
	}
	log.Printf("signature %+v", signature)

	version := make([]byte, 4)
	_, err = pack.Read(version)
	if err != nil {
		return err
	}
	// TODO use encoding/binary here
	log.Printf("version is %+v", version)
	v := version[3]
	switch v {
	case 2:
		// Parse version 2 packfile
		go func() {
			versionChan <- 2
		}()
		return parsePackV2(pack)

	default:
		return fmt.Errorf("cannot parse packfile with version %d", v)
	}

	return nil
}

const (
	OBJ_COMMIT    uint8 = 1
	OBJ_TREE            = 2
	OBJ_BLOB            = 3
	OBJ_TAG             = 4
	OBJ_OFS_DELTA       = 6
	OBJ_REF_DELTA       = 7
)

func parsePackV2(pack io.Reader) error {
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

		// The most-significant byte (MSB)
		// tells us whether we need to read more bytes
		// to get the encoded object size
		MSB := (_byte & 128) // will be either 128 or 0

		// This will extract the last three bits of
		// the first nibble in the byte
		// which tells us the object type
		_ = ((_byte >> 4) & 7)

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
		if n != objectSize {
			return fmt.Errorf("expected to read %d bytes, read %d", objectSize, n)
		}
		log.Printf("read %+v", object)
	}

	return nil
}

func parseIdx(idx io.Reader, versionChan <-chan int) (err error) {
	version := <-versionChan
	if version != 2 {
		return fmt.Errorf("cannot parse IDX with version %d")
	}
	// parse version 2 idxfile

	// Version 2 starts with a 4-byte magic number
	header := make([]byte, 4)
	n, err := idx.Read(header)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual([]byte{255, 116, 79, 99}, header) {
		return fmt.Errorf("invalid IDX header: %q", string(header))
	}

	// Then the version number in four bytes
	versionBts := make([]byte, 4)
	_, err = idx.Read(versionBts)
	if err != nil {
		return err
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
		return err
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

	return nil
}
