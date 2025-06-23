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
	"math/rand"
	"strconv"
	"strings"

	lua "github.com/Shopify/go-lua"
)

// LuaGenerator is a generator where the Nodes() method is implemented
// as a LUA script.
type LuaGenerator struct {
	params map[string]string // map of parameters
	lambda float64           // wavelength
	state  *lua.State        // state of LUA VM
	angles []float64         // local angles
}

// Init generator with given parameters
func (g *LuaGenerator) Init(param string, lambda float64) error {
	g.params = make(map[string]string)
	for _, p := range strings.Split(param, ",") {
		kv := strings.SplitN(p, "=", 2)
		if len(kv) == 2 {
			g.params[kv[0]] = kv[1]
		} else {
			g.params[kv[0]] = ""
		}
	}
	if _, ok := g.params["scr"]; !ok {
		return errors.New("no script specified")
	}
	g.lambda = lambda
	g.state = lua.NewState()
	lua.OpenLibraries(g.state)
	return nil
}

// Nodes returns the initial antenna geometry made from 'num' segments
// of equal length 'segL'. Volatile generators build varying geometries
// based on randomization.
func (g *LuaGenerator) Nodes(num int, segL float64, rnd *rand.Rand) []Node {
	g.angles = make([]float64, num)

	g.state.PushInteger(num)
	g.state.SetGlobal("num")
	g.state.PushNumber(segL)
	g.state.SetGlobal("segL")
	g.state.Register("rnd", func(state *lua.State) int {
		state.PushNumber(rnd.Float64())
		return 1
	})
	g.state.Register("setAngle", func(state *lua.State) int {
		i, _ := state.ToInteger(1)
		ang, _ := state.ToNumber(2)
		g.angles[i] = ang
		return 0
	})
	for k, v := range g.params {
		if k == "scr" {
			continue
		}
		vv := strings.SplitN(v, ":", 2)
		switch vv[0] {
		case "int":
			val, _ := strconv.Atoi(vv[1])
			g.state.PushInteger(val)
			g.state.SetGlobal(k)
		case "num":
			val, _ := strconv.ParseFloat(vv[1], 64)
			g.state.PushNumber(val)
			g.state.SetGlobal(k)
		}
	}
	if err := lua.DoFile(g.state, g.params["scr"]); err != nil {
		panic(err)
	}
	nodes := make([]Node, num)
	for i, ang := range g.angles {
		nodes[i] = NewNode2D(segL, ang)
	}
	return nodes
}

// Name of generator
func (g *LuaGenerator) Name() string {
	return g.params["scr"]
}

// Info about generator
func (g *LuaGenerator) Info() string {
	return "LUA script: " + g.Name()
}

// Volatile returns true if the generator is randomized
func (g *LuaGenerator) Volatile() bool {
	return true
}
