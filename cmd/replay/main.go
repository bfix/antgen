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
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/bfix/antgen/lib"
)

func main() {
	var (
		spec = new(lib.Specification)

		mode   string
		fIn    string
		evalS  string
		outDir string
		err    error
		eval   bool
		render lib.Canvas
	)
	flag.StringVar(&mode, "mode", "track", "operating mode [track,geo]")
	flag.StringVar(&fIn, "in", "", "input file/directory")
	flag.StringVar(&evalS, "eval", "", "evaluate at frequency")
	flag.StringVar(&outDir, "out", "./out", "output directory")
	flag.Parse()

	if len(fIn) == 0 {
		flag.Usage()
		log.Fatal("missing input file/directory")
	}

	// handle specified frequency (range)
	if len(evalS) > 0 {
		if spec.Source.Freq, _, err = lib.GetFrequencyRange(evalS); err != nil {
			log.Fatal(err)
		}
		eval = true
	}

	if mode == "track" {
		// read track file
		body, err := os.ReadFile(fIn)
		if err != nil {
			log.Fatal(err)
		}
		track := new(lib.TrackList)
		if err = json.Unmarshal(body, &track); err != nil {
			log.Fatal(err)
		}
		spec.Wire = track.Wire
		spec.Ground.Height = track.Height

		side := 1.1 * float64(track.Num) * track.SegL

		// setup rendering
		if render, err = lib.NewSDLCanvas(1024, 768, side); err != nil {
			log.Fatal(err)
		}

		// build initial geometry
		num := track.Num
		nodes := make([]*lib.Node, num)
		for i := range nodes {
			nodes[i] = lib.NewNode(track.SegL, 0, 0)
		}

		go func() {
			// iterate over changes
			var ant *lib.Antenna
			init := true
			step := 0
			for _, chg := range track.Track {
				switch chg.Pos {
				case lib.TRK_MARK:
					// marker ends initial geometry build
					init = false

				case lib.TRK_SHORT:
					// shorten leg
					num--
					nodes = nodes[:num]
					continue

				case lib.TRK_LENGTH:
					// lengthen leg
					num++
					nodes = append(nodes, lib.NewNode(track.SegL, 0, 0))
					continue

				default:
					// apply change
					n := nodes[chg.Pos]
					n.AddAngles(chg.Theta, chg.Phi)
				}

				// visualize antenna
				if !init {
					step++
					ant = lib.BuildAntenna("track", spec, nodes)
					if eval {
						if err = ant.Eval(spec.Source.Freq, spec.Wire, spec.Ground); err != nil {
							log.Fatal(err)
						}
					}
					render.Show(ant, chg.Pos, fmt.Sprintf("Step #%d", step))
				}
			}
			render.Show(ant, -1, "final geometry")
			render.Close()
		}()
		render.Run(func(ant *lib.Antenna, key rune, step int) (rc bool) {
			switch key {
			case 'X':
				// write current geometry file
				geo := new(lib.Geometry)
				geo.Nodes = nodes
				data, err := json.MarshalIndent(geo, "", "    ")
				if err != nil {
					log.Fatal(err)
				}
				fName := fmt.Sprintf("%s/geometry-%d.json", outDir, step)
				if err = os.WriteFile(fName, data, 0644); err != nil {
					log.Fatal(err)
				}
			}
			return
		})
	} else if mode == "geo" {
		// setup rendering
		if render, err = lib.NewSDLCanvas(1024, 768, 2.01); err != nil {
			log.Fatal(err)
		}
		render.SetHint("Keys: (p)revious, (n)ext")

		var geos []string
		log.Printf("Scanning directory '%s' for geometry files...", fIn)
		if err = filepath.Walk(fIn, func(path string, info fs.FileInfo, err error) error {
			if info == nil {
				return errors.New("invalid walk")
			}
			if strings.Contains(info.Name(), "geometry-") {
				log.Printf("   Processing '%s'...", path)
				geos = append(geos, path)
			}
			return nil
		}); err != nil {
			log.Fatal(err)
		}
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
				ant := lib.BuildAntenna("geo", spec, geo.Nodes)
				if eval {
					if err = ant.Eval(spec.Source.Freq, spec.Wire, spec.Ground); err != nil {
						log.Fatal(err)
					}
				}
				render.Show(ant, -1, path)
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
}
