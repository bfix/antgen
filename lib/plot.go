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
	"io"
	"math"
	"slices"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/palette/moreland"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

// list of plot targets
var PlotValues = []string{
	// performance value
	"Gmax",  // maximum gain
	"Gmean", // mean gain
	"SD",    // standard deviation
	"Zr",    // Resistance (Impedance)
	"Zi",    // Reactance (Impedance)

	// derived performance
	"Geff",   // maximum gain of matched antenna
	"Loss",   // loss due to impedance mismatch
	"PwrFac", // loss/inefficiency due to phase shift between U and I
}

// special graphs
var PlotSpecial = []string{
	"Smith",
}

//----------------------------------------------------------------------

type PlotSet struct {
	Tag   string    // plot tag
	Dir   string    // dataset directory
	Klist []float64 // list of possible 'k' values
	Plist []float64 // list of possible 'param' values
	Kidx  int       // 'k' selection in KList
	Pidx  int       // 'param' selection in PList
}

func NewPlotSet(dir string) *PlotSet {
	return &PlotSet{
		Dir:  dir,
		Kidx: -1,
		Pidx: -1,
	}
}

func (ds *PlotSet) Index(val float64, name string) int {
	switch name {
	case "k":
		return slices.Index(ds.Klist, val)
	case "param":
		return slices.Index(ds.Plist, val)
	}
	return -1
}

func (ds *PlotSet) Params() (k, param float64) {
	if ds.Kidx == -1 {
		k = math.NaN()
	} else {
		k = ds.Klist[ds.Kidx]
	}
	if ds.Pidx == -1 || ds.Pidx >= len(ds.Plist) {
		param = math.NaN()
		ds.Pidx = -1
	} else {
		param = ds.Plist[ds.Pidx]
	}
	return
}

//----------------------------------------------------------------------

// NumPlots is the number of plot selections available in the GUI
const NumPlots = 15

// Selection of plot parameters
type Selection struct {
	Target string             // Parameter (Gmax,Gmean,Zr,Zi)
	Sets   [NumPlots]*PlotSet // list of PlotSets selected
}

// NewSelection for given target
func NewSelection(target string) *Selection {
	return &Selection{
		Target: target,
	}
}

//----------------------------------------------------------------------

// Pre-defined colors for plotting
var (
	clrs = []color.RGBA{
		{R: 255, G: 0, B: 0, A: 255},
		{R: 0, G: 0, B: 255, A: 255},
		{R: 192, G: 0, B: 192, A: 255},
		{R: 0, G: 192, B: 0, A: 255},
		{R: 255, G: 192, B: 0, A: 255},
	}
	styles = []draw.LineStyle{
		{Width: vg.Points(1), Dashes: []vg.Length{}},
		{Width: vg.Points(1), Dashes: []vg.Length{vg.Points(5), vg.Points(3)}},
		{Width: vg.Points(1), Dashes: []vg.Length{vg.Points(2), vg.Points(2)}},
	}
	patterns = []string{"―――", "- - - -", "·······"}
)

func PlotStyle(pos int) (pat string, style draw.LineStyle) {
	idx := (pos / len(clrs)) % len(styles)
	style = styles[idx]
	style.Color = clrs[pos%len(clrs)]
	pat = patterns[idx]
	return
}

// Plotter for AntGen datasets
func Plotter(db *Database, sel *Selection, format string) (out map[string]string, err error) {
	// check for heatmap graph
	num := 0
	heatmap := false
	idx := -1
	for i, ps := range sel.Sets {
		if ps != nil {
			num++
			idx = i
			heatmap = (num == 1 && len(ps.Klist) > 1 && len(ps.Plist) > 1)
		}
	}
	if heatmap && !slices.Contains(PlotValues, sel.Target) {
		// heatmap not possible
		heatmap = false
	}
	// create plot
	var p *plot.Plot
	out = make(map[string]string)
	if heatmap {
		p, err = plotHeatmap(db, sel, idx)
		if err == nil {
			// handle legend separately (ignored by plot)
			out["legend"], err = plotLegend(p.Legend, 3, 18, format)
		}
		p.Legend = plot.NewLegend()
	} else {
		p, err = plotGraph(db, sel)
	}
	if err != nil {
		return
	}
	// create plot output
	var wrt io.WriterTo
	if wrt, err = p.WriterTo(18*vg.Centimeter, 18*vg.Centimeter, format); err != nil {
		return
	}
	buf := new(bytes.Buffer)
	if _, err = wrt.WriteTo(buf); err != nil {
		return
	}
	out["plot"] = buf.String()
	return
}

// Simple graph plot (2D with lines)
func plotGraph(db *Database, sel *Selection) (*plot.Plot, error) {
	// generate plot for value
	if slices.Contains(PlotValues, sel.Target) {
		return plotXY(db, sel)
	}
	// handle special plots
	switch sel.Target {
	case "Smith":
		return plotSmith(db, sel)
	}
	// unknown plot target
	return nil, fmt.Errorf("unhandled plot target '%s'", sel.Target)
}

