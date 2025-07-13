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
	"fmt"
	"log"
	"math"
	"math/cmplx"
	"plugin"
	"strings"
)

// Gain of antenna
type Gain struct {
	Max  float64 // maximum gain
	Mean float64 // mean gain
	SD   float64 // standard deviation of mean
}

// Performance of antenna
type Performance struct {
	Gain *Gain       // antenna gain
	Z    complex128  // antenna impedance
	Rp   *RadPattern // radiation pattern
}

// String returns a human-readable performance text
func (p *Performance) String() string {
	if p.Gain == nil {
		return ""
	}
	return fmt.Sprintf("Gain={Max: %.5f dB, Mean: %.5f±%.5f dB}, Impedance=%s Ω",
		p.Gain.Max, p.Gain.Mean, p.Gain.SD, FormatImpedance(p.Z, 5))
}

// SWR for (unmatched) antenna at source impedance
func (p *Performance) SWR(Zs complex128) float64 {
	g := cmplx.Abs((p.Z - Zs) / (p.Z + Zs))
	return (1 + g) / (1 - g)
}

// Loss (in dB) of transfering power from a source with impedance Zs to an
// unmatched antenna with impedance r.Z
func (p *Performance) Loss(Zs complex128) float64 {
	s := p.SWR(Zs)
	return 10 * math.Log10(4*s/Sqr(s+1))
}

// Power factor (in dB) of a matched antenna.
func (p *Performance) Attenuation(Zs complex128) float64 {
	// power factor (depends on phase shift between U and I)
	pf := real(p.Z) / cmplx.Abs(p.Z) // math.Cos(cmplx.Phase(r.Z))
	return 10 * math.Log10(pf)
}

// Resonance is the "virtual loss" (in dB) due to antenna reactance.
// No loss implies resonance.
func (p *Performance) Resonance() float64 {
	// r = 1 / (1 + Zi^2)
	return math.Log10(1 / (1 + imag(p.Z)*imag(p.Z)))
}

//----------------------------------------------------------------------

// Evaluate performance (metric value optimized to maximum)
type Evaluate func(perf *Performance, args string, feedZ complex128) float64

// CustomEvaluators is a list of custom comparator implementations
var CustomEvaluators = make(map[string]Evaluate)

// Comparator creates a standard metric for antenna results.
// It is used in the optimization loop to find improvements towards a goal.
// The optimization algorithms interprets higher values as "better" values.
type Comparator struct {
	targets []string
	args    map[string]string
	eval    []Evaluate
	pos     int
	spec    *Specification
}

// Create a new comparator for a target (and a possible target value).
// Known targets are:
// * Gmax: highest gain
// * Gmean: best mean gain
// * SD: smallest standard deviation
// * custom: custom comparator (possibly plugin)
func NewComparator(target string, spec *Specification) (cmp *Comparator, err error) {
	cmp = new(Comparator)
	cmp.targets = make([]string, 0)
	cmp.args = make(map[string]string)
	cmp.eval = make([]Evaluate, 0)

	for _, tgt := range strings.Split(target, ",") {
		parts := strings.SplitN(tgt, "=", 2)
		cmp.targets = append(cmp.targets, parts[0])

		// check for custom evaluator
		eval, ok := CustomEvaluators[parts[0]]
		var args string
		if !ok {
			// not a custom eval; check for plugin or LUA script
			ref := strings.SplitN(parts[0], ":", 2)
			switch ref[0] {
			case "plugin":
				if len(ref) < 2 {
					log.Fatal("incomplete plugin specification")
				}
				var pi *plugin.Plugin
				if pi, err = GetPlugin(ref[1]); err != nil {
					log.Fatal(err)
				}
				if eval, err = GetSymbol[Evaluate](pi, "Evaluate"); err != nil {
					log.Fatal(err)
				}
				if len(parts) > 1 {
					args = parts[1]
				}
			case "lua":
				if len(ref) < 2 {
					log.Fatal("incomplete LUA script specification")
				}
				ev, err := NewLuaEvaluator(ref[1])
				if err != nil {
					log.Fatal(err)
				}
				eval = ev.Evaluate
				if len(parts) > 2 {
					args = parts[1]
				}
			default:
				// standard evaluator
				if len(parts) > 1 {
					args = parts[1]
				}
				eval = cmp.value
			}
		} else {
			if len(parts) > 1 {
				args = parts[1]
			}

		}
		cmp.eval = append(cmp.eval, eval)
		cmp.args[parts[0]] = args
	}
	cmp.pos = 0
	cmp.spec = spec
	return
}

