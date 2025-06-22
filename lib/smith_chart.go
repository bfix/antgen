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
	"math"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

// steps and bounds of Smith curves
var (
	steps  = []float64{0.1, 0.2, 0.5, 1.0, 2.0, 5.0, 10.0}
	bounds = []float64{0.5, 1.0, 2.0, 5.0, 10.0, 20.0, 50.0}
)

// SmithChart holds tracks (sequences of impedances) for plotting.
// It implements the plot.Plotter interface.
type SmithChart struct {
	tracks [][]complex128
}

// Plot is a plot.Plotter implementation
func (sc *SmithChart) Plot(c draw.Canvas, plt *plot.Plot) {
	// draw the layout
	sc.constRG(c, 0)
	sc.constXB(c, 0, 1000)
	for i, step := range steps {
		// constant resistance (circles)
		sc.constRG(c, step)
		// constant reactance (curves)
		sc.constXB(c, step, bounds[i])
		sc.constXB(c, -step, bounds[i])
	}
	// focus
	pnts := []vg.Point{
		{X: c.X(0.50), Y: c.Y(0.51)},
		{X: c.X(0.51), Y: c.Y(0.50)},
		{X: c.X(0.50), Y: c.Y(0.49)},
		{X: c.X(0.49), Y: c.Y(0.50)},
		{X: c.X(0.50), Y: c.Y(0.51)},
	}
	sty := draw.LineStyle{
		Width:  vg.Points(1),
		Dashes: []vg.Length{},
		Color:  color.RGBA{R: 0, G: 255, B: 255, A: 255},
	}
	c.StrokeLines(sty, pnts)

	// plot track
	z0 := complex(50, 0)
	for idx, track := range sc.tracks {
		pnts := make([]vg.Point, 0)
		for _, z := range track {
			// convert to Smith coordinates
			g := (z - z0) / (z + z0)
			x := c.X((real(g) + 1) / 2)
			y := c.Y((imag(g) + 1) / 2)
			pt := vg.Point{X: x, Y: y}
			pnts = append(pnts, pt)
		}
		_, sty := PlotStyle(idx)
		c.StrokeLines(sty, pnts)
	}
}

// plot curves of constant reactance/susceptance
func (sc *SmithChart) constXB(c draw.Canvas, step float64, bounds float64) {
	pnts := make([]vg.Point, 0)
	k, f := 0., 0.1
	done := false
	x0 := 2 * c.X(0.5)
	for {
		z := complex(k, step)
		g := (z - 1) / (z + 1)
		x := c.X((real(g) + 1) / 2)
		y := c.Y((imag(g) + 1) / 2)
		pt := vg.Point{X: x, Y: y}
		pnts = append(pnts, pt)
		if done {
			break
		}
		k += (f * Sqr(k+1)) / (k + 2 - f*(k+1))
		if k > bounds {
			k = bounds
			done = true
		}
	}
	// draw reactance
	sty := draw.LineStyle{
		Width:  vg.Points(1),
		Dashes: []vg.Length{},
		Color:  color.RGBA{R: 224, G: 224, B: 224, A: 255},
	}
	c.StrokeLines(sty, pnts)

	// draw susceptance
	sty.Dashes = []vg.Length{vg.Points(2), vg.Points(2)}
	for i, pnt := range pnts {
		pnts[i] = vg.Point{X: x0 - pnt.X, Y: pnt.Y}
	}
	c.StrokeLines(sty, pnts)
}

// plot curves (circles) of constant resistance/conductance
func (sc *SmithChart) constRG(c draw.Canvas, step float64) {
	rad := (1 - (step-1)/(step+1)) / 2
	xr := 1 - rad
	x0 := 2 * c.X(0.5)
	pnts := make([]vg.Point, 0)
	ang, dAng := 0.0, 0.05/rad
	done := false
	for {
		xk := xr + rad*math.Cos(ang)
		yk := rad * math.Sin(ang)
		x := c.X((xk + 1) / 2)
		y := c.Y((yk + 1) / 2)
		pt := vg.Point{X: x, Y: y}
		pnts = append(pnts, pt)
		if done {
			break
		}
		if ang += dAng; ang > CircAng {
			ang = CircAng
			done = true
		}
	}

	// draw resistance
	sty := draw.LineStyle{
		Width:  vg.Points(1),
		Dashes: []vg.Length{},
		Color:  color.RGBA{R: 224, G: 224, B: 224, A: 255},
	}
	c.StrokeLines(sty, pnts)

	// draw susceptance
	sty.Dashes = []vg.Length{vg.Points(2), vg.Points(2)}
	for i, pnt := range pnts {
		pnts[i] = vg.Point{X: x0 - pnt.X, Y: pnt.Y}
	}
	c.StrokeLines(sty, pnts)
}
