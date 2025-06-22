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
	"math"
	"math/rand"
	"testing"
)

// Test the BestFitSphere function
func TestBestFitSphere(t *testing.T) {
	// Generate random points on sphere
	rnd := rand.New(rand.NewSource(0x19031962))
	pnts := make([]Vec3, 1000)
	rad := 100.
	ctr := NewVec3(23, 42, 67)
	for i := range pnts {
		elev := (rnd.Float64() - 0.5) * math.Pi
		azim := rnd.Float64() * CircAng
		pnts[i] = NewVec3(
			rad*math.Cos(azim)*math.Sin(elev),
			rad*math.Sin(azim)*math.Sin(elev),
			rad*math.Cos(elev),
		).Add(ctr)
	}
	// compute best fit
	r, c, f := BestFitSphere(pnts)

	// check results
	if !IsNull(r - rad) {
		t.Errorf("rad: %f != %f", rad, r)
	}
	if !IsNull(c.Sub(ctr).Length()) {
		t.Errorf("ctr: %v != %v", ctr, c)
	}
	if !IsNull(f) {
		t.Errorf("err: %f", f)
	}
}
