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
	"fmt"
	"log"
	"math"
)

// wire material properties used in simulation
type matProp struct {
	conductivity float64 // unit: S/m
	permeability float64 // relative to μ₀ (unit: H/m)
	permittivity float64 // relative to ε₀ (unit: F/m)
}

// list of known materials
var material = map[string]matProp{
	"Cu": { // Kupferdraht
		5.96e7,    // conductivity (S/m)
		0.9999936, // relative permeability (H/m)
		1.108,     // relative permittivity (F/m)
	},
	"CuL": { // Kupferlackdraht
		5.96e7,    // conductivity (S/m)
		0.1166507, // relative permeability (H/m)
		1.108,     // relative permittivity (F/m)
	},
	"Al": { // Aluminium
		3.5e7,     // conductivity (S/m)
		1.0000220, // relative permeability (H/m)
		1.3,       // relative permittivity (F/m)
	},
}

// MaterialProperties returns material properties for label
func MaterialProperties(label string, dia float64) (conductivity, inductance float64, err error) {
	mp, ok := material[label]
	if !ok {
		err = fmt.Errorf("unknown material '%s'", label)
	} else {
		conductivity = mp.conductivity
		// compute inductance
		inductance = (Mu_0 * mp.permeability / CircAng) * (math.Log(4/dia) - 1)
	}
	return
}

// GetWire for specified diameter and material
func GetWire(mat string, dia float64) Wire {
	G, L, err := MaterialProperties(mat, dia)
	if err != nil {
		log.Fatalf("unknown material '%s'", mat)
	}
	return Wire{
		Diameter:     dia,
		Conductivity: G,
		Inductance:   L,
	}
}
