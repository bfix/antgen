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
	"testing"
)

func TestMatch(t *testing.T) {

	Zs := complex(50, 0)
	Zl := complex(5, 0)
	f := 145000000.

	Z, matcher := Zmatch(Zs, Zl)

	t.Logf("AtSource=%v, Zmatch=%s\n", matcher.AtSource, FormatImpedance(Z, 5))

	Cp, Ls := matcher.LowPass(f)
	t.Logf("LP: Cp=%sF, Ls=%sH\n", FormatNumber(Cp, 4), FormatNumber(Ls, 4))
	Cs, Lp := matcher.HighPass(f)
	t.Logf("Cs=%sF, Lp=%sH\n", FormatNumber(Cs, 4), FormatNumber(Lp, 4))
}
