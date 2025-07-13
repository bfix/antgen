//----------------------------------------------------------------------
// This file is part of antgen.
// Copyright (C) 2024-present Bernd Fix >Y<,  DO3YQ
//
// antgen is free software: you can redistribute it and/or modify it
// under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License,
// or (at your option) any later version.
//
// antgen is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//
// SPDX-License-Identifier: AGPL3.0-or-later
//----------------------------------------------------------------------

package lib

import (
	_ "embed"
	"fmt"
	"image/color"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tfriedel6/canvas"
	"github.com/tfriedel6/canvas/sdlcanvas"
)

//----------------------------------------------------------------------
// SDL canvas
//----------------------------------------------------------------------

//go:embed ankacoder.ttf
var font []byte

// Task send via channel to render engine
type Task struct {
	Ant *Antenna // antenna to be rendered
	Pos int      // position of last change (-1 no change)
	Msg string   // additional message for display
}

// SDLCanvas for windowed display
type SDLCanvas struct {
	w, h              float64 // model size
	cw, ch            int     // current canvas size
	scale, offX, offY float64 // active scale and margin
	txtSize           float64 // text size (large)

	win *sdlcanvas.Window
	cv  *canvas.Canvas

	taskCh  chan Task   // channel to render loop
	curr    Task        // current render task
	lock    sync.Mutex  // lock for updating parameters
	count   int         // number of tasks processed
	waiting atomic.Bool // pause rendering?
	stepper atomic.Bool // single-step?
	hint    string      // hint for display
}

// NewSDLCanvas creates a new SDL canvas for display
func NewSDLCanvas(width, height int, side float64) (c *SDLCanvas, err error) {
	c = new(SDLCanvas)
	c.taskCh = make(chan Task)
	c.count = -1
	// create window
	if c.win, c.cv, err = sdlcanvas.CreateWindow(width, height, "Antenna optimization"); err != nil {
		return
	}
	// load font
	_, _ = c.cv.LoadFont(font)
	c.cw, c.ch = width, height
	c.rescale(1.2 * side)
	c.offX, c.offY = float64(width)/2, float64(height)/2
	return
}

// rescale for larger/small geometry extends
func (c *SDLCanvas) rescale(side float64) {
	c.w, c.h = 2*side, 2*side
	c.scale = min(float64(c.cw)/c.w, float64(c.ch)/c.h)
	c.txtSize = 36 / c.scale
}

// Close a canvas. No further operations are allowed
func (c *SDLCanvas) Close() error {
	close(c.taskCh)
	return nil
}

// Show antenna geometry with message and last change position
func (c *SDLCanvas) Show(ant *Antenna, pos int, msg string) {
	c.taskCh <- Task{ant, pos, msg}
}

func (c *SDLCanvas) SetHint(m string) {
	c.hint = m
}

