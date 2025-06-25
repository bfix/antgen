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

func TestLuaEvaluator(t *testing.T) {

	// construct antenna
	spec := &Specification{
		Wire: GetWire("CuL", 0.002),
		Ground: Ground{
			Height: 0,
			Mode:   0,
			Type:   -1,
			NRadl:  0,
			Epse:   0,
			Sig:    0,
		},
		Source: Source{
			Z:     Impedance{50, 0},
			Power: 1,
			Freq:  435000000,
			Span:  5000000,
		},
	}

	ev, err := NewLuaEvaluator("./evaluator_test.lua")
	if err != nil {
		t.Fatal(err)
	}
	for i := range 5 {
		nodes := make([]Node, 2)
		nodes[0] = NewNode2D(0.01, 0)
		nodes[1] = NewNode2D((0.25+float64(i)/20)*spec.Source.Lambda(), 0)
		ant := BuildAntenna("test", spec, nodes)

		// simulate antenna
		if err := ant.Eval(spec.Source.Freq, spec.Wire, spec.Ground); err != nil {
			t.Fatal(err)
		}

		res := ev.Evaluate(ant.Perf, "matched", spec.Source.Impedance())
		t.Log(res)
	}
}
