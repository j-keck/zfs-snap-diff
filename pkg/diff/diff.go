package diff

import (
	"bytes"
	"fmt"

	"github.com/j-keck/go-diff/diffmatchpatch"
)

// DeltaType classifies a patch action like 'delete', 'equal' or 'insert'
type DeltaType int

const (
	// Del represents a deleted patch chunk
	Del DeltaType = (iota - 1)
	// Eq represents a not changed patch chunk
	Eq
	// Ins represents a added patch chunk
	Ins
)

// Delta represents a patch chunk
type Delta struct {
	Type           DeltaType
	LineNrFrom     int
	LineNrTarget   int
	StartPosFrom   int64
	StartPosTarget int64
	Text           string
}

func (d *Delta) String() string {
	var t string
	switch d.Type {
	case Ins:
		t = "+"
	case Del:
		t = "-"
	case Eq:
		t = "="
	default:
		panic("Unexpected DeltaType")
	}
	return fmt.Sprintf("{%s:%d,%d:%d,%d:%s}", t, d.LineNrFrom, d.LineNrTarget, d.StartPosFrom, d.StartPosTarget, d.Text)
}

// Deltas represents all patch chunks of a file
type Deltas []Delta

func (deltas Deltas) String() string {
	var buffer bytes.Buffer
	for idx, delta := range deltas {
		if idx > 0 {
			buffer.WriteString(",")
		}
		buffer.WriteString(delta.String())
	}
	return buffer.String()
}

// Diff compares two strings and returns the DiffResult
func Diff(from, target string, contextSize int) DiffResult {

	// init diff-match-patch and create the (diff-match-patch) diff
	dmp := diffmatchpatch.New()
	//	dmp.DiffEditCost = 10

	// make a line based diff - for side by side
	fromLines, targetLines, lines := dmp.DiffLinesToChars(from, target)
	lineBasedDiffs := dmp.DiffMain(fromLines, targetLines, false)

	if len(lineBasedDiffs) == 1 && lineBasedDiffs[0].Type == 0 {
		// cancel if no differences found
		return DiffResult{Deltas{}, Deltas{}, []string{}}
	}
	lineBasedDiffs = dmp.DiffCleanupSemantic(lineBasedDiffs)
	lineBasedDiffs = dmp.DiffCharsToLines(lineBasedDiffs, lines)
	lineBasedDeltas := createDeltasFromDiffs(lineBasedDiffs, contextSize)

	// make a char based diff - for intext
	charBasedDiffs := dmp.DiffMain(from, target, true)
	charBasedDiffs = dmp.DiffCleanupSemantic(charBasedDiffs)
	charBasedDeltas := createDeltasFromDiffs(charBasedDiffs, contextSize)

	// create patches from char based -> smaller patches
	var gnuDiffs []string
	for _, patch := range dmp.PatchMake(charBasedDiffs) {
		gnuDiffs = append(gnuDiffs, patch.String())
	}

	return DiffResult{lineBasedDeltas, charBasedDeltas, gnuDiffs}
}

