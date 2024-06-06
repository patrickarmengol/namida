package main

import (
	"fmt"
	"math/rand/v2"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
)

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
			if width < 50 || height < 30 {
				pause <- struct{}{}
			} else {
				resume <- struct{}{}
			}
		}
	}
}

type cell struct {
	col int
	row int
}

func runSim(s tcell.Screen, quit <-chan struct{}, pause <-chan struct{}, resume <-chan struct{}, hss []string) {
	t := time.NewTicker(tickLength)
	running := false

	ds := []drop{}

	hai := newHaiku(hss[rand.IntN(len(hss))], randomAnchorPos(width, height))

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
			hai = newHaiku(hss[rand.IntN(len(hss))], randomAnchorPos(width, height))
			s.Clear()
			running = true
		case <-t.C:
			if running {
				// logic

				// update haiku state machine
				hai.updateState()
				if hai.state == "done" {
					hai = newHaiku(hss[rand.IntN(len(hss))], randomAnchorPos(width, height))
				}

				// check existing drop positions intersect haiku letters
				toFlip := map[cell]rune{}
				for _, d := range ds {
					toFlip[d.pos] = intersect(hai, d)
				}

				// create new drop
				c := rand.IntN(width)
				c = c - (c % 2) // skip odd columns
				ds = append(ds, drop{cell{c, -1}, false})
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

				// render changed characters
				for ce, ch := range toFlip {
					s.SetContent(ce.col, ce.row, ch, nil, textStyle)
				}

				// render drops; should overwrite / be drawn over characters
				for _, d := range ds {
					// set drop tile
					s.SetContent(d.pos.col, d.pos.row, JPSPC, nil, dropStyle)
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
