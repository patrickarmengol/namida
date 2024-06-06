package main

import (
	_ "embed"
	"errors"
	"fmt"
	"math/rand/v2"
	"strings"
)

//go:embed haikus.txt
var haikusFile string

type haiku struct {
	anchorCell cell

	letters   map[cell]rune
	letterVis map[cell]struct{} // hashset

	state string // new, linger, erase, cooldown, done
	steps uint
}

func (h *haiku) isVisible() bool {
	return len(h.letterVis) != 0
}

func (h *haiku) isFullyVisible() bool {
	return len(h.letterVis) == len(h.letters)
}

func (h *haiku) updateState(opts options) {
	switch h.state {
	case "new":
		if h.isFullyVisible() {
			h.state = "linger"
		}
		// case default:
		//     h.state = "done"
	case "linger":
		if h.steps > opts.lingerFrames {
			h.state = "erase"
		} else {
			h.steps++
		}
	case "erase":
		if !h.isVisible() {
			h.state = "cooldown"
			h.steps = 0
		}
	case "cooldown":
		if h.steps > opts.cooldownFrames {
			h.state = "done"
		} else {
			h.steps++
		}
	case "done":
	default:
		h.state = "done"
	}
}

func newHaiku(haikuString string, anchorCell cell) haiku {
	haikuLines := strings.Split(haikuString, "\n")
	if len(haikuLines) != 4 {
		// panic since this should have been checked in parsing
		panic(fmt.Errorf("haiku doesn't have 4 lines: %v %d", haikuLines, len(haikuLines)))
	}

	letters := map[cell]rune{}
	for i, l := range haikuLines {
		colOffset := -(4 * i)
		if i == 3 {
			colOffset = colOffset - 4
		}
		rowOffset := i
		for j, c := range []rune(l) {
			letters[cell{anchorCell.col + colOffset, anchorCell.row + rowOffset + j}] = c
		}
	}

	return haiku{
		anchorCell,
		letters,
		map[cell]struct{}{},
		"new",
		0,
	}
}

func parseHaikuStrings(hf string) ([]string, error) {
	if len(hf) == 0 {
		return nil, errors.New("haiku file empty")
	}

	fullHaikus := strings.Split(strings.TrimSpace(hf), "\n\n")
	if len(fullHaikus) == 1 {
		return nil, errors.New("could only find 1 haiku; delimited by 2 newlines")
	}

	for _, h := range fullHaikus {
		haikuLines := strings.Split(h, "\n")
		if len(haikuLines) != 4 {
			return nil, fmt.Errorf("haiku doesn't have 4 lines: %v %d", haikuLines, len(haikuLines))
		}
	}
	return fullHaikus, nil
}

func randomAnchorPos(w int, h int) cell {
	margin := 4
	minCol := margin + 16
	maxCol := w - margin
	minRow := margin
	maxRow := h - margin - 20

	col := rand.IntN(maxCol-minCol) + minCol
	col = col - (col % 2)
	row := rand.IntN(maxRow-minRow) + minRow

	return cell{col, row}
}
