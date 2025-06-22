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
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"

	"github.com/bfix/antgen/lib"
)

// List of all available models
var mdls = make(map[string]func(int) (lib.Model, error))

// GetModel by name
func GetModel(name string, spec *lib.Specification, k float64, gen lib.Generator, verbose int) (mdl lib.Model, side float64, err error) {
	s := strings.SplitN(name, ":", 2)
	mdlF, ok := mdls[s[0]]
	if !ok {
		err = fmt.Errorf("no such model '%s'", name)
		return
	}
	if mdl, err = mdlF(verbose); err != nil {
		return
	}
	params := ""
	if len(s) > 1 {
		params = s[1]
	}
	side, err = mdl.Init(params, spec, k, gen)
	return
}

//----------------------------------------------------------------------

// Model of a dipole antenna with symmetrical wings. Each wings is made
// from segments of equal length.
type ModelDipole struct {
	kind string             // kind of model
	spec *lib.Specification // antenna specs

	nodes []lib.Node // list of segments
	size  int        // initial number of segments
	num   int        // current number of segments
	segL  float64    // segment length

	rnd   *rand.Rand    // randomizer
	track []*lib.Change // list of changes
}

// Init base model
func (mdl *ModelDipole) Init(params string, spec *lib.Specification, k float64, gen lib.Generator) (side float64, err error) {
	mdl.spec = spec
	num, segL := lib.ModelWings(k, spec)
	side = float64(num) * segL
	mdl.size = num
	mdl.num = mdl.size
	mdl.segL = segL
	mdl.kind = fmt.Sprintf("%.3f λ dipole", 2*k)
	return
}

// Finalize model (write track and geometry files)
func (mdl *ModelDipole) Finalize(tag, outDir, outPrf string, cmts []string) {
	if len(mdl.track) > 0 {
		// write track file
		o := new(lib.TrackList)
		o.SegL = mdl.segL
		o.Num = mdl.size
		o.Track = mdl.track
		o.Wire = mdl.spec.Wire
		o.Height = mdl.spec.Ground.Height
		o.Cmts = cmts

		data, err := json.MarshalIndent(o, "", "    ")
		if err != nil {
			log.Fatal(err)
		}
		fName := fmt.Sprintf("%s/%strack-%s.json", outDir, outPrf, tag)
		if err = os.WriteFile(fName, data, 0644); err != nil {
			log.Fatal(err)
		}
	}
	// write current geometry file
	geo := new(lib.Geometry)
	geo.Num = mdl.num
	geo.SegL = mdl.segL
	geo.Cmts = cmts
	geo.Wire = mdl.spec.Wire
	geo.Height = mdl.spec.Ground.Height
	geo.Bends = make([]float64, mdl.num)
	for i, n := range mdl.nodes {
		_, angle := n.Polar()
		geo.Bends[i] = angle
	}
	data, err := json.MarshalIndent(geo, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	fName := fmt.Sprintf("%s/%sgeometry-%s.json", outDir, outPrf, tag)
	if err = os.WriteFile(fName, data, 0644); err != nil {
		log.Fatal(err)
	}
}
