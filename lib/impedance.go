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
	"math/cmplx"
)

// Matcher between impedances.
// The shunt element (Cp/Lp) is located at the side with higher impedance
// (load if swap=false or source if swap=true). The matcher is either a
// low-pass (Cp/Ls) or a high-pass (Cs/Lp) filter.
type Matcher struct {
	AtSource bool // placement of shunt element
	xp, xr   float64
}

// HighPass element values at given frequency
func (m *Matcher) HighPass(freq float64) (Lp, Cs float64) {
	w := 2 * math.Pi * freq
	Cs, Lp = 1/(w*m.xr), m.xp/w
	return
}

// LowPass element values at given frequency
func (m *Matcher) LowPass(freq float64) (Cp, Ls float64) {
	w := 2 * math.Pi * freq
	Cp, Ls = 1/(w*m.xp), m.xr/w
	return
}

// Zmatch the source impedance Zs to the load impedance Zl.
// Z_L: load impedance (R_L + X_L*j)
// Z_P: Reactance parallel to load (X_P*j)
// [maxima-start]
//
//	    Zl: Rl+Xl*%i;
//	    Zp: Xp*%i;
//	    Z:  rectform((Zl*Zp)/(Zl+Zp));
//	    R:  expand(realpart(Z));
//	-->     (Rl*Xp^2) / ((Xp+Xl)^2 + Rl^2) == Rs
//	    solve(R-Rs,Xp);
//	-->     Xp = (sqrt(Rl*Rs*Xl^2 - Rl^2*Rs^2 + Rl^3*Rs) + Rs*Xl) / (Rl-Rs)
//	    Xr: imagpart(Z);
//
// [maxima-end]
func Zmatch(Zs, Zl complex128) (Z complex128, m *Matcher) {
	m = new(Matcher)

	// swap source and load if Zl < Zs
	if cmplx.Abs(Zs) > cmplx.Abs(Zl) {
		Zs, Zl = Zl, Zs
		m.AtSource = true
	}

	Rs, Xs := real(Zs), imag(Zs)
	Rl, Xl := real(Zl), imag(Zl)

	m.xp = (math.Sqrt(Rl*Rs*Xl*Xl-Rl*Rl*Rs*Rs+Rl*Rl*Rl*Rs) + Rs*Xl) / (Rl - Rs)
	m.xr = m.xp*(Rl*Rl+Xl*m.xp+Xl*Xl)/(Rl*Rl+(m.xp+Xl)*(m.xp+Xl)) - Xs

	Zp := complex(0, m.xp)
	Z = (Zl * Zp) / (Zl + Zp)
	return
}

// ToReflection computes the complex reflection factor between Z and Z0.
// The value is within a unit circle in the complex plane (Smith chart).
func ToReflection(z, z0 complex128) complex128 {
	return (z - z0) / (z + z0)
}

// FromReflection computes the impedance Z if a reference impedance Z0 and
// a complex reflection (Smith chart coordinate) are given.
func FromReflection(g, z0 complex128) complex128 {
	k := (1 + g) / (1 - g)
	return k * z0
}
