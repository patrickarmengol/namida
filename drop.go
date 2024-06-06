package main

type drop struct {
	pos  cell
	done bool
}

func (d *drop) step() {
	if d.pos.row >= height {
		d.done = true
	}
	d.pos.row++
}

func intersect(h haiku, d drop) rune {
	r, ok := h.letters[d.pos]
	// if haiku doesn't include cell, return space
	if !ok {
		return JPSPC
	}
	switch h.state {
	case "new":
		h.letterVis[d.pos] = struct{}{}
		return r
	case "linger":
		return r
	case "erase":
		delete(h.letterVis, d.pos)
		return JPSPC
	case "cooldown":
		return JPSPC
	default:
		return JPSPC
	}
}
