package gitgo

type keyType string

// SHA represents the SHA-1 hash used by git
type SHA string

const (
	treeKey      keyType = "tree"
	parentKey            = "parent"
	authorKey            = "author"
	committerKey         = "committer"
)
