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
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

// Generator interface
type Generator interface {

	// Init generator with given parameters
	Init(param string, lambda float64) error

	// Nodes returns the initial antenna geometry made from 'num' segments
	// of equal length 'segL'. Volatile generators build varying geometries
	// based on randomization.
	Nodes(num int, segL float64, rnd *rand.Rand) []*Node

	// Name of generator
	Name() string

	// Info about generator
	Info() string

	// Volatile returns true if the generator is randomized
	Volatile() bool
}

// list of implemented generators (by name)
var gens map[string]Generator

// register implemented generators.
func init() {
	set := func(g Generator) {
		gens[g.Name()] = g
	}
	gens = make(map[string]Generator)
	set(new(GenStraight))
	set(new(GenV))
	set(new(GenWalk))
	set(new(GenStroll))
	set(new(GenTrespass))
	set(new(GenGeo))
}

// GetGenerator by name
func GetGenerator(name string, lambda float64) (g Generator, err error) {
	s := strings.SplitN(name, ":", 2)
	param := ""
	if len(s) > 1 {
		param = s[1]
	}
	if s[0] == "lua" {
		g = new(LuaGenerator)
		err = g.Init(param, lambda)
		return
	}
	var ok bool
	if g, ok = gens[s[0]]; !ok {
		return nil, fmt.Errorf("unknown generator '%s'", name)
	}
	err = g.Init(param, lambda)
	return
}

//----------------------------------------------------------------------

// BendMax returns the max. bending angle between two segments of given
// length such that a resulting curve has a minimum radius of r.
func BendMax(r, segL float64) float64 {
	U := CircAng * r
	n := math.Ceil(U / segL)
	return math.Pi / n
}

//----------------------------------------------------------------------

// GenStraight returns all segments lined-up in a straight line.
type GenStraight struct{}

// Init generator with given parameters
func (g *GenStraight) Init(params string, lambda float64) error {
	return nil
}

// Nodes returns the initial antenna geometry made from 'num' segments
// of equal length 'segL'.
func (g *GenStraight) Nodes(num int, segL float64, rnd *rand.Rand) []*Node {
	nodes := make([]*Node, num)
	for i := range num {
		nodes[i] = NewNode(segL, 0, 0)
	}
	return nodes
}

// Name of generator
func (g *GenStraight) Name() string {
	return "straight"
}

// Info about generator
func (g *GenStraight) Info() string {
	return g.Name()
}

// Volatile returns true if the generator is randomized
func (g *GenStraight) Volatile() bool {
	return false
}

//----------------------------------------------------------------------

// GenV returns a V-shaped dipole with 120° (¾π) angle between legs.
type GenV struct {
	lambda float64 // wavelength
	ang    float64 // opening angle
	rad    float64 // bending radius
	end    bool    // bend back at end of leg
	params string  // supplied parameters
}

// Init generator with given parameters
func (g *GenV) Init(params string, lambda float64) (err error) {
	// set parameter default values
	g.lambda = lambda
	g.rad = 5
	g.end = false
	g.ang = math.Pi / 6
	g.params = params

	// handle parameters
	for _, p := range strings.Split(params, ",") {
		v := strings.SplitN(p, "=", 2)
		switch v[0] {
		case "rad":
			if g.rad, err = strconv.ParseFloat(v[1], 64); err != nil {
				return
			}
		case "end":
			g.end = true
		case "ang":
			if g.ang, err = strconv.ParseFloat(v[1], 64); err != nil {
				return
			}
			g.ang = (180 - g.ang) * math.Pi / 360
		}
	}
	return nil
}

// Nodes returns the initial antenna geometry made from 'num' segments
// of equal length 'segL'.
func (g *GenV) Nodes(num int, segL float64, rnd *rand.Rand) []*Node {
	rnum := 1
	if !IsNull(g.rad) {
		bendMax := g.rad * segL / g.lambda
		rnum = int(math.Ceil(g.ang / bendMax))
	}
	dAng := g.ang / float64(rnum)
	nodes := make([]*Node, num)
	for i := range num {
		ang := 0.
		if i < rnum {
			ang = dAng
		} else if g.end && i >= num-rnum {
			ang = -dAng
		}
		nodes[i] = NewNode(segL, ang, 0)
	}
	return nodes
}

// Info about generator
func (g *GenV) Info() string {
	if len(g.params) > 0 {
		return fmt.Sprintf("%s[%s]", g.Name(), g.params)
	}
	return g.Name()
}

// Name of generator
func (g *GenV) Name() string {
	return "v"
}

// Volatile returns true if the generator is randomized
func (g *GenV) Volatile() bool {
	return false
}

//----------------------------------------------------------------------

// GenWalk grows a line (dipole leg) by moving in one direction, so the
// maximum direction vector of a segment is ±½π.
type GenWalk struct {
	lambda float64
	rng    int
	params string
}

// Init generator with given parameters
func (g *GenWalk) Init(params string, lambda float64) error {
	g.lambda = lambda
	g.params = params
	for _, p := range strings.Split(params, ",") {
		v := strings.SplitN(p, "=", 2)
		switch v[0] {
		case "smooth":
			k, err := strconv.Atoi(v[1])
			if err != nil {
				return err
			}
			g.rng = int(k)
		}
	}
	return nil
}

