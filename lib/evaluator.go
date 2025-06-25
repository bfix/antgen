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
	"log"
	"math"
)

// add custom comparators
func init() {
	CustomEvaluators["isotrope"] = IsotropeEvaluate
	CustomEvaluators["Gmin"] = GminEvaluate
}

// IsotropeEvaluate implements the Compare prototype
// It returns a value representing how spherical the radiation pattern is.
func IsotropeEvaluate(p *Performance, args string, feedZ complex128) (val float64) {
	// metric value is log(∑error² + 1)
	val = -10 * math.Log10(p.Rp.Spherical()+1)

	// handle argument
	if args == "unmatched" {
		val += p.Loss(feedZ)
	} else if args == "matched" {
		val += p.Attenuation(feedZ)
	} else if args == "resonant" {
		val += p.Resonance()
	} else if len(args) > 0 {
		log.Fatalf("invalid argument '%s' for 'isotrope'", args)
	}
	return
}

// Gmin evaluator (minimizing Gmax)
func GminEvaluate(p *Performance, args string, feedZ complex128) (val float64) {
	val = -p.Gain.Max

	// handle argument
	if args == "unmatched" {
		val += p.Loss(feedZ)
	} else if args == "matched" {
		val += p.Attenuation(feedZ)
	} else if args == "resonant" {
		val += p.Resonance()
	} else if len(args) > 0 {
		log.Fatalf("invalid argument '%s' for 'Gmin'", args)
	}
	return
}
