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
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/bfix/antgen/lib"
)

// Plot data from database
func plotToFile(db *lib.Database, _ string, args []string) {
	var (
		target string
		sets   string
		fOut   string
	)
	fs := flag.NewFlagSet("plot", flag.ContinueOnError)
	fs.StringVar(&target, "target", "Gmax", "plot target")
	fs.StringVar(&sets, "sets", "", "plot sets")
	fs.StringVar(&fOut, "out", "out.svg", "output file (SVG)")
	fs.Parse(args)

	// build selection
	sel := lib.NewSelection(target)

	// get plot sets
	s := strings.Split(sets, ",")
	if len(s) == 0 {
		log.Fatal("missing plot sets")
	}
	if len(s) > lib.NumPlots {
		log.Printf("WARN: plot sets trimmed to %d", lib.NumPlots)
		s = s[:lib.NumPlots]
	}
	for i, set := range s {
		t := strings.Split(set, ":")
		sel.Sets[i] = &lib.PlotSet{
			Tag:  t[0],
			Dir:  t[1],
			Kidx: -1,
			Pidx: -1,
		}
	}
	out, err := lib.Plotter(db, sel, "svg")
	if err != nil {
		log.Fatal(err)
	}
	buf := []byte(out["plot"])
	if err = os.WriteFile(fOut, buf, 0644); err != nil {
		log.Fatal(err)
	}
}

//======================================================================
// handle plot request
//======================================================================

// persistent (single) user selection
var sel lib.Selection

// Message as a response from the handler
type Message struct {
	Mode string // mode ["ERROR", "WARN", "INFO"]
	Text string // message text
}

// PlotData holds all information to render the view
type PlotData struct {
	Prefix string       // URL prefix
	Stats  *lib.DbStats // database statistics

	Targets []string                // list of possible plot targets
	Sets    map[string]*lib.PlotSet // list of available plot sets
	Styles  [lib.NumPlots]string    // list of plot styles

	Select *lib.Selection    // current selection
	Graphs map[string]string // SVG-encoded graphs
	Msgs   []*Message        // list of messages
}

// AddMsg to list
func (pd *PlotData) AddMsg(mode, text string) {
	pd.Msgs = append(pd.Msgs, &Message{mode, text})
}

// handle request (main entry page)
func plotHandler(w http.ResponseWriter, r *http.Request) {
	pd := new(PlotData)
	pd.Stats = db.Stats()
	pd.Msgs = make([]*Message, 0)
	for i := range pd.Styles {
		pat, ls := lib.PlotStyle(i)
		R, G, B, _ := ls.Color.RGBA()
		s := fmt.Sprintf("<td style='background-color: #%02x%02x%02x'>%s</td>", R>>8, G>>8, B>>8, pat)
		pd.Styles[i] = s
	}

	// check for POST request: generate plot from user settings if true
	if r.Method == "POST" {
		// parse form data
		err := r.ParseForm()
		if err != nil {
			pd.AddMsg("ERROR", "ParseForm: "+err.Error())
		}
		// collect user settings
		var keys []string
		values := make(map[string]string)
		for key, v := range r.PostForm {
			values[key] = v[0]
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

		// handle user settings
		for _, key := range keys {
			value := values[key]
			parts := strings.Split(key, "_")
			switch parts[0] {
			case "target":
				// value to be plotted
				sel.Target = value
			case "plotset":
				var idx int
				if idx, err = strconv.Atoi(parts[1]); err != nil {
					pd.AddMsg("ERROR", "plotset index: "+err.Error())
					break
				}
				ps := sel.Sets[idx]
				if ps == nil {
					ps = lib.NewPlotSet("")
					sel.Sets[idx] = ps
				}
				switch parts[2] {
				// parameters
				case "k":
					if len(value) == 0 {
						ps.Kidx = -1
						break
					}
					var v float64
					if v, err = strconv.ParseFloat(value, 64); err != nil {
						pd.AddMsg("ERROR", "Option 'k': "+err.Error())
						ps.Kidx = -1
					} else {
						ps.Kidx = ps.Index(v, "k")
					}
				case "param":
					if len(value) == 0 {
						ps.Pidx = -1
						break
					}
					var v float64
					if v, err = strconv.ParseFloat(value, 64); err != nil {
						pd.AddMsg("ERROR", "Option 'param': "+err.Error())
						ps.Pidx = -1
					} else {
						ps.Pidx = ps.Index(v, "param")
					}
				case "tag":
					ps.Tag = value
				case "dir":
					ps.Dir = value
				}
			}
		}
		// set parameter ranges and remove empty plot sets
		for i, ps := range sel.Sets {
			if ps == nil {
				continue
			}
			if len(ps.Dir) == 0 {
				sel.Sets[i] = nil
				continue
			}
			if s, ok := sets[ps.Dir]; ok {
				ps.Klist = s.Klist
				ps.Plist = s.Plist
			} else {
				pd.AddMsg("ERROR", "setting parameter ranges")
			}
		}
		// create plot
		if pd.Graphs, err = lib.Plotter(db, &sel, "svg"); err != nil {
			pd.AddMsg("ERROR", err.Error())
		}
	}
	// collect information for view
	pd.Prefix = prefix
	pd.Select = &sel
	pd.Targets = append(lib.PlotValues, lib.PlotSpecial...)
	pd.Sets = sets

	// show plot view
	renderPage(w, pd, "plot")
}

//======================================================================
// Helper methods
//======================================================================

// render a webpage with given data and template reference
func renderPage(w io.Writer, data interface{}, body string) {
	// create content section
	t := tpl.Lookup(body)
	if t == nil {
		io.WriteString(w, "No template '"+body+"' found")
		return
	}
	content := new(bytes.Buffer)
	if err := t.Execute(content, data); err != nil {
		io.WriteString(w, err.Error())
		return
	}
	// emit final page
	t = tpl.Lookup("main")
	if t == nil {
		io.WriteString(w, "No main template found")
		return
	}
	if err := t.Execute(w, content.String()); err != nil {
		io.WriteString(w, err.Error())
	}
}