// Nodes returns the initial antenna geometry made from 'num' segments
// of equal length 'segL'.
func (g *GenWalk) Nodes(num int, segL float64, rnd *rand.Rand) []*Node {
	bendMax := BendMax(Cfg.Sim.MinRadius*g.lambda, segL)
	nodes := make([]*Node, num)
	dir := 0.
	for i := range num {
		ang := 2 * (rnd.Float64() - 0.5) * bendMax
		if math.Abs(dir+ang) > RectAng {
			ang = -ang
		}
		nodes[i] = NewNode(segL, ang, 0)
		dir += ang
	}
	if g.rng > 0 {
		nodes = Smooth2D(nodes, g.rng)
	}
	return nodes
}

// Info about generator
func (g *GenWalk) Info() string {
	if len(g.params) > 0 {
		return fmt.Sprintf("%s[%s]", g.Name(), g.params)
	}
	return g.Name()
}

// Name of generator
func (g *GenWalk) Name() string {
	return "walk"
}

// Volatile returns true if the generator is randomized
func (g *GenWalk) Volatile() bool {
	return true
}

//----------------------------------------------------------------------

// GenStroll grows a line (dipole leg) by moving in any direction but
// bounded by to positive x-coordinates.
type GenStroll struct {
	lambda float64
	rng    int
	params string
}

// Init generator with given parameters
func (g *GenStroll) Init(params string, lambda float64) error {
	g.lambda = lambda
	g.params = params
	for _, p := range strings.Split(params, ",") {
		v := strings.SplitN(p, "=", 2)
		switch v[0] {
		case "smooth":
			k, err := strconv.Atoi(v[1])
			if err != nil {
				return err
			}
			g.rng = int(k)
		}
	}
	return nil
}

// Nodes returns the initial antenna geometry made from 'num' segments
// of equal length 'segL'.
func (g *GenStroll) Nodes(num int, segL float64, rnd *rand.Rand) []*Node {
	bendMax := BendMax(Cfg.Sim.MinRadius*g.lambda, segL)
	nodes := make([]*Node, num)
	dir := 0.
	x := 0.
	for i := range num {
		ang := 2 * (rnd.Float64() - 0.5) * bendMax
		xn := x + segL*math.Cos(dir+ang)
		if xn < 4*segL && InRange(dir, RectAng, 3*RectAng) {
			mDir := (3 * segL / (xn - segL)) * math.Pi / 6
			cDir := dir - math.Pi
			ang = mDir - cDir
			if cDir < 0 {
				ang = -ang
			}
		}
		nodes[i] = NewNode(segL, ang, 0)
		x = xn
		dir = math.Mod(CircAng+dir+ang, CircAng)
	}
	if g.rng > 0 {
		return Smooth2D(nodes, g.rng)
	}
	return nodes
}

// Info about generator
func (g *GenStroll) Info() string {
	if len(g.params) > 0 {
		return fmt.Sprintf("%s[%s]", g.Name(), g.params)
	}
	return g.Name()
}

// Name of generator
func (g *GenStroll) Name() string {
	return "stroll"
}

// Volatile returns true if the generator is randomized
func (g *GenStroll) Volatile() bool {
	return true
}

//----------------------------------------------------------------------

// GenGeo reads a geometry file instead of generating something new.
type GenGeo struct {
	fName string
}

// Init generator with given parameters
func (g *GenGeo) Init(params string, lambda float64) error {
	g.fName = params
	return nil
}

// Nodes returns the initial antenna geometry made from 'num' segments
// of equal length 'segL'.
func (g *GenGeo) Nodes(num int, segL float64, rnd *rand.Rand) []*Node {

	// read geometry file
	body, err := os.ReadFile(g.fName)
	if err != nil {
		log.Fatal(err)
	}
	geo := new(Geometry)
	if err = json.Unmarshal(body, &geo); err != nil {
		log.Fatal(err)
	}

	return geo.Nodes
}

// Info about generator
func (g *GenGeo) Info() string {
	return fmt.Sprintf("%s[%s]", g.Name(), g.fName)
}

// Name of generator
func (g *GenGeo) Name() string {
	return "geo"
}

// Volatile returns true if the generator is randomized
func (g *GenGeo) Volatile() bool {
	return false
}

//----------------------------------------------------------------------

// GenTrespass grows a line (dipole leg) by moving in any direction
// and widthout bounds.
type GenTrespass struct {
	lambda float64
	rng    int
	params string
}

// Init generator with given parameters
func (g *GenTrespass) Init(params string, lambda float64) error {
	g.lambda = lambda
	g.params = params
	for _, p := range strings.Split(params, ",") {
		v := strings.SplitN(p, "=", 2)
		switch v[0] {
		case "smooth":
			k, err := strconv.Atoi(v[1])
			if err != nil {
				return err
			}
			g.rng = int(k)
		}
	}
	return nil
}

// Nodes returns the initial antenna geometry made from 'num' segments
// of equal length 'segL'.
func (g *GenTrespass) Nodes(num int, segL float64, rnd *rand.Rand) []*Node {
	bendMax := BendMax(Cfg.Sim.MinRadius*g.lambda, segL)
	nodes := make([]*Node, num)
	dir := 0.
	for i := range num {
		ang := 2 * (rnd.Float64() - 0.5) * bendMax
		nodes[i] = NewNode(segL, ang, 0)
		dir += ang
	}
	return nodes
}

// Info about generator
func (g *GenTrespass) Info() string {
	if len(g.params) > 0 {
		return fmt.Sprintf("%s[%s]", g.Name(), g.params)
	}
	return g.Name()
}

// Name of generator
func (g *GenTrespass) Name() string {
	return "trespass"
}

// Volatile returns true if the generator is randomized
func (g *GenTrespass) Volatile() bool {
	return true
}
