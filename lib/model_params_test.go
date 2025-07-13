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

func TestParse(t *testing.T) {
	lines := []string{
		"Source: 435000000:50.000000:0.000000",
		"Wire: 0.002:CuL:5.960e+07:1.100e-07",
		"Feedpoint: 0.005:0.000",
		"Ground: 0.000:0:-1:0:0.000000:0.000000",
		"Param: 0.100000::100",
		"Mode: bend2d:straight:1000:none",
		"Init: 2.290155:-2.211046:41.871386:7.280286:-449.239881",
		"Result: 2.290155:-2.211046:41.871386:7.280286:-449.239881",
		"Stats: 0:0:0:0",
	}
	p, ok, err := ParseMdlParams(lines)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("!OK")
	}
	t.Logf("%v", p)
}
