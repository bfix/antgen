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
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// Callback when optimization improves
type Callback func(ant *Antenna, pos int, msg string)

// Model of an antenna.
type Model interface {
	// Init model with antenna parameters and generator.
	Init(params string, spec *Specification, gen Generator) (side float64, err error)

	// Prepare initial geometry
	Prepare(seed int64, cb Callback) (ant *Antenna, err error)

	// Optimize antenna geometry based on random seed and comparator
	// (to evaluate progress during optimization)
	Optimize(seed int64, iter int, cmp *Comparator, cb Callback) (ant *Antenna, stats Stats, err error)

	// Info about the model (parameters)
	Info() string

	// Finalize model after optimization (write track and geometry files).
	Finalize(tag, outDir, outPrf string, cmts []string)
}

//----------------------------------------------------------------------

// Model of a dipole antenna with symmetrical legs. Each leg is made
// from segments of equal length.
type ModelDipole struct {
	Kind string         // kind of model
	Spec *Specification // antenna specs

	Nodes []*Node // list of segments
	Num   int     // number of segments
	SegL  float64 // segment length

	Track []*Change // list of changes
}

// Init base model
func (mdl *ModelDipole) Init(params string, spec *Specification, gen Generator) (side float64, err error) {
	mdl.Spec = spec

	// check if wire diameter works for wavelength
	// NEC2: wire << lambda / 2π
	lambda := spec.Source.Lambda()
	if a := Cfg.Sim.WireMax * lambda; spec.Wire.Diameter > a {
		spec.Wire.Diameter = a
	}
	// compute segment length with lower bound
	// NEC2: dx > segMin*lambda, dx > SegRatio*wire
	dx := max(Cfg.Sim.SegMinLambda*lambda, Cfg.Sim.SegMinWire*spec.Wire.Diameter)

	// init model parameters
	span := 2*spec.K*lambda - spec.Feedpt.Gap
	num := int(span / dx)
	if IsNull(spec.Feedpt.Gap) {
		if num%2 == 0 {
			num++
		}
		spec.Feedpt.Gap = span / float64(num)
	} else {
		if num%2 == 1 {
			num++
		}
	}
	mdl.SegL = span / float64(num)
	mdl.Num = num / 2
	side = float64(mdl.Num) * mdl.SegL
	mdl.Kind = fmt.Sprintf("%.3f λ dipole", 2*spec.K)
	return
}

// Finalize model (write track and geometry files)
func (mdl *ModelDipole) Finalize(tag, outDir, outPrf string, cmts []string) {
	if len(mdl.Track) > 0 {
		// write track file
		o := new(TrackList)
		o.SegL = mdl.SegL
		o.Num = mdl.Num
		o.Track = mdl.Track
		o.Wire = mdl.Spec.Wire
		o.Height = mdl.Spec.Ground.Height
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
	geo := new(Geometry)
	geo.Cmts = cmts
	geo.Wire = mdl.Spec.Wire
	geo.Feedpt = mdl.Spec.Feedpt
	geo.Height = mdl.Spec.Ground.Height
	geo.Nodes = mdl.Nodes
	data, err := json.MarshalIndent(geo, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	fName := fmt.Sprintf("%s/%sgeometry-%s.json", outDir, outPrf, tag)
	if err = os.WriteFile(fName, data, 0644); err != nil {
		log.Fatal(err)
	}
}