// Simple X-Y-plot
func plotXY(db *Database, sel *Selection) (p *plot.Plot, err error) {
	// collect data sets
	data := make([]*Set, len(sel.Sets))
	tags := make([]string, len(sel.Sets))
	for i, ps := range sel.Sets {
		if ps == nil {
			continue
		}
		if len(ps.Tag) == 0 {
			tags[i] = fmt.Sprintf("#%d", i)
		} else {
			tags[i] = ps.Tag
		}
		filter := NewIndex(ps.Params())
		if data[i], err = db.Set(ps.Dir, filter); err != nil {
			return
		}
	}
	// make sure all sets vary the same way
	var varying int
	sweep := NewIndexList()
	tagList := make([]string, 0)
	refList := make([]int, 0)
	first := true
	for i, set := range data {
		if set == nil {
			continue
		}
		refList = append(refList, i)
		tagList = append(tagList, tags[i])
		if first {
			varying = set.Varying(sweep)
			first = false
			continue
		}
		if set.Varying(sweep) != varying {
			err = fmt.Errorf("set '%s' not compatible", tags[i])
			return
		}
	}

	// create new table
	tbl := new(Table)
	tbl.Name = sel.Target

	// assemble column header
	tbl.Dims = make([]string, 0)
	tbl.Refs = make([]int, 0)
	if varying&VaryK != 0 {
		tbl.Dims = append(tbl.Dims, "k")
		tbl.Refs = append(tbl.Refs, -1)
		tbl.NumIdx++
	}
	if varying&VaryP != 0 {
		tbl.Dims = append(tbl.Dims, "param")
		tbl.Refs = append(tbl.Refs, -1)
		tbl.NumIdx++
	}
	tbl.Dims = append(tbl.Dims, tagList...)
	tbl.Refs = append(tbl.Refs, refList...)

	// collect table values
	tbl.Vals = make([][]any, 0)
	for _, idx := range sweep.Sorted() {
		valList := make([]any, 0)
		if varying&VaryK != 0 {
			valList = append(valList, idx.K())
		}
		if varying&VaryP != 0 {
			valList = append(valList, idx.Param())
		}
		for _, tag := range tagList {
			pos := slices.Index(tags, tag)
			val := data[pos].Value(idx, sel.Target)
			valList = append(valList, val)
		}
		tbl.Vals = append(tbl.Vals, valList)
	}

	// plot table
	p = plot.New()
	p.Title.Text = tbl.Name
	p.X.Label.Text = tbl.Dims[0]
	p.Y.Label.Text = ""

	numCols, numRows := len(tbl.Dims), len(tbl.Vals)
	var graph *plotter.Line
	for col := tbl.NumIdx; col < numCols; col++ {
		// convert table data to plotter.Values
		data := make(plotter.XYs, 0)
		for row := range numRows {
			val := TblValue[float64](tbl, row, col)
			if math.IsNaN(val) {
				continue
			}
			data = append(data, plotter.XY{
				X: TblValue[float64](tbl, row, 0),
				Y: val,
			})
		}
		if graph, err = plotter.NewLine(data); err != nil {
			return
		}
		_, graph.LineStyle = PlotStyle(tbl.Refs[col])
		p.Add(graph)
		p.Legend.Add(tbl.Dims[col], graph)
	}

	// handle optional value ticks/lines
	Ytgt := math.NaN()
	switch sel.Target {
	case "Zr":
		Ytgt = 50
		p.Legend.Top = true
	case "Zi":
		Ytgt = 0
	}
	if !math.IsNaN(Ytgt) {
		data := plotter.XYs{
			plotter.XY{
				X: TblValue[float64](tbl, 0, 0),
				Y: Ytgt,
			},
			plotter.XY{
				X: TblValue[float64](tbl, numRows-1, 0),
				Y: Ytgt,
			},
		}
		if graph, err = plotter.NewLine(data); err != nil {
			return
		}
		graph.LineStyle = draw.LineStyle{
			Width: vg.Points(1), Color: color.RGBA{0, 0, 0, 255},
			Dashes: []vg.Length{vg.Points(5), vg.Points(3)}}
		p.Add(graph)
	}
	return
}

//----------------------------------------------------------------------

// Grid implements the interface for heatmap grids
type Grid struct {
	target  string
	dataset *Set
	plotset *PlotSet
}

// NewGrid instantiates a new grid object from database
func NewGrid(db *Database, sel *Selection, idx int) (g *Grid, err error) {
	g = new(Grid)
	g.target = sel.Target
	g.plotset = sel.Sets[idx]
	filter := NewIndex(g.plotset.Params())
	g.dataset, err = db.Set(g.plotset.Dir, filter)
	return
}

// Dims returns the grid dimensions
func (g *Grid) Dims() (c, r int) {
	return len(g.plotset.Klist), len(g.plotset.Plist)
}

// X returns the x-axis value in grid column c
func (g *Grid) X(c int) float64 {
	return g.plotset.Klist[c]
}

// Y returns the y-axis value in grid row r
func (g *Grid) Y(r int) float64 {
	return g.plotset.Plist[r]
}

