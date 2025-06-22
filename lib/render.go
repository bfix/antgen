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
	"image/color"
)

// Color definitions for drawing
var (
	ClrWhite = &color.RGBA{255, 255, 255, 0}
	ClrRed   = &color.RGBA{255, 0, 0, 0}
	ClrRedTr = &color.RGBA{255, 0, 0, 224}
	ClrPink  = &color.RGBA{255, 0, 255, 0}
	ClrBlack = &color.RGBA{0, 0, 0, 0}
	ClrGray  = &color.RGBA{127, 127, 127, 0}
	ClrBlue  = &color.RGBA{0, 0, 255, 0}
	ClrGreen = &color.RGBA{0, 255, 0, 0}
	ClrCyan  = &color.RGBA{0, 255, 255, 0}
)

// Callback on key press
type Action func(ant *Antenna, key rune, step int) bool

// Canvas for drawing the antenna
type Canvas interface {
	// Start a new (dynamic) rendering
	Run(Action)

	// Show antenna
	Show(ant *Antenna, pos int, msg string)

	SetHint(m string)

	// Circle primitive
	Circle(x, y, r, w float64, clrBorder, clrFill *color.RGBA)

	// Text primitive
	Text(x, y, fs float64, s string, clr *color.RGBA)

	// Line primitive
	Line(x1, y1, x2, y2, w float64, clr *color.RGBA)

	// Dump canvas to file
	Dump(fName string) error

	// Close a canvas. No further operations are allowed
	Close() error
}

// GetCanvas returns a canvas for drawing (factory)
func GetCanvas(kind string, width, height int, side float64) (c Canvas, err error) {
	switch kind {
	case "svg":
		return NewSVGCanvas(width, height, side)
	case "sdl":
		return NewSDLCanvas(width, height, side)
	}
	return
}

// GetCanvasFromCfg returns a canvas from configuration
func GetCanvasFromCfg(cfg *RenderConfig, side float64) (Canvas, error) {
	return GetCanvas(cfg.Canvas, cfg.Width, cfg.Height, side)
}
