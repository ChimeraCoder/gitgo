package gitgo

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"strings"
)

type GitObject struct {
	Type string

	// Commit fields
	Tree      string
	Parents   []string
	Author    string
	Committer string
	Message   string
	Size      string

	// Tree blob
	Blobs []treePart
	Trees []treePart
}

// a treePart contains the hash, permissions, and filename
// corresponding either to a blob (leaf) or another tree
type treePart struct {
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

func parseObj(obj string) (result GitObject, err error) {

	parts := strings.Split(obj, "\x00")
	parts = strings.Fields(parts[0])
	result.Type = parts[0]
	result.Size = parts[1]
	nullIndex := strings.Index(obj, "\x00")

	lines := strings.Split(obj[nullIndex+1:], "\n")

	log.Printf("obj %s", obj)
	switch result.Type {
	case "commit":
		for i, line := range lines {
			// The next line is the commit message
			if len(strings.Fields(line)) == 0 {
				result.Message = strings.Join(lines[i+1:], "\n")
				break
			}
			parts := strings.Fields(line)
			key := parts[0]
			switch KeyType(key) {
			case TreeKey:
				result.Tree = parts[1]
			case ParentKey:
				result.Parents = append(result.Parents, parts[1])
			case AuthorKey:
				result.Author = strings.Join(parts[1:], " ")
			case CommitterKey:
				result.Committer = strings.Join(parts[1:], " ")
			default:
				err = fmt.Errorf("Encounterd unknown field in commit: %s", key)
				return
			}
		}
	case "tree":
		log.Printf("obj is %q", obj)

		scanner := bufio.NewScanner(bytes.NewBuffer([]byte(obj)))
		scanner.Split(ScanNullLines)

		var result treePart

		var resultObjs []treePart

		for count := 0; ; count++ {
			done := !scanner.Scan()
			if done {
				break
			}

			txt := scanner.Text()
			log.Printf("Text is %q", txt)

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
				result.Perms = fields[0]
				result.filename = fields[1]
				continue
			}

			// after the second time through, scanner.Text() will be
			// <sha><perms2> <file2>
			// where perms2 and file2 refer to the permissions and filename (respectively)
			// of the NEXT object, and <sha> is the first 20 bytes exactly.
			// If there is no next object (this is the last object)
			// then scanner.Text() will yield exactly 20 bytes.

			// decode the next 20 bytes to get the SHA
			result.Hash = SHA(hex.EncodeToString([]byte(txt[:20])))
			resultObjs = append(resultObjs, result)
			if len(txt) <= 20 {
				// We've read the last line
				break
			}

			// Now, result points to the next object in the tree listing
			result = treePart{}
			remainder := txt[20:]
			fields := strings.Fields(remainder)
			result.Perms = fields[0]
			result.filename = fields[1]
		}

		if err := scanner.Err(); err != nil && err != io.EOF {
			return GitObject{}, err
		}

		log.Printf("Result objects are %+v", resultObjs)
		/*
			items := obj

			// Items come in the form
			// A FNS
			// http://stackoverflow.com/questions/17910016/unzip-git-tree-object

			log.Printf("Items is %q", items)
			permsEnd := strings.Index(items, " ")
			perms := items[:permsEnd]
			log.Printf("perms are %q", perms)
			filenameEnd := strings.Index(items, "\x00")
			filename := items[permsEnd+1 : filenameEnd]

			log.Printf("filename is %s", filename)

			// the sha will be exactly 20 bytes
			shaBts := []byte(items[filenameEnd+1 : filenameEnd+1+20])
			sha := SHA(hex.EncodeToString(shaBts))
			log.Printf("Sha is %q", sha)
		*/

		/*
			for _, line := range lines {
				fields := strings.Fields(line)
				perms := fields[0]
				_type := fields[1]
				hash := SHA(fields[2])
				filename := fields[3]

				log.Printf(" line is %s", line)
				switch _type {
				case "blob":
					result.Blobs = append(result.Blobs, struct {
						Hash     SHA
						Perms    string
						filename string
					}{hash, perms, filename})
				case "tree":
					result.Trees = append(result.Trees, struct {
						Hash     SHA
						Perms    string
						filename string
					}{hash, perms, filename})
				default:
					err = fmt.Errorf("Encountered unknown type in tree: %s", _type)
					return
				}
			}
		*/
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
