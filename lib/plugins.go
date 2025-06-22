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
	"plugin"
)

// list of known (and loaded plugins)
var plugins = make(map[string]*plugin.Plugin)

// GetPlugin by name.
// If name is prefixed with '@', it references a plugin entry
// in the configuration.
func GetPlugin(name string) (pi *plugin.Plugin, err error) {
	var ok bool
	// check for config reference
	if name[0] == '@' {
		if name, ok = Cfg.Plugins[name[1:]]; !ok {
			err = fmt.Errorf("referenced plugin '%s' not defined", name[1:])
			return
		}
	}
	// load plugin with path to shared library
	if pi, ok = plugins[name]; !ok {
		if pi, err = plugin.Open(name); err == nil {
			plugins[name] = pi
		}
	}
	return
}

// GetSymbol from plugin (exported variable or function)
func GetSymbol[T any](pi *plugin.Plugin, name string) (sym T, err error) {
	var f plugin.Symbol
	if f, err = pi.Lookup(name); err == nil {
		sym = f.(T)
	}
	return
}
