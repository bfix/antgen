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
	"bytes"
	"fmt"
	"image/color"
	"os"

	svg "github.com/ajstarks/svgo"
)

//----------------------------------------------------------------------
// SVG canvas
//----------------------------------------------------------------------

// SVGCanvas for writing SVG streams
type SVGCanvas struct {
	svg        *svg.SVG
	prec       float64 // precision 0.01mm
	offX, offY float64
	margin     int
	txtSize    float64
	buf        *bytes.Buffer
}

// NewSVGCanvas creates a new SVG canvas
func NewSVGCanvas(_, _ int, _ float64) (*SVGCanvas, error) {
	c := new(SVGCanvas)
	c.buf = new(bytes.Buffer)
	c.prec = 1e-5
	c.txtSize = 0.1
	c.margin = int(0.1 / c.prec)
	c.svg = svg.New(c.buf)
	return c, nil
}

// Perform rendering
func (c *SVGCanvas) Run(cb Action) {}

func (c *SVGCanvas) SetHint(m string) {}

// Show antenna on canvas
func (c *SVGCanvas) Show(ant *Antenna, _ int, msg string) {

	// compute bounding box and antenna length
	box := NewBoundingBox()
	length := 0.
	for _, seg := range ant.segs {
		length += seg.Length()
		p := seg.Start()
		box.Include(p)
		p = seg.End()
		box.Include(p)
	}
	// width and height of SVG canvas
	width := int((box.Xmax - box.Xmin) / c.prec)
	height := int((box.Ymax - box.Ymin) / c.prec)
	c.offX, c.offY = box.Xmin, box.Ymin

	c.svg.Start(width+2*c.margin, height+2*c.margin)

	y := box.Ymax + 2*c.txtSize
	if len(msg) > 0 {
		c.Text(0, y, c.txtSize, msg, ClrBlack)
	}
	for idx, seg := range ant.segs {
		clr := ClrBlue
		if idx == ant.excite {
			clr = ClrRed
		}
		c.Line(seg.start[0], seg.start[1], seg.end[0], seg.end[1], ant.dia, clr)
	}
	y += c.txtSize
	c.Text(0, y, c.txtSize/2, ant.Perf.String(), ClrRed)
	c.svg.End()
}

// Circle primitive
func (c *SVGCanvas) Circle(x, y, r, w float64, clrBorder, clrFill *color.RGBA) {
	fill := "none"
	if clrFill != nil {
		fill = fmt.Sprintf("#%02x%02x%02x", clrFill.R, clrFill.G, clrFill.B)
	}
	border := ""
	if w > 0 && clrBorder != nil {
		border = fmt.Sprintf("stroke:#%02x%02x%02x;stroke-width:%d;",
			clrBorder.R, clrBorder.G, clrBorder.B, int(w/c.prec))
	}
	style := fmt.Sprintf("%sfill:%s", border, fill)
	cx, cy := c.xlate(x, y)
	c.svg.Circle(cx, cy, int(r/c.prec), style)
}

// Text primitive
func (c *SVGCanvas) Text(x, y, fs float64, s string, clr *color.RGBA) {
	style := fmt.Sprintf("text-anchor:middle;font-size:%dpx", int(fs/c.prec))
	cx, cy := c.xlate(x, y)
	c.svg.Text(cx, cy, s, style)
}

// Line primitive
func (c *SVGCanvas) Line(x1, y1, x2, y2, w float64, clr *color.RGBA) {
	style := "stroke:black;stroke-width:1"
	if w > 0 && clr != nil {
		style = fmt.Sprintf("stroke:#%02x%02x%02x;stroke-width:%d;",
			clr.R, clr.G, clr.B, int(w/c.prec))
	}
	cx1, cy1 := c.xlate(x1, y1)
	cx2, cy2 := c.xlate(x2, y2)
	c.svg.Line(cx1, cy1, cx2, cy2, style)
}

// coordinate translation
func (c *SVGCanvas) xlate(x, y float64) (int, int) {
	return int((x-c.offX)/c.prec) + c.margin, int((y-c.offY)/c.prec) + c.margin
}

// Close a canvas. No further operations are allowed
func (c *SVGCanvas) Close() (err error) {
	c.buf = nil
	return
}

// Dump canvas to file
func (c *SVGCanvas) Dump(fName string) (err error) {
	var f *os.File
	if f, err = os.Create(fName); err != nil {
		return
	}
	defer f.Close()
	_, err = f.Write(c.buf.Bytes())
	return nil
}
