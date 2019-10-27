package diff

import (
	"bytes"
	"fmt"
	"html"
	"strings"
)

// DiffResult represents the result from a diff
type DiffResult struct {
	LineBasedDeltas Deltas
	CharBasedDeltas Deltas
	GNUDiffs        []string
}

func splitDeltasByContext(deltas Deltas) []Deltas {

	var splitted []Deltas

	var block Deltas
	for idx, delta := range deltas {
		// when the current and the previous are eq blocks
		if delta.Type == Eq && idx > 0 && deltas[idx-1].Type == Eq {
			// and there is a line distance
			if delta.LineNrFrom-deltas[idx-1].LineNrFrom > 1 {
				// we are in the middle of a context switch
				splitted = append(splitted, block)
				block = Deltas{}
			}
		}
		block = append(block, delta)
	}
	splitted = append(splitted, block)

	return splitted
}

// DeltasByContext splits the Delta at their context bounds
func (dr *DiffResult) DeltasByContext() []Deltas {
	return splitDeltasByContext(dr.LineBasedDeltas)
}

// AsSideBySideHTML creates a SideBySide Diff
func (dr *DiffResult) AsSideBySideHTML() []string {
	var htmlBlocks []string

	if len(dr.LineBasedDeltas) == 0 {
		// noting to do
		return htmlBlocks
	}

	// function to cleanup a text to embed it in html
	//   * escape special characters
	//   * convert tabs and spaces to &nbsp;
	cleanupText := func(text string) string {
		escaped := html.EscapeString(text)
		r := strings.NewReplacer("\t", "&nbsp;&nbsp;", " ", "&nbsp;")
		return r.Replace(escaped)
	}

	for _, deltas := range splitDeltasByContext(dr.LineBasedDeltas) {
		var buf bytes.Buffer

		deltaIdx := 0
		nextDeltaIfTypeIs := func(deltaType DeltaType) (Delta, bool) {
			if deltaIdx < len(deltas) && deltas[deltaIdx].Type == deltaType {
				next := deltas[deltaIdx]
				deltaIdx++
				return next, true
			}
			return Delta{}, false
		}

		addEqDelta := func(delta Delta) {
			format := "<tr><td class='line-nr'>%d</td><td>%s</td><td class='line-nr'>%d</td><td>%s</td></tr>"
			for i, line := range splitText(cleanupText(delta.Text)) {
				buf.WriteString(fmt.Sprintf(format, delta.LineNrFrom+i, line, delta.LineNrTarget+i, line))
			}
		}

		for deltaIdx < len(deltas) {
			var delDelta, insDelta Delta
			var delLines, insLines []string
			var deltaFound bool

			// prev context
			if delta, deltaFound := nextDeltaIfTypeIs(Eq); deltaFound {
				addEqDelta(delta)
			}

			if delDelta, deltaFound = nextDeltaIfTypeIs(Del); deltaFound {
				delLines = splitText(cleanupText(delDelta.Text))
			}

			if insDelta, deltaFound = nextDeltaIfTypeIs(Ins); deltaFound {
				insLines = splitText(cleanupText(insDelta.Text))
			}

			delLinesLen := len(delLines)
			insLinesLen := len(insLines)
			linesLenMax := max(delLinesLen, insLinesLen)
			for i := 0; i < linesLenMax; i++ {
				buf.WriteString("<tr>")
				if i < delLinesLen {
					buf.WriteString(fmt.Sprintf("<td class='line-nr'>%d</td><td class='del'>%s</td>", delDelta.LineNrFrom+i, delLines[i]))
				} else {
					buf.WriteString("<td class='line-nr'></td><td></td>")
				}
				if i < insLinesLen {
					buf.WriteString(fmt.Sprintf("<td class='line-nr'>%d</td><td class='ins'>%s</td>", insDelta.LineNrTarget+i, insLines[i]))
				} else {
					buf.WriteString("<td class='line-nr'></td><td></td>")
				}
				buf.WriteString("</tr>")
			}

			// after context
			if delta, ok := nextDeltaIfTypeIs(Eq); ok {
				addEqDelta(delta)
			}

		}
		// add context-block
		htmlBlocks = append(htmlBlocks, buf.String())
	}

	return htmlBlocks
}

// AsIntextHTML creates an Inline Diff
func (dr *DiffResult) AsIntextHTML() []string {
	var htmlBlocks []string

	if len(dr.CharBasedDeltas) == 0 {
		// nothing to do
		return htmlBlocks
	}

	// split deltas at context-bounds
	for _, deltas := range splitDeltasByContext(dr.CharBasedDeltas) {

		// build intext diff
		var contextBlock bytes.Buffer
		for _, delta := range deltas {
			var className string
			switch delta.Type {
			case Ins:
				className = "ins"
			case Del:
				className = "del"
			case Eq:
				className = "eq"
			}

			text := html.EscapeString(delta.Text)
			if text == "\n" {
				// FIXME: geht im chrome nicht - svg mit inkscape machen
				text = "&#9252;\n"
			}
			snippet := fmt.Sprintf("<span class='%s'>%s</span>", className, text)
			contextBlock.WriteString(snippet)
		}

		// split at new-lines and add line numbers
		lineNr := deltas[0].LineNrTarget
		var buffer bytes.Buffer
		for i, line := range splitText(contextBlock.String()) {
			buffer.WriteString(fmt.Sprintf("<span class='line-nr'>%d</span> %s", lineNr+i, line))
		}
		htmlBlocks = append(htmlBlocks, buffer.String())
	}

	return htmlBlocks
}
