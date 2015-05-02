package gitgo

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

// GitObject represents a commit, tree, or blob.
// Under the hood, these may be objects stored directly
// or through packfiles
type GitObject interface {
	Type() string
	//Contents() string
}

type gitObject struct {
	Type string

	// Commit fields
	Tree      string
	Parents   []string
	Author    string
	Committer string
	Message   []byte
	Size      string

	// Tree
	Blobs []objectMeta
	Trees []objectMeta

	// Blob
	Contents []byte
}

// A Blob compresses content from a file
type Blob struct {
	_type    string
	size     string
	Contents []byte
}

func (b Blob) Type() string {
	return b._type
}

type Commit struct {
	_type     string
	Name      SHA
	Tree      string
	Parents   []SHA
	Author    string
	Committer string
	Message   []byte
	size      string
}

func (c Commit) Type() string {
	return c._type
}

type Tree struct {
	_type string
	Blobs []objectMeta
	Trees []objectMeta
	size  string
}

func (t Tree) Type() string {
	return t._type
}

// objectMeta contains the metadata
// (hash, permissions, and filename)
// corresponding either to a blob (leaf) or another tree
type objectMeta struct {
	Hash     SHA
	Perms    string
	filename string
}

func NewObject(input SHA, basedir string) (obj GitObject, err error) {
	if basedir == "" {
		basedir = ".git"
	}

	filename := path.Join(basedir, "objects", string(input[:2]), string(input[2:]))

	f, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			obj, err := searchPacks(input, basedir)
			if err != nil {
				return nil, err
			}
			return obj.normalize(basedir)
		}
		return
	}
	defer f.Close()
	r, err := zlib.NewReader(f)
	if err != nil {
		return nil, err
	}
	return parseObj(r, input, basedir)
}

func normalizePerms(perms string) string {
	// TODO don't store permissions as a string
	for len(perms) < 6 {
		perms = "0" + perms
	}
	return perms
}

func parseObj(r io.Reader, name SHA, basedir string) (result GitObject, err error) {
	// TODO fixme
	bts, err := ioutil.ReadAll(r)
	if err != nil {
		return result, err
	}
	obj := bts

	parts := bytes.Split(obj, []byte("\x00"))
	parts = bytes.Fields(parts[0])
	resultType := string(parts[0])
	resultSize := string(parts[1])
	nullIndex := bytes.Index(obj, []byte("\x00"))

	switch resultType {
	case "commit":
		return parseCommit(bytes.NewReader(obj[nullIndex+1:]), resultSize, name)
	case "tree":
		return parseTree(bytes.NewReader(obj), resultSize, basedir)

	case "blob":
		return parseBlob(bytes.NewReader(obj[nullIndex+1:]), resultSize)
	default:
		err = fmt.Errorf("Received unknown object type %s", resultType)
	}

	return
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

func parseCommit(r io.Reader, resultSize string, name SHA) (Commit, error) {
	var commit = Commit{_type: "commit", size: resultSize}
	bts, err := ioutil.ReadAll(r)
	if err != nil {
		return commit, err
	}
	lines := bytes.Split(bts, []byte("\n"))
	for i, line := range lines {
		// The next line is the commit message
		if len(bytes.Fields(line)) == 0 {
			commit.Message = bytes.Join(lines[i+1:], []byte("\n"))
			break
		}
		parts := bytes.Fields(line)
		key := parts[0]
		switch keyType(key) {
		case treeKey:
			commit.Tree = string(parts[1])
		case parentKey:
			commit.Parents = append(commit.Parents, SHA(string(parts[1])))
		case authorKey:
			commit.Author = string(bytes.Join(parts[1:], []byte(" ")))
		case committerKey:
			commit.Committer = string(bytes.Join(parts[1:], []byte(" ")))
		default:
			err = fmt.Errorf("encountered unknown field in commit: %s", key)
			return commit, err
		}
	}
	commit.Name = name
	return commit, nil
}

func parseTree(r io.Reader, resultSize string, basedir string) (Tree, error) {
	var tree = Tree{_type: "tree", size: resultSize}

	scanner := bufio.NewScanner(r)
	scanner.Split(ScanNullLines)

	var tmp objectMeta

	var resultObjs []objectMeta

	for count := 0; ; count++ {
		done := !scanner.Scan()
		if done {
			break
		}

		txt := scanner.Text()

		if count == 0 {
			// the first time through, scanner.Text() will be
			// "tree <size>"
			continue
		}
		if count == 1 {
			// the second time through, scanner.Text() will be
			// <perms> <filename>
			// separated by a space
			fields := strings.Fields(txt)
			tmp.Perms = normalizePerms(fields[0])
			tmp.filename = fields[1]
			continue
		}

		// after the second time through, scanner.Text() will be
		// <sha><perms2> <file2>
		// where perms2 and file2 refer to the permissions and filename (respectively)
		// of the NEXT object, and <sha> is the first 20 bytes exactly.
		// If there is no next object (this is the last object)
		// then scanner.Text() will yield exactly 20 bytes.

		// decode the next 20 bytes to get the SHA
		tmp.Hash = SHA(hex.EncodeToString([]byte(txt[:20])))
		resultObjs = append(resultObjs, tmp)
		if len(txt) <= 20 {
			// We've read the last line
			break
		}

		// Now, tmp points to the next object in the tree listing
		tmp = objectMeta{}
		remainder := txt[20:]
		fields := strings.Fields(remainder)
		tmp.Perms = normalizePerms(fields[0])
		tmp.filename = fields[1]
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return tree, err
	}

	for _, part := range resultObjs {
		obj, err := NewObject(part.Hash, basedir)
		if err != nil {
			return tree, err
		}
		switch obj.Type() {
		case "tree":
			tree.Trees = append(tree.Trees, part)
		case "blob":
			tree.Blobs = append(tree.Blobs, part)
		default:
			return tree, fmt.Errorf("Unknown type found: %s", obj.Type())
		}
	}
	return tree, nil
}

func parseBlob(r io.Reader, resultSize string) (Blob, error) {
	var blob = Blob{_type: "blob", size: resultSize}
	bts, err := ioutil.ReadAll(r)
	blob.Contents = bts
	return blob, err
}

func findFromPrefix(prefix SHA, basedir string) (GitObject, error) {
	objectsDir := path.Join(basedir, "objects", string(prefix[:2]))
	files, err := ioutil.ReadDir(objectsDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	file, err := findUniquePrefix(prefix, files)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		// Try the packfile
		obj, err := searchPacks(prefix, basedir)
		if err != nil {
			return nil, err
		}
		return obj.normalize(basedir)
	}
	f, err := os.Open(file.Name())
	defer f.Close()
	r, err := zlib.NewReader(f)
	if err != nil {
		return nil, err
	}
	return parseObj(r, prefix, basedir)

}

func findUniquePrefix(prefix SHA, files []os.FileInfo) (os.FileInfo, error) {
	var result os.FileInfo
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasPrefix(file.Name(), string(prefix)) {
			if result != nil {
				return nil, fmt.Errorf("prefix is not unique: %s", prefix)
			}
			result = file
		}
	}
	if result == nil {
		return nil, os.ErrNotExist
	}
	return result, nil
}
