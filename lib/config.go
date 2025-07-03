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
	"os"
)

// Default values (command-line options)
type Default struct {
	K      float64 `json:"k"`      // default 'k'(wing in wavelength)
	Wire   Wire    `json:"wire"`   // default wire parameters
	Ground Ground  `json:"ground"` // ground parameters
	Source Source  `json:"source"` // source parameters
}

// Simulation parameters
type Simulation struct {
	// optimization parameters (termination conditions)
	MaxRounds     int     `json:"maxRounds"`     // max. number of rounds in optimization
	MinZr         float64 `json:"minZr"`         // min. resistance of antenna
	MaxZr         float64 `json:"maxZr"`         // max. resistance of antenna
	MinChange     float64 `json:"minChange"`     // progress check: min. change in target value
	ProgressCheck int     `json:"progressCheck"` // number of steps between progress check
	MinBend       float64 `json:"minBend"`       // min. bending angle (fraction of max. angle)

	// simulation-related constants (NEC2 simulation)
	ExciteU   float64 `json:"exciteU"`   // excitation voltage
	PhiStep   float64 `json:"phiStep"`   // azimut step (degree)
	ThetaStep float64 `json:"thetaStep"` // elevation step (degree)

	// geometry-related constraints (NEC2 simulation)
	WireMax      float64 `json:"wireMax"`      // max. wire diameter (in wavelength)
	SegMinLambda float64 `json:"segMinLambda"` // min. segment length (in wavelength)
	SegMinWire   float64 `json:"segMinWire"`   // min. segment length (in wire diameter)
	MinRadius    float64 `json:"minRadius"`    // min. curve radius (in wavelength)
}

// Material spec for wires
type Material struct {
	Conductivity float64 `json:"conductivity"` // wire conductivity (S/m)
	Inductance   float64 `json:"inductance"`   // wire inductance (H/m)
}

// RenderConfig for rendering-related settings
type RenderConfig struct {
	Canvas string `json:"canvas"` // render engine/canvas
	Width  int    `json:"width"`  // width of canvas (usually in pixels)
	Height int    `json:"height"` // height of canvas (usually in pixels)
}

// Config for AntGen
type Config struct {
	Def     *Default             `json:"default"`
	Sim     *Simulation          `json:"simulation"`
	Mat     map[string]*Material `json:"material"`
	Render  *RenderConfig        `json:"render"`
	Plugins map[string]string    `json:"plugins"`
}

// Cfg is the globally-accessible configuration (pre-set)
var Cfg = &Config{
	// default values (command-line options)
	Def: &Default{
		K: 0.25,
		Wire: Wire{
			Diameter:     0.002,
			Material:     "CuL",
			Conductivity: 5.96e7,
			Inductance:   1.54e-7,
		},
		Ground: Ground{
			Height: 0,
			Mode:   0,
			Type:   -1,
			NRadl:  0,
			Epse:   0,
			Sig:    0,
		},
		Source: Source{
			Z:     Impedance{50, 0},
			Power: 1,
			Freq:  435000000,
			Span:  5000000,
		},
	},
	// Simulation parameters
	Sim: &Simulation{
		// optimization parameters (termination conditions)
		MaxRounds:     5,
		MinZr:         1,
		MaxZr:         20000,
		MinChange:     0.001,
		ProgressCheck: 10,
		MinBend:       0.01,

		// simulation-related constants (NEC2 simulation)
		ExciteU:   1.0,
		PhiStep:   5.0,
		ThetaStep: 5.0,

		// geometry-related constraints (NEC2 simulation)
		WireMax:      0.008,
		SegMinLambda: 0.002,
		SegMinWire:   4,
		MinRadius:    0.02,
	},
	// rendering parameters
	Render: &RenderConfig{
		Canvas: "sdl",
		Width:  1024,
		Height: 768,
	},
	// wire materials
	Mat: map[string]*Material{
		"Cu": { // cupper wire
			5.96e7,      // conductivity (S/m)
			1.320172e-6, // inductance (H/m)
		},
		"CuL": { // emailled copper wire
			5.96e7,  // conductivity (S/m)
			1.54e-7, // inductance (H/m)
		},
		"Al": { // Aluminium
			3.5e7,      // conductivity (S/m)
			1.32021e-6, // inductance (H/m)
		},
	},
	// no pre-defined plugins
	Plugins: make(map[string]string),
}

// ReadConfig from file
func ReadConfig(fname string) (err error) {
	var data []byte
	if data, err = os.ReadFile(fname); err == nil {
		err = json.Unmarshal(data, &Cfg)
	}
	return
}