// Value returns the evaluated value from perfomance data.
func (cmp *Comparator) Value(p *Performance) float64 {
	target := cmp.targets[cmp.pos]
	args := cmp.args[target]
	return cmp.eval[cmp.pos](p, args, cmp.spec.Source.Impedance())
}

// standard evaluation
func (cmp *Comparator) value(p *Performance, args string, feedZ complex128) (val float64) {
	switch cmp.targets[cmp.pos] {
	case "Gmax":
		// opt for best directional pattern
		if len(args) == 0 || args == "raw" {
			val = p.Gain.Max
		} else if args == "unmatched" {
			val = p.Loss(feedZ) + p.Gain.Max
		} else if args == "matched" {
			val = p.Attenuation(feedZ) + p.Gain.Max
		} else if args == "resonant" {
			val = p.Resonance() + p.Gain.Max
		} else {
			log.Fatalf("invalid argument '%s' for 'Gmax'", args)
		}
	case "Gmean":
		// opt for best quasi-isotrope pattern
		if len(args) == 0 || args == "raw" {
			val = p.Gain.Mean
		} else if args == "unmatched" {
			val = p.Loss(feedZ) + p.Gain.Mean
		} else if args == "matched" {
			val = p.Attenuation(feedZ) + p.Gain.Mean
		} else if args == "resonant" {
			val = p.Resonance() + p.Gain.Mean
		} else {
			log.Fatalf("invalid argument '%s' for 'Gmean'", args)
		}
	case "SD":
		// opt for smaller SD
		val = -p.Gain.SD
	case "Z":
		// opt for matching impedance
		val = p.Loss(feedZ)
	case "none":
		val = 0
	default:
		log.Fatalf("unknown optimization target '%s'", cmp.targets[cmp.pos])
	}
	return
}

// Compare antenna results based on the optimization target.
// Returns 0 if same, -1 if worse, 1 if better
func (cmp *Comparator) Compare(curr, old *Performance) (sign int, val float64) {
	// execute comparator
	eps := 1e-9
	val = cmp.Value(curr)
	chg := val - cmp.Value(old)

	// calculate improvement
	sign = 0
	if chg > eps {
		sign = 1
	} else if chg < -eps {
		sign = -1
	}
	return
}

// Target returns the current optimization target
func (cmp *Comparator) Target() string {
	return fmt.Sprintf("%s (%d/%d)", cmp.targets[cmp.pos], cmp.pos+1, len(cmp.targets))
}

// Next optimization target
func (cmp *Comparator) Next() (ok bool) {
	if ok = (cmp.pos < len(cmp.targets)-1); ok {
		cmp.pos++
	}
	return
}

//----------------------------------------------------------------------

// RadPattern is the radiation pattern of an antenna
type RadPattern struct {
	NPhi, NTheta int
	Min, Max     float64
	Values       [][]float64
}

// Spherical is a metric for the isotropicity of a radition pattern.
// Values are positive; smaller numbers are "better". A value is
// calculated as ∑error(i)²/n over all points (with i = 1..n).
func (rp *RadPattern) Spherical() (f float64) {
	// build list of 3D points
	// Θ (Theta): angle measured between the positive Z semiaxis and the
	//            ground plane XY (elevation angle: π/2 - Θ)
	// Φ (Phi):   angle measured between the positive X semiaxis and the
	//            YZ plane (azimuth = π/2 - Φ)

	dTheta := CircAng / float64(rp.NTheta)
	dPhi := CircAng / float64(2*rp.NPhi)
	pnts := make([]Vec3, 0, rp.NPhi*rp.NTheta)
	for iTheta, row := range rp.Values {
		elev := RectAng - float64(iTheta)*dTheta
		for iPhi, val := range row {
			azim := RectAng - float64(iPhi)*dPhi
			pt := NewVec3(
				val*math.Cos(azim)*math.Sin(elev),
				val*math.Sin(azim)*math.Sin(elev),
				val*math.Cos(elev),
			)
			pnts = append(pnts, pt)
		}
	}
	// least-square fitted sphere
	_, _, f = BestFitSphere(pnts)
	// mean squared error
	f /= float64(len(pnts))
	return
}
