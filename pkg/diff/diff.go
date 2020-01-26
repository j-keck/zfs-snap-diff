package diff

import (
	"github.com/j-keck/go-diff/diffmatchpatch"
	"github.com/j-keck/zfs-snap-diff/pkg/fs"
)

type Diff struct {
	Deltas                     []Deltas `json:"deltas"`
	Patches                    []string `json:"patches"`
	SideBySideDiffHTMLFragment []string `json:"sideBySideDiffHTMLFragment"`
	InlineDiffHTMLFragment     []string `json:"inlineDiffHTMLFragment"`
}

func NewDiff(from, target string, contextSize int) Diff {

	// init diff-match-patch and create the (diff-match-patch) diff
	dmp := diffmatchpatch.New()
	//	dmp.DiffEditCost = 10

	// make a line based diff - for side by side
	fromLines, targetLines, lines := dmp.DiffLinesToChars(from, target)
	lineBasedDiffs := dmp.DiffMain(fromLines, targetLines, false)

	if len(lineBasedDiffs) == 1 && lineBasedDiffs[0].Type == 0 {
		// cancel if no differences found
		return Diff{[]Deltas{}, []string{}, []string{}, []string{}}
	}

	lineBasedDiffs = dmp.DiffCleanupSemantic(lineBasedDiffs)
	lineBasedDiffs = dmp.DiffCharsToLines(lineBasedDiffs, lines)
	lineBasedDeltas := createDeltasFromDiffs(lineBasedDiffs, contextSize)

	// make a char based diff - for inline diff
	charBasedDiffs := dmp.DiffMain(from, target, true)
	charBasedDiffs = dmp.DiffCleanupSemantic(charBasedDiffs)
	charBasedDeltas := createDeltasFromDiffs(charBasedDiffs, contextSize)

	// deltas
	deltas := splitDeltasByContext(lineBasedDeltas)

	// create patches from char based -> smaller patches
	var patches []string
	for _, patch := range dmp.PatchMake(charBasedDiffs) {
		patches = append(patches, patch.String())
	}

	return Diff{
		deltas,
		patches,
		createSideBySideDiffHTMLFragment(lineBasedDeltas),
		createInlineDiffHTMLFragment(charBasedDeltas),
	}
}

func NewDiffFromPath(from, target string, contextSize int) (Diff, error) {
	fromContent, err := readTextFile(from)
	if err != nil {
		return Diff{}, err
	}

	targetContent, err := readTextFile(target)
	if err != nil {
		return Diff{}, err
	}

	return NewDiff(fromContent, targetContent, contextSize), nil
}

func readTextFile(path string) (string, error) {
	fh, err := fs.NewFileHandle(path)
	if err != nil {
		return "", err
	}

	return fh.ReadString()
}
