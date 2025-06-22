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
	"testing"
)

func TestLoss(t *testing.T) {
	Zs := complex(50, 0)
	r := new(Performance)
	for _, Zl := range []complex128{
		complex(27, 0),
		complex(108, 74),
		complex(71, 28),
		complex(5, 0),
	} {
		r.Z = Zl
		f := r.Loss(Zs)
		t.Logf("Zl=%s, f=%f\n", FormatImpedance(Zl, 5), f)
	}
}

func TestSWR(t *testing.T) {
	Zs := complex(50, 0)
	r := new(Performance)
	for _, Zl := range []complex128{
		7.664 - 569.8i,
		40.684128 - 172.193310i,
		43.725893 - 152.978284i,
		46.851551 - 133.481797i,
		50.294661 - 114.383460i,
		53.856060 - 95.046731i,
		57.764091 - 75.926341i,
		61.832717 - 56.592882i,
		66.282562 - 37.317784i,
		70.945204 - 17.839074i,
		76.030469 + 1.720735i,
	} {
		r.Z = Zl
		f := r.SWR(Zs)
		t.Logf("Zl=%s, %f\n", FormatImpedance(Zl, 5), f)
	}
}

func TestEval(t *testing.T) {

	Zs := complex(50, 0)
	freq := 145e6
	r := new(Performance)
	for _, Zl := range []complex128{
		40.684128 - 172.193310i,
		43.725893 - 152.978284i,
		46.851551 - 133.481797i,
		50.294661 - 114.383460i,
		53.856060 - 95.046731i,
		57.764091 - 75.926341i,
		61.832717 - 56.592882i,
		66.282562 - 37.317784i,
		70.945204 - 17.839074i,
		76.030469 + 1.720735i,
	} {
		r.Z = Zl

		z := Zl / Zs
		b := (Sqr(imag(z))+1)/real(z) + real(z)
		s := (b + math.Sqrt(Sqr(b)-4)) / 2
		f1 := 10 * math.Log10(1-Sqr((s-1)/(s+1)))

		f2 := 10 * math.Log10(real(r.Z)/cmplx.Abs(r.Z))

		_, m := Zmatch(Zs, Zl)
		C, L := m.LowPass(freq)
		w := CircAng * freq
		k := 1 / complex(1+Sqr(w)*L*C, w*L/real(r.Z))
		f3 := 10 * math.Log10(real(k)/cmplx.Abs(k))

		t.Logf("Zl=%s, [%.3f,%.3f,%.3f]\n",
			FormatImpedance(Zl, 5),
			f1, f2, f3)
	}
}

func TestEval2(t *testing.T) {

	Zs := complex(50, 0)
	for _, Zl := range []complex128{
		40.684128 - 172.193310i,
		43.725893 - 152.978284i,
		46.851551 - 133.481797i,
		50.294661 - 114.383460i,
		53.856060 - 95.046731i,
		57.764091 - 75.926341i,
		61.832717 - 56.592882i,
		66.282562 - 37.317784i,
		70.945204 - 17.839074i,
		76.030469 + 1.720735i,
	} {
		k := cmplx.Abs(Zl/Zs) + 1
		a := -10 * math.Log10(k)
		t.Logf("k=%f, a=%f", k, a)
	}
}
