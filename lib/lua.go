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
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"

	lua "github.com/Shopify/go-lua"
)

// LuaGenerator is a generator where the Nodes() method is implemented
// as a LUA script.
type LuaGenerator struct {
	script string            // script filename
	params map[string]string // map of parameters
	lambda float64           // wavelength
	state  *lua.State        // state of LUA VM
	angles []float64         // local angles
}

// Init generator with given parameters
func (g *LuaGenerator) Init(param string, lambda float64) error {
	g.lambda = lambda
	g.params = make(map[string]string)
	list := strings.SplitN(param, ":", 2)
	g.script = list[0]
	if len(list) > 1 {
		for _, p := range strings.Split(list[1], ",") {
			kv := strings.SplitN(p, "=", 2)
			if len(kv) == 2 {
				g.params[kv[0]] = kv[1]
			} else {
				g.params[kv[0]] = "bool:true"
			}
		}
	}
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
		vv := strings.SplitN(v, ":", 2)
		switch vv[0] {
		case "int":
			val, _ := strconv.Atoi(vv[1])
			g.state.PushInteger(val)
		case "num":
			val, _ := strconv.ParseFloat(vv[1], 64)
			g.state.PushNumber(val)
		case "bool":
			val, _ := strconv.ParseBool(vv[1])
			g.state.PushBoolean(val)
		default:
			g.state.PushString(vv[1])
		}
		g.state.SetGlobal(k)
	}
	if err := lua.DoFile(g.state, g.script); err != nil {
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
	return g.script
}

// Info about generator
func (g *LuaGenerator) Info() string {
	return "LUA script: " + g.Name()
}

// Volatile returns true if the generator is randomized
func (g *LuaGenerator) Volatile() bool {
	return true
}

//----------------------------------------------------------------------

// LuaEvaluator provides an Evaluate() function for optimization
// written in LUA script.
type LuaEvaluator struct {
	script string     // script filename
	prgm   string     // program
	state  *lua.State // state of LUA VM

	perf   *Performance // performance to evaluate
	args   string       // target mode
	feedZ  complex128   // source impedance
	result float64      // return value
}

// NewLuaEvaluator instantiates a new LUA evaluator:
// 'param' is of form '<script filename>:<opt1>=<val>,<opt2>=...'
func NewLuaEvaluator(script string) (ev *LuaEvaluator, err error) {
	var data []byte
	if data, err = os.ReadFile(script); err != nil {
		return
	}
	ev = new(LuaEvaluator)
	ev.script = script
	ev.prgm = string(data)
	ev.state = lua.NewState()
	lua.OpenLibraries(ev.state)

	ev.state.Register("source", func(state *lua.State) int {
		state.PushNumber(real(ev.feedZ))
		state.PushNumber(imag(ev.feedZ))
		return 2
	})
	ev.state.Register("args", func(state *lua.State) int {
		state.PushString(ev.args)
		return 1
	})
	ev.state.Register("perf_gain", func(state *lua.State) int {
		state.PushNumber(ev.perf.Rp.Min)
		state.PushNumber(ev.perf.Gain.Max)
		state.PushNumber(ev.perf.Gain.Mean)
		state.PushNumber(ev.perf.Gain.SD)
		return 4
	})
	ev.state.Register("perf_z", func(state *lua.State) int {
		state.PushNumber(real(ev.perf.Z))
		state.PushNumber(imag(ev.perf.Z))
		return 2
	})
	ev.state.Register("perf_rp_idx", func(state *lua.State) int {
		state.PushInteger(ev.perf.Rp.NPhi)
		state.PushInteger(ev.perf.Rp.NTheta)
		return 2
	})
	ev.state.Register("perf_rp_val", func(state *lua.State) int {
		phi, _ := state.ToInteger(1)
		theta, _ := state.ToInteger(2)
		state.PushNumber(ev.perf.Rp.Values[phi][theta])
		return 1
	})
	ev.state.Register("result", func(state *lua.State) int {
		ev.result, _ = state.ToNumber(1)
		return 0
	})

	return
}

// Evaluate antenna performance and return result
func (ev *LuaEvaluator) Evaluate(perf *Performance, args string, feedZ complex128) float64 {
	ev.perf, ev.args, ev.feedZ = perf, args, feedZ

	if err := lua.DoString(ev.state, ev.prgm); err != nil {
		log.Fatal(err)
	}
	return ev.result
}
