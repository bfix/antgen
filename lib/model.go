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
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

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

// Specification of antenna parameters
type Specification struct {
	Wire   Wire   // wire parameters
	Ground Ground // ground parameters
	Source Source // source parameters
}

// Stats return the optimization statistics
type Stats struct {
	NumMthds int
	NumSteps int
	NumSims  int
	Elapsed  time.Duration
}

// Callback when optimization improves
type Callback func(ant *Antenna, pos int, msg string)

// Model of an antenna.
type Model interface {
	// Init model with antenna parameters and generator.
	Init(params string, spec *Specification, k float64, gen Generator) (side float64, err error)

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

// ModelWings returns the number and length of segments for a dipole wing
func ModelWings(k float64, spec *Specification) (num int, segL float64) {
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
	length := 2 * k * lambda
	num = int(length / dx)
	if num%2 == 0 {
		num--
	}
	segL = length / float64(num)
	num = (num - 1) / 2
	return
}

// ParseMdlParams from model file (extract performance parameters)
func ParseMdlParams(fName, dirIn string) (p *Record, ok bool, err error) {
	// open model file
	var fIn *os.File
	if fIn, err = os.Open(fName); err != nil {
		return
	}
	defer fIn.Close()

	// read information
	p = new(Record)
	p.Path = strings.ReplaceAll(filepath.Dir(fName), dirIn+"/", "")

	rdr := bufio.NewReader(fIn)
	var buf []byte
	found := 0
	for {
		if buf, _, err = rdr.ReadLine(); err != nil {
			if err == io.EOF {
				err = nil
				break
			}
			return
		}
		line := string(buf)
		kind, vals := SplitParam(line)
		switch kind {

		// >>>>> Source: freq:Zr:Zi
		case "Source":
			if p.Freq, err = strconv.ParseInt(vals[0], 10, 64); err != nil {
				return
			}
			found++

		// >>>>> Wire: dia:material:conductivity:inductance
		case "Wire":
			if p.Wire.Diameter, err = strconv.ParseFloat(vals[0], 64); err != nil {
				return
			}
			p.Wire.Material = vals[1]
			if p.Wire.Conductivity, err = strconv.ParseFloat(vals[2], 64); err != nil {
				return
			}
			if p.Wire.Inductance, err = strconv.ParseFloat(vals[3], 64); err != nil {
				return
			}
			found++

		// >>>>> Ground: height:mode:type:...
		case "Ground":
			if p.Gnd.Height, err = strconv.ParseFloat(vals[0], 64); err != nil {
				return
			}
			if p.Gnd.Mode, err = strconv.Atoi(vals[1]); err != nil {
				return
			}
			if p.Gnd.Type, err = strconv.Atoi(vals[2]); err != nil {
				return
			}
			found++

		// >>>>> Param: k:param:tag
		case "Param":
			if p.K, err = strconv.ParseFloat(vals[0], 64); err != nil {
				return
			}
			p.Param = math.NaN()
			if len(vals[1]) > 0 {
				if p.Param, err = strconv.ParseFloat(vals[1], 64); err != nil {
					return
				}
			}
			p.Tag = vals[2]
			found++

		// >>>>> Mode: model:generator:seed:optimizer
		case "Mode":
			p.Mdl = vals[0]
			p.Gen = vals[1]
			if p.Seed, err = strconv.ParseInt(vals[2], 10, 64); err != nil {
				return
			}
			p.Opt = vals[3]
			found++

		// >>>>> Result: Gmax:Gmean:SD:Zr:Zi
		case "Result":
			p.Perf.Gain = new(Gain)

			if p.Perf.Gain.Max, err = strconv.ParseFloat(vals[0], 64); err != nil {
				return
			}
			if p.Perf.Gain.Mean, err = strconv.ParseFloat(vals[1], 64); err != nil {
				return
			}
			if p.Perf.Gain.SD, err = strconv.ParseFloat(vals[2], 64); err != nil {
				return
			}
			var Zr, Zi float64
			if Zr, err = strconv.ParseFloat(vals[3], 64); err != nil {
				return
			}
			if Zi, err = strconv.ParseFloat(vals[4], 64); err != nil {
				return
			}
			p.Perf.Z = complex(Zr, Zi)
			found++

		// >>>>> Stats: mthds:steps:sims:elapsed
		case "Stats":
			if p.Stats.NumMthds, err = strconv.Atoi(vals[0]); err != nil {
				return
			}
			if p.Stats.NumSteps, err = strconv.Atoi(vals[1]); err != nil {
				return
			}
			if p.Stats.NumSims, err = strconv.Atoi(vals[2]); err != nil {
				return
			}
			var t int
			if t, err = strconv.Atoi(vals[3]); err != nil {
				return
			}
			p.Stats.Elapsed = time.Duration(t) * time.Second
			found++
		}
	}
	ok = (found > 0)
	return
}

// SplitParam dissects a parameter string
func SplitParam(line string) (kind string, vals []string) {
	if !strings.HasPrefix(line, "CM ") {
		return
	}
	idx := strings.IndexRune(line, ':')
	if idx == -1 {
		return
	}
	kind = line[3:idx]
	vals = strings.Split(line[idx+2:], ":")
	return
}
