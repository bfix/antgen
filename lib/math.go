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

	"gonum.org/v1/gonum/mat"
)

//----------------------------------------------------------------------

// Global settings and defaults
const (
	eps = 1e-9 // lower bound for non-zero

	// mathematical constants
	RectAng = math.Pi / 2 // right angle
	CircAng = 2 * math.Pi //full circle
)

// IsNull returns true if number is zero (within tolerance)
func IsNull(f float64) bool {
	return math.Abs(f) < eps
}

// InRange returns true if value v is in range (with tolerance)
func InRange(v, from, to float64) bool {
	return v-from > -eps && to-v > -eps
}

// Sqr returns the square of a value
func Sqr(v float64) float64 {
	return v * v
}

// ----------------------------------------------------------------------

// BestFitSphere returns the radius and center point of a sphere
// that bests fits the given points (least square fit).
func BestFitSphere(pnts []Vec3) (r float64, ctr Vec3, err float64) {
	num := len(pnts)
	aVal := make([]float64, 4*num)
	fVal := make([]float64, num)
	for i, pt := range pnts {
		for j := range 3 {
			aVal[4*i+j] = pt[j] * 2
		}
		aVal[4*i+3] = 1
		fVal[i] = Sqr(pt[0]) + Sqr(pt[1]) + Sqr(pt[2])
	}
	A := mat.NewDense(num, 4, aVal)
	f := mat.NewVecDense(num, fVal)

	var x mat.VecDense
	x.SolveVec(A, f)

	ctr = NewVec3(x.At(0, 0), x.At(1, 0), x.At(2, 0))
	r = math.Sqrt(x.At(3, 0) + Sqr(ctr.Length()))

	// sum squared error
	for _, pt := range pnts {
		err += Sqr(pt.Sub(ctr).Length() - r)
	}
	return
}
