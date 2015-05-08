package gitgo

import (
	"testing"
	"time"
)

func Test_parseAuthorString(t *testing.T) {
	const input = "aditya <dev@chimeracoder.net> 1428349755 -0400"
	const expectedAuthor = "aditya"
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}

	expectedDate := time.Unix(1428349755, 0)
	expectedDate = expectedDate.In(loc)
	author, date, err := parseAuthorString(input)
	if err != nil {
		t.Fatal(err)
	}
	if author != expectedAuthor {
		t.Errorf("expected author %s and received %s", expectedAuthor, author)
	}
	if !expectedDate.Equal(date) {
		t.Errorf("expected date %s and received %s", expectedDate, date)
	}
}