func createDeltasFromDiffs(diffs []diffmatchpatch.Diff, contextSize int) Deltas {
	var deltas Deltas
	var lineNrFrom, lineNrTarget int = 1, 1       // first line is line nr. 1
	var startPosFrom, startPosTarget int64 = 1, 1 // first char ist at pos 1

	idx := 0
	nextDiffIfTypeIs := func(diffType diffmatchpatch.Operation) (diffmatchpatch.Diff, bool) {
		if idx < len(diffs) && diffs[idx].Type == diffType {
			next := diffs[idx]
			idx++
			return next, true
		}
		return diffmatchpatch.Diff{}, false
	}

	createPrevEqDelta := func(diff diffmatchpatch.Diff) Delta {
		lineCount := countNewLines(diff.Text)
		lines := splitText(diff.Text)

		var text string
		if lineCount > contextSize {
			fromIndex := lineCount - contextSize
			text = joinLines(lines[fromIndex:])
		} else {
			text = diff.Text
		}
		count := countNewLines(text)
		length := int64(len(text))

		return Delta{
			Eq,
			lineNrFrom - count,
			lineNrTarget - count,
			startPosFrom - length,
			startPosTarget - length,
			text,
		}
	}

	createAfterEqDelta := func(diff diffmatchpatch.Diff) (Delta, bool) {
		lineCount := countNewLines(diff.Text)
		lines := splitText(diff.Text)

		var text string
		nextPrevEqMerged := false
		if idx < len(diffs) && lineCount < contextSize*2 {
			// merge - but not the last element
			text = diff.Text
			nextPrevEqMerged = true
		} else if lineCount < contextSize {
			text = diff.Text
		} else {
			text = joinLines(lines[:contextSize])
		}

		return Delta{
			Eq,
			lineNrFrom,
			lineNrTarget,
			startPosFrom,
			startPosTarget,
			text,
		}, nextPrevEqMerged
	}

	// first equal block if there is one
	if diff, ok := nextDiffIfTypeIs(diffmatchpatch.DiffEqual); ok {
		lineCount := countNewLines(diff.Text)
		textLength := int64(len(diff.Text))

		lineNrFrom += lineCount
		lineNrTarget += lineCount
		startPosFrom += textLength
		startPosTarget += textLength

		if contextSize > 0 {
			deltas = append(deltas, createPrevEqDelta(diff))
		}
	}
	for idx < len(diffs) {
		var delDiff, insDiff diffmatchpatch.Diff
		var hasDel, hasIns bool

		// add del-delta if there is a delete
		delDiff, hasDel = nextDiffIfTypeIs(diffmatchpatch.DiffDelete)
		if hasDel {
			deltas = append(deltas, Delta{
				Del,
				lineNrFrom,
				lineNrTarget,
				startPosFrom,
				startPosTarget,
				delDiff.Text,
			})
		}

		// add ins-delta if there is a insert
		insDiff, hasIns = nextDiffIfTypeIs(diffmatchpatch.DiffInsert)
		if hasIns {
			deltas = append(deltas, Delta{
				Ins,
				lineNrFrom,
				lineNrTarget,
				startPosFrom,
				startPosTarget,
				insDiff.Text,
			})

		}

		// update lineNr / startPos
		//   * after the delta additions!
		//   * if not after, a text replace has the wrong lineNr / startPos
		//     in the ins-delta
		if hasDel {
			lineCount := countNewLines(delDiff.Text)
			textLength := int64(len(delDiff.Text))

			lineNrFrom += lineCount
			startPosFrom += textLength
		}
		if hasIns {
			lineCount := countNewLines(insDiff.Text)
			textLength := int64(len(insDiff.Text))

			lineNrTarget += lineCount
			startPosTarget += textLength
		}

		// handle equal
		//   * add after context
		//   * update line / pos
		//   * add prev context for the next
		if diff, ok := nextDiffIfTypeIs(diffmatchpatch.DiffEqual); ok {
			lineCount := countNewLines(diff.Text)
			textLength := int64(len(diff.Text))

			var afterEqDelta Delta
			var nextPrevEqMerged bool
			if contextSize > 0 {
				afterEqDelta, nextPrevEqMerged = createAfterEqDelta(diff)
				deltas = append(deltas, afterEqDelta)
			}

			lineNrFrom += lineCount
			lineNrTarget += lineCount
			startPosFrom += textLength
			startPosTarget += textLength

			// don't add prevEq if
			//  * no context requested
			//  * we are at the end
			//  * the afterEqDelta has merged the next (this) prevEqDelta, because the context overlap
			//  * the afterEqDelta contains the whole text
			if contextSize > 0 && idx < len(diffs) && !nextPrevEqMerged && afterEqDelta.Text != diff.Text {
				deltas = append(deltas, createPrevEqDelta(diff))
			}
		}
	}

	return deltas
}
