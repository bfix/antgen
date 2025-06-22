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

func TestNumbers(t *testing.T) {
	EPS := 1e-5
	for i := 0; i < 100; i++ {
		v := math.Round((rand.Float64() * 100000))
		e := rand.Intn(19) - 9
		k := math.Pow10(e)
		s := float64(2*(rand.Int()%2) - 1)
		f := s * v * k

		sf := FormatNumber(f, 5)
		ft, err := ParseNumber(sf)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%e -- %s -- %e", f, sf, ft)
		if d := math.Abs(ft-f) / f; d > EPS {
			t.Errorf("failed: %f", d)
		}
	}
}

func TestComplex(t *testing.T) {
	s := []string{
		"10", "23+j42", "-35.4-6.8*i",
	}
	for _, x := range s {
		if k, err := ParseImpedance(x); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("%s -- %v", x, k)
		}
	}
}
