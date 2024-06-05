package main

import (
	_ "embed"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

//go:embed haikus.txt
var haikusFile string

const JPSPC = '　'

var (
	width      int
	height     int
	tickLength = time.Second / 30
	defStyle   = tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	textStyle  = tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorWhite)
	dropStyle  = tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorWhite)
)

func watchEvents(s tcell.Screen, quit chan<- struct{}, pause chan<- struct{}, resume chan<- struct{}) {
	for {
		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape, tcell.KeyCtrlC:
				close(quit)
				return
			case tcell.KeyCtrlL:
				s.Sync()
			}
		case *tcell.EventResize:
			s.Sync()
			width, height = ev.Size()
			if width < 40 || height < 20 {
				pause <- struct{}{}
			} else {
				resume <- struct{}{}
			}
		}
	}
}

type drop struct {
	col  int
	row  int
	done bool
}

func (d *drop) step() {
	if d.row >= height {
		d.done = true
	}
	d.row++
}

type line struct {
	col    int
	start  int // row index
	length int
	phrase string
	vis    bool
}

type haiku struct {
	lines [4]*line
	state string // new, visisible, erasing
	steps int
}

func (h haiku) hasCol(col int) bool {
	for _, l := range h.lines {
		if l.col == col {
			return true
		}
	}
	return false
}

func (h haiku) isVis() bool {
	for _, l := range h.lines {
		if !l.vis {
			return false
		}
	}
	return true
}

func (h haiku) isClear() bool {
	for _, l := range h.lines {
		if l.vis {
			return false
		}
	}
	return true
}

func randomAnchorPos(w int, h int) (int, int) {
	margin := 4
	col := rand.Intn(w-(40-2*margin)) + margin + (40 - 2*margin)
	col = col - (col % 2)
	row := rand.Intn(h-(20-2*margin)) + margin
	return col, row
}

func newHaiku(haikuString string) haiku {
	haikuLines := strings.Split(haikuString, "\n")
	if len(haikuLines) != 4 {
		// panic since this should have been checked in parsing
		panic(fmt.Errorf("haiku doesn't have 4 lines: %v %d", haikuLines, len(haikuLines)))
	}

	anchorCol, anchorRow := randomAnchorPos(width, height)

	return haiku{
		lines: [4]*line{
			{anchorCol, anchorRow, utf8.RuneCountInString(haikuLines[0]), haikuLines[0], false},
			{anchorCol - 4, anchorRow + 1, utf8.RuneCountInString(haikuLines[1]), haikuLines[1], false},
			{anchorCol - 8, anchorRow + 2, utf8.RuneCountInString(haikuLines[2]), haikuLines[2], false},
			{anchorCol - 16, anchorRow + 3, utf8.RuneCountInString(haikuLines[3]), haikuLines[3], false},
		},
		state: "new",
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

func intersect(h haiku, col int, row int) (rune, bool) {
	for _, l := range h.lines {
		if col == l.col {
			diff := row - l.start
			if diff >= 0 && diff < l.length {
				var r rune
				if h.state == "erasing" || h.state == "erased" {
					r = JPSPC
					if diff == 0 {
						l.vis = false
					}
				} else {
					r = []rune(l.phrase)[diff]
					if diff == 0 {
						l.vis = true
					}
				}
				return r, true
			}
			return JPSPC, false
		}
	}
	return JPSPC, false
}

type letter struct {
	col  int
	row  int
	char rune
}

func runSim(s tcell.Screen, quit <-chan struct{}, pause <-chan struct{}, resume <-chan struct{}, hss []string) {
	t := time.NewTicker(tickLength)
	running := false

	ds := []drop{}

	hai := newHaiku(hss[rand.Intn(len(hss))])

	for {
		select {
		case <-quit:
			return
		case <-pause:
			if running {
				running = false
			}
		case <-resume:
			ds = []drop{}
			hai = newHaiku(hss[rand.Intn(len(hss))])
			s.Clear()
			running = true
		case <-t.C:
			if running {

				// logic

				// update haiku state machine
				if hai.state == "new" && hai.isVis() {
					hai.state = "visible"
				} else if hai.state == "visible" {
					if hai.steps > 400 {
						hai.state = "erasing"
					} else {
						hai.steps++
					}
				} else if hai.state == "erasing" {
					if hai.isClear() {
						hai.state = "erased"
					}
				} else if hai.state == "erased" {
					if hai.steps > 800 {
						hai = newHaiku(hss[rand.Intn(len(hss))])
					} else {
						hai.steps++
					}
				}

				// check existing drop positions intersect haiku letters
				toFlip := []letter{}
				for _, d := range ds {
					r, ok := intersect(hai, d.col, d.row)
					if ok {
						toFlip = append(toFlip, letter{d.col, d.row, r})
					} else {
						toFlip = append(toFlip, letter{d.col, d.row, JPSPC})
					}
				}

				// create new drop
				c := rand.Intn(width)
				c = c - (c % 2) // skip odd columns
				ds = append(ds, drop{c, -1, false})
				// step all drops and filter out finished
				newds := []drop{}
				for _, d := range ds {
					d.step()
					if !d.done {
						newds = append(newds, d)
					}
				}
				ds = newds

				// render

				// render drops
				for _, d := range ds {
					// set drop tile
					s.SetContent(d.col, d.row, JPSPC, nil, dropStyle)
				}
				// render changed characters
				for _, l := range toFlip {
					s.SetContent(l.col, l.row, l.char, nil, textStyle)
				}

				// update display
				s.Show()
			}
		}
	}
}

func main() {
	// initialize tcell screen
	screen, err := tcell.NewScreen()
	if err != nil {
		panic(err)
	}
	err = screen.Init()
	if err != nil {
		panic(err)
	}
	// clean up on quit
	cleanUp := func() {
		maybePanic := recover()
		screen.Fini()
		if maybePanic != nil {
			panic(maybePanic)
		}
		// fmt.Println("asdf")
		os.Exit(0)
	}
	defer cleanUp()

	// set styling
	screen.HideCursor()
	screen.SetStyle(defStyle)
	screen.Clear()

	// get initial bounds
	width, height = screen.Size() // non-inclusive

	// load haikus
	hss, err := parseHaikuStrings(haikusFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	// start sim
	quit := make(chan struct{})
	pause := make(chan struct{}, 1)
	resume := make(chan struct{})

	go watchEvents(screen, quit, pause, resume)
	runSim(screen, quit, pause, resume, hss)
}
