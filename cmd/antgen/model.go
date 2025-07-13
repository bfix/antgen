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

package main

import (
	"fmt"
	"strings"

	"github.com/bfix/antgen/lib"
)

// List of all available models in AntGen
var mdls = make(map[string]func(int) (lib.Model, error))

// GetModel by name
func GetModel(name string, spec *lib.Specification, gen lib.Generator, verbose int) (mdl lib.Model, side float64, err error) {
	s := strings.SplitN(name, ":", 2)
	mdlF, ok := mdls[s[0]]
	if !ok {
		err = fmt.Errorf("no such model '%s'", name)
		return
	}
	if mdl, err = mdlF(verbose); err != nil {
		return
	}
	params := ""
	if len(s) > 1 {
		params = s[1]
	}
	side, err = mdl.Init(params, spec, gen)
	return
}
