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
	"fmt"
	"log"
	"os"

	"github.com/bfix/antgen/lib"
)

// convert antgen geometries to other formats
func main() {
	var (
		spec = new(lib.Specification)

		fGeo  string  // name of geometry file
		mode  string  // conversion mode
		fOut  string  // output file/directory
		freqS string  // frequency range
		v     float64 // velocity factor
	)
	// handle command-line arguments
	flag.StringVar(&mode, "mode", "svg", "conversion mode [svg]")
	flag.StringVar(&fGeo, "in", "", "geometry input")
	flag.StringVar(&freqS, "freq", "", "operating frequency")
	flag.Float64Var(&v, "v", 1.0, "velocity factor")
	flag.StringVar(&fOut, "out", "", "output")
	flag.Parse()

	// check mandatory args
	if len(fGeo) == 0 {
		flag.Usage()
		log.Fatal("missing geometry filename")
	}

	// handle specified frequency (range)
	var err error
	if len(freqS) > 0 {
		if spec.Source.Freq, _, err = lib.GetFrequencyRange(freqS); err != nil {
			log.Fatal(err)
		}
	}

	// read geometry file
	var body []byte
	if body, err = os.ReadFile(fGeo); err != nil {
		log.Fatal(err)
	}
	geo := new(lib.Geometry)
	if err = json.Unmarshal(body, &geo); err != nil {
		log.Fatal(err)
	}
	spec.Wire = geo.Wire

	// handle conversion
	switch mode {
	case "svg":
		err = convert2SVG(fGeo, fOut, geo, spec, v)
	default:
		err = fmt.Errorf("unknown conversion '%s'", mode)
	}
	if err != nil {
		log.Fatal(err)
	}
}