// Z returns the target value in grid cell
func (g *Grid) Z(c, r int) float64 {
	idx := NewIndex(g.X(c), g.Y(r))
	return g.dataset.Value(idx, g.target)
}

// Plot heatmap from plotset
func plotHeatmap(db *Database, sel *Selection, idx int) (p *plot.Plot, err error) {
	// build heatmap
	var g *Grid
	if g, err = NewGrid(db, sel, idx); err != nil {
		return
	}
	// create heatmap
	pal := moreland.SmoothBlueRed().Palette(30)
	hm := plotter.NewHeatMap(g, pal)

	// assemble plot
	p = plot.New()
	p.Title.Text = sel.Target
	p.Add(hm)

	// Create a legend.
	thumbs := plotter.PaletteThumbnailers(pal)
	for i := len(thumbs) - 1; i >= 0; i-- {
		t := thumbs[i]
		if i != 0 && i != len(thumbs)-1 {
			p.Legend.Add("", t)
			continue
		}
		var val float64
		switch i {
		case 0:
			val = hm.Min
		case len(thumbs) - 1:
			val = hm.Max
		}
		p.Legend.Add(fmt.Sprintf("%.2g", val), t)
	}
	return
}

// Plot legend separately (used with heatmaps)
func plotLegend(legend plot.Legend, width, height float64, format string) (out string, err error) {
	// create plot
	p := plot.New()
	p.Legend = legend
	p.HideAxes()

	// generate plot output
	var wrt io.WriterTo
	if wrt, err = p.WriterTo(vg.Length(width)*vg.Centimeter, vg.Length(height)*vg.Centimeter, format); err != nil {
		return
	}
	buf := new(bytes.Buffer)
	if _, err = wrt.WriteTo(buf); err != nil {
		return
	}
	out = buf.String()
	return
}

// plot Smith chart for selections
func plotSmith(db *Database, sel *Selection) (p *plot.Plot, err error) {
	// assemle Smith chart
	sc := new(SmithChart)

	// collect data sets
	data := make([]*Set, len(sel.Sets))
	tags := make([]string, len(sel.Sets))
	for i, ps := range sel.Sets {
		if ps == nil {
			continue
		}
		if len(ps.Tag) == 0 {
			tags[i] = fmt.Sprintf("#%d", i)
		} else {
			tags[i] = ps.Tag
		}
		filter := NewIndex(ps.Params())
		if data[i], err = db.Set(ps.Dir, filter); err != nil {
			return
		}
	}
	// make sure all sets vary the same way (only one free parameter)
	var varying int
	sweep := NewIndexList()
	tagList := make([]string, 0)
	refList := make([]int, 0)
	first := true
	for i, set := range data {
		if set == nil {
			continue
		}
		refList = append(refList, i)
		tagList = append(tagList, tags[i])
		if first {
			varying = set.Varying(sweep)
			first = false
			continue
		}
		if set.Varying(sweep) != varying {
			err = fmt.Errorf("set '%s' not compatible", tags[i])
			return
		}
	}
	if varying == 3 {
		err = fmt.Errorf("no multiple parameters allowed")
		return
	}

	// create new table
	tbl := new(Table)
	tbl.Name = sel.Target

	// assemble column header
	tbl.Dims = make([]string, 0)
	tbl.Refs = make([]int, 0)
	if varying&VaryK != 0 {
		tbl.Dims = append(tbl.Dims, "k")
		tbl.Refs = append(tbl.Refs, -1)
		tbl.NumIdx++
	}
	if varying&VaryP != 0 {
		tbl.Dims = append(tbl.Dims, "param")
		tbl.Refs = append(tbl.Refs, -1)
		tbl.NumIdx++
	}
	tbl.Dims = append(tbl.Dims, tagList...)
	tbl.Refs = append(tbl.Refs, refList...)

	// collect table values
	tbl.Vals = make([][]any, 0)
	for _, idx := range sweep.Sorted() {
		valList := make([]any, 0)
		if varying&VaryK != 0 {
			valList = append(valList, idx.K())
		}
		if varying&VaryP != 0 {
			valList = append(valList, idx.Param())
		}
		for _, tag := range tagList {
			pos := slices.Index(tags, tag)
			zr := data[pos].Value(idx, "Zr")
			zi := data[pos].Value(idx, "Zi")
			valList = append(valList, complex(zr, zi))
		}
		tbl.Vals = append(tbl.Vals, valList)
	}

	sc.tracks = make([][]complex128, 0)
	numCols, numRows := len(tbl.Dims), len(tbl.Vals)
	for col := tbl.NumIdx; col < numCols; col++ {
		track := make([]complex128, 0)
		for row := range numRows {
			// get impedance from table
			z := TblValue[complex128](tbl, row, col)
			if math.IsNaN(real(z)) {
				continue
			}
			track = append(track, z)
		}
		sc.tracks = append(sc.tracks, track)
	}
	return plotSmithRaw(sc)
}

// plot raw Smith chart
func plotSmithRaw(sc *SmithChart) (p *plot.Plot, err error) {
	// create plot
	p = plot.New()
	p.HideAxes()
	p.Add(sc)
	return
}
