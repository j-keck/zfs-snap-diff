package main

import (
	"fmt"
	"strings"
	"testing"
)

var tiny1 = strings.Join([]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n"}, "\n") + "\n"
var tiny2 = strings.Join([]string{"a", "b", "C", "d", "e", "f", "g", "h", "I", "j", "k", "l", "m", "n"}, "\n") + "\n"

var small1 = `first line
second line
fourth line`
var small2 = `first line
second line
third line
fourth line`

func TestDiffInsert(t *testing.T) {
	res := Diff(small1, small2, 1)
	assertStringEq(t, res.LineBasedDeltas.String(), "{=:2,2:12,12:second line\n},{+:3,3:24,24:third line\n},{=:3,4:24,35:fourth line}")
}

func TestDiffDeletion(t *testing.T) {
	// switch small1 with small2 to get a delete
	res := Diff(small2, small1, 1)
	assertStringEq(t, res.LineBasedDeltas.String(), "{=:2,2:12,12:second line\n},{-:3,3:24,24:third line\n},{=:4,3:35,24:fourth line}")
}

func TestDiffContext(t *testing.T) {

	// without context
	res := Diff(tiny1, tiny2, 0)
	assertStringEq(t, res.LineBasedDeltas.String(), "{-:3,3:5,5:c\n},{+:3,3:5,5:C\n},{-:9,9:17,17:i\n},{+:9,9:17,17:I\n}")

	// with context
	res = Diff(tiny1, tiny2, 1)
	assertStringEq(t, res.LineBasedDeltas.String(),
		"{=:2,2:3,3:b\n},{-:3,3:5,5:c\n},{+:3,3:5,5:C\n},{=:4,4:7,7:d\n},"+
			"{=:8,8:15,15:h\n},{-:9,9:17,17:i\n},{+:9,9:17,17:I\n},{=:10,10:19,19:j\n}")

	// with overlapped context
	res = Diff(tiny1, tiny2, 3)
	assertStringEq(t, res.LineBasedDeltas.String(),
		"{=:1,1:1,1:a\nb\n},{-:3,3:5,5:c\n},{+:3,3:5,5:C\n},{=:4,4:7,7:d\ne\nf\ng\nh\n},{-:9,9:17,17:i\n},{+:9,9:17,17:I\n},{=:10,10:19,19:j\nk\nl\n}")

}

func assertIntEq(t *testing.T, i1, i2 int) {
	if i1 != i2 {
		msg := fmt.Sprintf("ints not equal! i1=%d, i2=%d", i1, i2)
		fmt.Println(msg)
		t.Error(msg)
	}
}
func assertStringEq(t *testing.T, s1, s2 string) {
	logString := func(name, s string) {
		cleanupString := func(s string) string {
			return strings.Replace(s, "\n", "\\n", -1)
		}
		fmt.Printf("%s: len=%d, content:'%s'\n", name, len(s), cleanupString(s))
	}

	if s1 != s2 {
		logString("s1", s1)
		logString("s2", s2)
		t.Error("strings not equal!")
	}
}