// Run the canvas (new rendering begins)
func (c *SDLCanvas) Run(cb Action) {

	// get render task from channel
	go func() {
		for task := range c.taskCh {
			if task.Ant == nil {
				return
			}
			// idle on wait
			for c.waiting.Load() {
				time.Sleep(100 * time.Millisecond)
			}
			// update geometry, message and change pos
			c.lock.Lock()
			c.curr = task

			// pause in single step mode and on track mark
			if task.Pos == TRK_MARK || c.stepper.Load() {
				c.waiting.Store(true)
			}

			c.count++
			c.lock.Unlock()
		}
	}()

	// pause/resume on key press ("Enter" key)
	c.waiting.Store(false)
	c.win.KeyDown = func(scancode int, rn rune, name string) {
		// handle custom callback
		if cb != nil {
			if cb(c.curr.Ant, rn, c.count) {
				c.waiting.Store(!c.waiting.Load())
				c.stepper.Store(false)
				return
			}
		}
		// handle key presses
		switch name {
		case "Enter":
			c.waiting.Store(!c.waiting.Load())
			c.stepper.Store(false)
		case "Space":
			if c.waiting.Load() {
				c.stepper.Store(true)
				c.waiting.Store(false)
			}
		}
	}

	// render loop
	c.win.MainLoop(func() {
		// nothing to render
		if c.curr.Ant == nil {
			return
		}

		c.lock.Lock()

		// clear screen
		c.cv.SetFillStyle("#FFF")
		c.cv.FillRect(0, 0, float64(c.cw), float64(c.ch))

		// compute extend of antenna
		extend := 0.
		for _, seg := range c.curr.Ant.segs {
			extend += seg.Length()
		}
		c.rescale(0.6 * extend)

		y := 2*c.txtSize - c.h/2
		if len(c.curr.Msg) > 0 {
			c.Text(0, y, c.txtSize, c.curr.Msg, ClrBlack)
		} else {
			c.Text(0, y, c.txtSize, fmt.Sprintf("Step #%d", c.count), ClrBlack)
		}
		for idx, seg := range c.curr.Ant.segs {
			clr := ClrBlue
			if idx == c.curr.Ant.excite {
				clr = ClrRed
			}
			c.Line(seg.start[0], seg.start[1], seg.end[0], seg.end[1], c.curr.Ant.dia, clr)
		}
		if c.curr.Pos >= 0 {
			p := c.curr.Ant.segs[2*c.curr.Pos+1].Start()
			c.Circle(p[0], p[1], c.txtSize/6, 0, nil, ClrGreen)
			c.Circle(-p[0], p[1], c.txtSize/6, 0, nil, ClrGreen)
		}
		y += c.txtSize
		c.Text(0, y, c.txtSize/2, c.curr.Ant.Perf.String(), ClrRed)

		y += c.txtSize
		k := extend / c.curr.Ant.Lambda
		info := fmt.Sprintf("%d segments, length: %.3fm (%.3f λ)", len(c.curr.Ant.segs), extend, k)
		c.Text(0, y, c.txtSize/2, info, ClrBlack)

		y = c.h/2 - 2*c.txtSize
		c.Text(0, y, c.txtSize/2, c.hint, ClrPink)

		c.lock.Unlock()
	})
}

// Line primitive
func (c *SDLCanvas) Line(x1, y1, x2, y2, w float64, clr *color.RGBA) {
	cx1, cy1 := c.xlate(x1, y1)
	cx2, cy2 := c.xlate(x2, y2)
	cw := c.scale * w
	c.cv.SetStrokeStyle(clr.R, clr.G, clr.B)
	c.cv.SetLineWidth(cw)
	c.cv.BeginPath()
	c.cv.MoveTo(cx1, cy1)
	c.cv.LineTo(cx2, cy2)
	c.cv.ClosePath()
	c.cv.Stroke()
}

// Circle primitive
func (c *SDLCanvas) Circle(x, y, r, w float64, clrBorder, clrFill *color.RGBA) {
	cx, cy := c.xlate(x, y)
	cr := c.scale * r
	cw := c.scale * w
	if clrFill != nil {
		c.cv.SetFillStyle(clrFill.R, clrFill.G, clrFill.B)
		c.cv.BeginPath()
		c.cv.Arc(cx, cy, cr, 0, math.Pi*2, false)
		c.cv.ClosePath()
		c.cv.Fill()
	}
	if clrBorder != nil {
		c.cv.SetStrokeStyle(clrBorder.R, clrBorder.G, clrBorder.B)
		c.cv.SetLineWidth(cw)
		c.cv.BeginPath()
		c.cv.Arc(cx, cy, cr, 0, math.Pi*2, false)
		c.cv.ClosePath()
		c.cv.Stroke()
	}
}

// Text primitive
func (c *SDLCanvas) Text(x, y, fs float64, s string, clr *color.RGBA) {
	cx, cy := c.xlate(x, y)
	cfs := c.scale * fs
	c.cv.SetFillStyle(clr.R, clr.G, clr.B)
	c.cv.SetTextAlign(canvas.Center)
	c.cv.SetTextBaseline(canvas.Middle)
	c.cv.SetFont(nil, cfs)
	c.cv.FillText(s, cx, cy)
}

// Dump canvas to file
func (c *SDLCanvas) Dump(fName string) error {
	return nil
}

// coordinate translation
func (c *SDLCanvas) xlate(x, y float64) (float64, float64) {
	return x*c.scale + c.offX, y*c.scale + c.offY
}
