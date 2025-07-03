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
	"encoding/json"
	"flag"
	"log"
	"os"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/bfix/antgen/lib"
)

// show models with best performance
func showBest(db *lib.Database, in string, args []string) {
	// handle command-line arguments
	var (
		target string // opt. parameter
		band   string // frequency band
		zRange string // impedance range [min_Zr,max_Zr,|Zi|]

		spec = new(lib.Specification)
		err  error
	)
	fs := flag.NewFlagSet("best", flag.ContinueOnError)
	fs.StringVar(&target, "target", "Gmax", "opt. parameter")
	fs.StringVar(&band, "band", "2m", "frequency band")
	fs.StringVar(&zRange, "zRange", "any", "impedance range: [min_Zr,max_Zr,|Zi|]")
	fs.Parse(args)

	// handle impedance range
	var zClause string
	addZ := func(s string) {
		if len(zClause) > 0 {
			zClause += " and "
		}
		zClause += s
	}
	switch zRange {
	case "any":
		zClause = ""
	case "resonant":
		zClause = "abs(Zi) < 1"
	case "good":
		zClause = "Zr > 30 and Zr < 70 and abs(Zi) < 20"
	case "matched":
		zClause = "Zr > 48 and Zr < 52 and abs(Zi) < 1"
	case "loss":
		zClause = "Zr/sqrt(Zr*Zr+Zi*Zi) > 0.95"
	default:
		zRange = strings.Trim(zRange, "[]")
		parts := strings.Split(zRange, ",")
		if len(parts) != 3 {
			log.Fatal("invalid zRange")
		}
		if len(parts[0]) > 0 {
			if _, err = strconv.ParseFloat(parts[0], 64); err != nil {
				log.Fatal(err)
			}
			addZ("Zr > " + parts[0])
		}
		if len(parts[1]) > 0 {
			if _, err = strconv.ParseFloat(parts[1], 64); err != nil {
				log.Fatal(err)
			}
			addZ("Zr < " + parts[1])
		}
		switch parts[2] {
		case "@":
			addZ("Zr/sqrt(Zr*Zr+Zi*Zi) > 0.95")
		case "!":
			addZ("abs(Zi) < 1")
		default:
			if _, err = strconv.ParseFloat(parts[2], 64); err != nil {
				log.Fatal(err)
			}
			addZ("abs(Zi) < " + parts[2])
		}
	}
	// handle specified frequency (range)
	switch band {
	case "2m":
		spec.Source.Freq = 145000000
	case "70cm":
		spec.Source.Freq = 435000000
	case "35cm":
		spec.Source.Freq = 868000000
	default:
		log.Fatalf("unknown band '%s'", band)
	}

	// target-dependent database query
	var order string
	switch target {
	case "Gmax":
		order = "Gmax desc"
	case "Gmax_u":
		order = "Gmax+10*log10(Zr/sqrt(Zr*Zr+Zi*Zi)) desc"
	case "Gmin":
		order = "Gmax asc"
	case "Gmin_u":
		order = "-Gmax+10*log10(Zr/sqrt(Zr*Zr+Zi*Zi)) desc"
	case "Gmean":
		order = "Gmean desc"
	case "Gmean_u":
		order = "Gmean+10*log10(Zr/sqrt(Zr*Zr+Zi*Zi)) desc"
	case "SD":
		order = "SD asc"
	case "none":
		order = "abs(Zi) asc"
	default:
		log.Fatalf("unknown target '%s'", target)
	}
	// assemble model/geometry list from database
	var geos []string
	rows, err := db.GetRows(zClause, order)
	if err != nil {
		log.Fatal(err)
	}
	for _, r := range rows {
		_, dir, tag := r.Reference()
		if strings.HasPrefix(dir, band) {
			f := in + "/" + dir + "/geometry-" + tag + ".json"
			geos = append(geos, f)
		}
	}

	// setup rendering
	var render lib.Canvas
	if render, err = lib.NewSDLCanvas(1024, 768, 2.01); err != nil {
		log.Fatal(err)
	}
	render.SetHint("Keys: (p)revious, (n)ext")

	var gpos atomic.Uint32
	gpos.Store(0)
	cont := make(chan int)

	go func() {
		for {
			pos := int(gpos.Load())
			path := geos[pos]

			// read geometry file
			body, err := os.ReadFile(path)
			if err != nil {
				log.Fatal(err)
			}
			geo := new(lib.Geometry)
			if err = json.Unmarshal(body, &geo); err != nil {
				log.Fatal(err)
			}
			spec.Wire = geo.Wire

			// build initial geometry
			num := geo.Num
			nodes := make([]lib.Node, num)
			for i := range nodes {
				nodes[i] = lib.NewNode2D(geo.SegL, geo.Bends[i])
			}

			ant := lib.BuildAntenna("geo", spec, nodes)
			if err = ant.Eval(spec.Source.Freq, spec.Wire, spec.Ground); err != nil {
				log.Fatal(err)
			}
			name := strings.TrimPrefix(path, in)
			render.Show(ant, -1, name)
			if rc := <-cont; rc < 0 {
				break
			}
		}
		render.Close()
	}()
	// run render main loop with key-press callback
	render.Run(func(_ *lib.Antenna, key rune, _ int) (rc bool) {
		switch key {
		case 'P':
			if k := gpos.Load(); k > 0 {
				gpos.Store(k - 1)
				rc = true
				cont <- 0
			}
		case 'N', '\n':
			if k := gpos.Load(); int(k) < len(geos)-1 {
				gpos.Store(k + 1)
				rc = true
				cont <- 0
			}
		}
		return
	})
}
