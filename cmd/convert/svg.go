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

package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bfix/antgen/lib"
	"github.com/twpayne/go-svg"
	"github.com/twpayne/go-svg/svgpath"
)

// convert geometry to SVG file
func convert2SVG(fGeo, fOut string, geo *lib.Geometry, spec *lib.Specification, v float64) (err error) {
	// set output filename if not given
	if len(fOut) == 0 {
		fOut = fGeo + ".svg"
	}
	// scaling factor
	f := 1000 * v

	// extract title and description from comments
	var title svg.CharData
	var desc []svg.Element
	for _, s := range geo.Cmts {
		if strings.HasPrefix(s, "Antgen") {
			title = svg.CharData(s)
			continue
		}
		p := strings.Split(s, ":")
		switch p[0] {
		case "Spec", "Param", "Init", "Result", "Stats":
			desc = append(desc, svg.CharData(s))
		}
	}

	// build geometry:
	// (1) dipole leg as a "line" (sequence of 2D points)
	// (2) "holes" (every five segments or if curvature is above limit)
	var line, holes []lib.Vec3
	pos := lib.NewVec3(0, 0, 0)
	line = append(line, pos)
	holes = append(holes, pos)
	hStep := 0
	lastHole := pos
	dir := 0.
	bb := lib.NewBoundingBox()
	bb.Include(pos)
	for _, node := range geo.Nodes {
		dir += node.Theta
		end := pos.Move2D(node.Length, dir)
		line = append(line, end)
		hStep++
		deviation := float64(hStep) * node.Length / end.Sub(lastHole).Length()
		if hStep == 5 || deviation > 1.02 {
			hStep = 0
			holes = append(holes, end)
			lastHole = end
		}
		bb.Include(end)
		pos = end
	}
	holes = append(holes, pos)

	log.Printf("BoundingBox: (%.2f,%.2f) - (%.2f,%.2f)",
		f*bb.Xmin, f*bb.Ymin, f*bb.Xmax, f*bb.Ymax)

	// convert to SVG path
	scale := func(p lib.Vec3) []float64 {
		return []float64{f * p[0], f * p[1]}
	}
	path := svgpath.New()
	path.MoveToAbs(scale(line[0]))
	for _, p := range line[1:] {
		path.LineToAbs(scale(p))
	}
	style := svg.String(fmt.Sprintf(
		"stroke:#000000;stroke-opacity:1;stroke-width:%.2f;stroke-dasharray:none",
		1000*spec.Wire.Diameter))
	leg := svg.Path().
		Style(style).
		Fill("none").
		D(path)

	// place hole markers
	var circles []svg.Element
	for _, hole := range holes {
		p := scale(hole)
		circ := svg.Circle().CXCYR(p[0], p[1], 2.5, svg.Number).Fill("none").Stroke("black")
		circles = append(circles, circ)
	}
	// create SVG
	graph := svg.New()
	w, h := f*(bb.Xmax-bb.Xmin), f*(bb.Ymax-bb.Ymin)
	log.Printf("Width= %.3fmm, Height=%.3fmm", w, h)
	graph.WidthHeight(w, h, svg.MM)
	graph.ViewBox(f*bb.Xmin, f*bb.Ymin, w, h)
	graph.AppendChildren(
		svg.Title(title),
		svg.Desc(desc...),
		leg,
	)
	graph.AppendChildren(circles...)

	// output SVG file
	var fp *os.File
	if fp, err = os.Create(fOut); err != nil {
		return
	}
	if _, err = graph.WriteToIndent(fp, "", "  "); err != nil {
		return
	}
	err = fp.Close()
	return
}
