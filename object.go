package gitgo

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
)

type GitObject interface {
	Type() string
}

type gitObject struct {
	Type string

	// Commit fields
	Tree      string
	Parents   []string
	Author    string
	Committer string
	Message   string
	Size      string

	// Tree
	Blobs []objectMeta
	Trees []objectMeta

	// Blob
	Contents string
}

type Blob struct {
	_type    string
	size     string
	Contents string
}

func (b Blob) Type() string {
	return b._type

}

type Commit struct {
	_type     string
	Tree      string
	Parents   []string
	Author    string
	Committer string
	Message   string
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

func NewObject(input SHA) (obj GitObject, err error) {
	str, err := CatFile(input)
	if err != nil {
		return
	}
	return parseObj(str)
}

func normalizePerms(perms string) string {
	// TODO don't store permissions as a string
	for len(perms) < 6 {
		perms = "0" + perms
	}
	return perms
}

func parseObj(obj string) (result GitObject, err error) {

	parts := strings.Split(obj, "\x00")
	parts = strings.Fields(parts[0])
	resultType := parts[0]
	resultSize := parts[1]
	nullIndex := strings.Index(obj, "\x00")

	lines := strings.Split(obj[nullIndex+1:], "\n")

	switch resultType {
	case "commit":
		var commit = Commit{_type: resultType, size: resultSize}
		for i, line := range lines {
			// The next line is the commit message
			if len(strings.Fields(line)) == 0 {
				commit.Message = strings.Join(lines[i+1:], "\n")
				break
			}
			parts := strings.Fields(line)
			key := parts[0]
			switch KeyType(key) {
			case TreeKey:
				commit.Tree = parts[1]
			case ParentKey:
				commit.Parents = append(commit.Parents, parts[1])
			case AuthorKey:
				commit.Author = strings.Join(parts[1:], " ")
			case CommitterKey:
				commit.Committer = strings.Join(parts[1:], " ")
			default:
				err = fmt.Errorf("Encounterd unknown field in commit: %s", key)
				return
			}
		}
		result = commit
	case "tree":
		var tree = Tree{_type: resultType, size: resultSize}

		scanner := bufio.NewScanner(bytes.NewBuffer([]byte(obj)))
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
			obj, err := NewObject(part.Hash)
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
		result = tree
	case "blob":
		var blob = Blob{_type: resultType, size: resultSize}
		blob.Contents = obj[nullIndex+1:]
		result = blob
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
