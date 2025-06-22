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

import "testing"

func TestMaterialProps(t *testing.T) {
	for mat := range material {
		c, i, err := MaterialProperties(mat, 0.002)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("Material: %s", mat)
		t.Logf("  Conductance = %e S/m", c)
		t.Logf("  Inductance  = %e H/m", i)
	}
}
