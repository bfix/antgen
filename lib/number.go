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
	"math"
	"strconv"
	"strings"
)

// ParseImpedance (complex value) from string.
// A valid string is formed from one or two numbers combined; the single
// number or one of the two numbers can be tagged by a "j" or "i" as
// imaginary. Spaces and multiplication signs in a string are ignored.
//
// Examples of valid strings:
// * "50"     		// only real part -> (50,0)
// * "-j30.624"		// only imaginary part -> (0,-30.624)
// * "87.37+j41.74" // complex number -> (87.37,41.74)
// * "j41.74+87.37" // complex number -> (87.37,41.74)
func ParseImpedance(s string) (Z complex128, err error) {
	// remove redundant runes from string
	var t string
	for _, r := range s {
		if !strings.ContainsRune(" *·", r) {
			t += string(r)
		}
	}
	s = strings.ReplaceAll(t, "i", "j")

	// parse impedance
	var r, i float64
	if pos := max(strings.IndexRune(s, '+'), strings.IndexRune(s, '-')); pos < 1 {
		// only one part
		im := strings.ContainsRune(s, 'j')
		s = strings.ReplaceAll(t, "j", "")
		if r, err = ParseNumber(s); err != nil {
			return
		}
		if im {
			Z = complex(0, r)
		} else {
			Z = complex(r, 0)
		}

	} else {
		// split string into two values
		sign := (s[pos] == '-')
		v1, v2 := s[:pos], s[pos+1:]
		if strings.ContainsRune(v1, 'j') {
			v1, v2 = v2, v1
		}
		v2 = strings.Replace(v2, "j", "", 1)

		if r, err = ParseNumber(v1); err != nil {
			return
		}
		if i, err = ParseNumber(v2); err != nil {
			return
		}
		if sign {
			i = -i
		}
		Z = complex(r, i)
	}
	return
}

// FormatImpedance with scaled numbers (magnitude)
func FormatImpedance(z complex128, n int) string {
	if ic := imag(z); math.Abs(ic) > 1e-12 {
		s := '+'
		if ic < 0 {
			s = '-'
			ic = math.Abs(ic)
		}
		return fmt.Sprintf("%s %c j·%s",
			FormatNumber(real(z), n), s, FormatNumber(ic, n),
		)
	} else {
		return FormatNumber(real(z), n)
	}
}

const (
	mags = "fpnum kMGTP" // magnitudes from -15 to 15
)

// ParseNumber with magnitude
func ParseNumber(s string) (float64, error) {
	rs := []rune(strings.TrimSpace(s))
	lr := len(rs)
	if lr == 0 {
		return 0, errors.New("empty number string")
	}
	f := 1.
	if i := strings.IndexRune(mags, rs[lr-1]); i != -1 {
		f = math.Pow10(-15 + 3*i)
		rs = rs[:lr-1]
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(string(rs)), 64)
	if err != nil {
		return 0, err
	}
	return f * v, nil
}

// FormatNumber with magnitude
func FormatNumber(v float64, n int) string {
	sign := ' '
	if v < 0 {
		sign = '-'
	}
	v = math.Abs(v)
	for i, mag := range mags {
		f := v / math.Pow10(-15+3*i)
		if f < 1000 || i == len(mags)-1 {
			k := (n - 1) - int(math.Log10(f))
			return strings.TrimSpace(fmt.Sprintf("%c%*.*f %c", sign, n, k, f, mag))
		}
	}
	return ""
}
