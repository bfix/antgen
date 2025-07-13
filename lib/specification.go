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
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

// Specification of antenna parameters
type Specification struct {
	K      float64 `json:"k"`      // leg in wavelength
	Wire   Wire    `json:"wire"`   // wire parameters
	Ground Ground  `json:"ground"` // ground parameters
	Source Source  `json:"source"` // source parameters
	Feedpt Feedpt  `json:"feedpt"` // feed point parameters
}

// Stats return the optimization statistics
type Stats struct {
	NumMthds int
	NumSteps int
	NumSims  int
	Elapsed  time.Duration
}

//----------------------------------------------------------------------

// Wire parameters
type Wire struct {
	Diameter     float64 `json:"dia"`      // wire diameter
	Material     string  `json:"material"` // wire material
	Conductivity float64 `json:"G"`        // wire conductivity (S/m)
	Inductance   float64 `json:"L"`        // wire inductivity (H/m)
}

// ParseWire converts a specification string into a Wire
func ParseWire(wireS string, warn bool) (w Wire, err error) {
	w = Cfg.Def.Wire
	if len(wireS) == 0 {
		if warn {
			log.Printf("no wire parameters defined - using defaults.")
		}
		return
	}
	parts := strings.Split(wireS, ":")
	if len(parts[0]) > 0 {
		if w.Diameter, err = strconv.ParseFloat(parts[0], 64); err != nil {
			return
		}
	}
	if len(parts) > 1 && len(parts[1]) > 0 {
		if parts[1][0] == '&' {
			w.Material = parts[1][1:]
			w.Conductivity, w.Inductance, err = MaterialProperties(parts[1][1:], w.Diameter)
			return
		} else if w.Conductivity, err = strconv.ParseFloat(parts[1], 64); err != nil {
			return
		}
	}
	if len(parts) > 2 && len(parts[2]) > 0 {
		if w.Inductance, err = strconv.ParseFloat(parts[2], 64); err != nil {
			return
		}
	}
	return
}

//----------------------------------------------------------------------

// Ground parameters
type Ground struct {
	Height float64 `json:"height"` // height of antenna above ground
	Mode   int     `json:"mode"`   // ground mode (0=no ground, 1=sym ground, -1=no-sym ground)
	Type   int     `json:"type"`   // NEC2 ground type (-1: free space, 0: finite, 1:conductive, 2: finite(SN))
	NRadl  int     `json:"nradl"`  // number of radial wires in the ground screen
	Epse   float64 `json:"epse"`   // relative dielectric constant for ground in the vicinity of the antenna
	Sig    float64 `json:"sig"`    // conductivity in mhos/meter of the ground in the vicinity of the antenna
}

// ParseGround converts a ground spec into Ground
func ParseGround(groundS string, warn bool) (gnd Ground, err error) {
	gnd = Cfg.Def.Ground
	if len(groundS) == 0 {
		if warn {
			log.Printf("no ground parameters defined - using defaults.")
		}
		return
	}
	var i int64
	for _, p := range strings.Split(groundS, ",") {
		fp := strings.SplitN(p, "=", 2)
		switch fp[0] {
		case "height":
			if len(fp) != 2 {
				log.Fatal("ground: missing height value")
			}
			if gnd.Height, err = strconv.ParseFloat(fp[1], 64); err != nil {
				return
			}
		case "mode":
			if len(fp) != 2 {
				log.Fatal("ground: missing mode value")
			}
			if i, err = strconv.ParseInt(fp[1], 10, 64); err != nil {
				return
			}
			gnd.Mode = int(i)
		case "type":
			if len(fp) != 2 {
				log.Fatal("ground: missing type value")
			}
			if i, err = strconv.ParseInt(fp[1], 10, 64); err != nil {
				return
			}
			gnd.Type = int(i)
		case "nradl":
			if len(fp) != 2 {
				log.Fatal("ground: missing nradl value")
			}
			if i, err = strconv.ParseInt(fp[1], 10, 64); err != nil {
				return
			}
			gnd.NRadl = int(i)
		case "epse":
			if len(fp) != 2 {
				log.Fatal("ground: missing epse value")
			}
			if gnd.Epse, err = strconv.ParseFloat(fp[1], 64); err != nil {
				return
			}
		case "sig":
			if len(fp) != 2 {
				log.Fatal("ground: missing sig value")
			}
			if gnd.Sig, err = strconv.ParseFloat(fp[1], 64); err != nil {
				return
			}
		default:
			err = fmt.Errorf("unknown ground parameter '%s'", fp[0])
			return
		}
	}
	// sanity check
	if !IsNull(gnd.Height) && gnd.Mode == 0 {
		err = errors.New("ground: height set, but no ground mode defined")
	}
	if IsNull(gnd.Height) && gnd.Mode != 0 {
		err = errors.New("ground: height not set, but ground mode defined")
	}
	return
}

//----------------------------------------------------------------------

// Impedance (complex)
type Impedance struct {
	R float64 `json:"R"` // resistance
	X float64 `json:"X"` // reactance
}

// Source parameters
type Source struct {
	Z     Impedance `json:"Z"`     // source impedance
	Power float64   `json:"power"` // source power
	Freq  int64     `json:"freq"`  // frequency
	Span  int64     `json:"span"`  // freq span
}

// Impedance of source
func (src Source) Impedance() complex128 {
	return complex(src.Z.R, src.Z.X)
}

// Lambda (wavelength) of source frequency
func (src Source) Lambda() float64 {
	return C / float64(src.Freq)
}

// ParseSource converts a source spec into Source
func ParseSource(sourceS string, warn bool) (src Source, err error) {
	src = Cfg.Def.Source
	if len(sourceS) == 0 {
		if warn {
			log.Printf("no source parameters defined - using defaults.")
		}
		return
	}
	for _, p := range strings.Split(sourceS, ",") {
		fp := strings.SplitN(p, "=", 2)
		switch fp[0] {
		case "Z":
			if len(fp) != 2 {
				err = errors.New("source: missing Z value")
				return
			}
			var Z complex128
			if Z, err = ParseImpedance(fp[1]); err != nil {
				return
			}
			src.Z.R, src.Z.X = real(Z), imag(Z)
		case "Pwr":
			if len(fp) != 2 {
				log.Fatal("source: missing Power value")
			}
			if src.Power, err = ParseNumber(fp[1]); err != nil {
				return
			}
		}
	}
	return
}

//----------------------------------------------------------------------

type Feedpt struct {
	Gap       float64 `json:"gap"`       // distance between legs at feed point
	Extension float64 `json:"extension"` // extension of wire away from feedpt
}

// ParseFeedpt converts a feedpoint spec
func ParseFeedpt(feedptS string, warn bool) (fpt Feedpt, err error) {
	fpt = Cfg.Def.Feedpt
	if len(feedptS) == 0 {
		if warn {
			log.Printf("no feedpoint parameters defined - using defaults.")
		}
		return
	}
	for _, p := range strings.Split(feedptS, ",") {
		fp := strings.SplitN(p, "=", 2)
		switch fp[0] {
		case "gap":
			if len(fp) != 2 {
				err = errors.New("feedpt: missing gap value")
				return
			}
			if fpt.Gap, err = ParseNumber(fp[1]); err != nil {
				return
			}
		case "ext":
			if len(fp) != 2 {
				log.Fatal("feedpt: missing extension value")
			}
			if fpt.Extension, err = ParseNumber(fp[1]); err != nil {
				return
			}
		}
	}
	return
}
