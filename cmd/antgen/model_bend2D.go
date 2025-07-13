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
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/bfix/antgen/lib"
)

func init() {
	mdls["bend2d"] = NewModelBend2D
}

//----------------------------------------------------------------------

// ModelBend2D is a dipole model where the joints
// of two segments can be bended (in the XY plane).
type ModelBend2D struct {
	lib.ModelDipole

	rnd  *rand.Rand    // randomizer
	seed int64         // randomizer seed
	gen  lib.Generator // reference to generator
	best *lib.Antenna  // antenna with best performance

	verbose int // verbosity

	bendStep float64
	bendMin  float64
	bendMax  float64
}

// NewModelBend2D instaniates a new optimizer model
func NewModelBend2D(verbose int) (lib.Model, error) {
	return &ModelBend2D{verbose: verbose}, nil
}

// Init model
func (mdl *ModelBend2D) Init(params string, spec *lib.Specification, gen lib.Generator) (side float64, err error) {
	// no parameters expected
	if len(params) > 0 {
		err = errors.New("no parameters expected")
		return
	}
	// check for valid generator
	if gen == nil {
		err = errors.New("no generator defined")
		return
	}
	mdl.gen = gen

	// init dipole
	side, err = mdl.ModelDipole.Init(params, spec, gen)

	// compute bending angles (min, max, step)
	mdl.bendMax = lib.BendMax(lib.Cfg.Sim.MinRadius*spec.Source.Lambda(), mdl.SegL)
	mdl.bendMin = mdl.bendMax * lib.Cfg.Sim.MinBend
	mdl.bendStep = mdl.bendMax / 3

	return
}

// Info returns model information
func (mdl *ModelBend2D) Info() string {
	return "bend2d"
}

// Prepare initial geometry.
func (mdl *ModelBend2D) Prepare(seed int64, cb lib.Callback) (ant *lib.Antenna, err error) {
	// deterministic random numbers
	mdl.rnd = lib.Randomizer(seed)
	mdl.seed = seed

	// generate the initial geometry
	mdl.Nodes = mdl.gen.Nodes(mdl.Num, mdl.SegL, mdl.rnd)
	mdl.Num = len(mdl.Nodes)
	if mdl.best, err = mdl.eval(); err != nil {
		return
	}
	ant = mdl.best

	// track folding into initial geometry
	mdl.Track = lib.Changes(mdl.Nodes)
	mdl.Track = append(mdl.Track, &lib.Change{Pos: lib.TRK_MARK})

	cb(mdl.best, -1, "initial geometry")
	return
}

// Optimize model and return best antenna geometry
func (mdl *ModelBend2D) Optimize(seed int64, iter int, cmp *lib.Comparator, cb lib.Callback) (ant *lib.Antenna, stats lib.Stats, err error) {

	// pick random segments and change their angle (direction).
	// revert change if gain is not increasing
	start := time.Now()
	stats.NumMthds = 1

	// optimize antenna by bending
	var steps, sims int
	if ant, steps, sims, err = mdl.optBend(iter, cmp, cb); err != nil {
		return
	}
	stats.NumSteps += steps
	stats.NumSims += sims

	stats.Elapsed = time.Since(start).Round(time.Second)
	cb(ant, -1, fmt.Sprintf("optimized geometry (%s)", cmp.Target()))
	return
}

// Optimize geometry by bending the wire at joints between segments
func (mdl *ModelBend2D) optBend(iter int, cmp *lib.Comparator, cb lib.Callback) (ant *lib.Antenna, steps, sims int, err error) {

	lastVal, valChange, dw := math.NaN(), math.NaN(), 0.
	pos, tries, maxTries := -1, 0, 0

	for i := 1; ; i++ {
		// show progress
		if ant != nil && mdl.verbose > 0 {
			fmt.Printf("\r%d: bend [%4d] %5d -- %.6f / %.6f  %s\033[0K",
				mdl.seed, steps, i, valChange, lastVal, mdl.best.Perf.String())
		}
		// pick a random position if not set
		if pos == -1 {
			pos = mdl.rnd.Intn(mdl.Num)
		}

		// vary bend angle of node
		dw = 2 * (mdl.rnd.Float64() - 0.5) * mdl.bendStep
		if math.Abs(dw) < mdl.bendMin {
			pos = -1
			continue
		}
		node := mdl.Nodes[pos]
		// limit bending to max
		if math.Abs(node.Theta+dw) > mdl.bendMax {
			pos = -1
			continue
		}
		// check geometry
		node.AddAngles(dw, 0)
		if !mdl.checkGeometry() {
			node.AddAngles(-dw, 0)
			pos = -1
			continue
		}
		// evaluate new antenna geometry
		ant, err = mdl.eval()
		if err != nil {
			return
		}
		sims++

		// NEC2 safe-guard: terminate optimization if resistance
		// goes below 1Ω or above 20kΩ (defaults, can use custom range)
		if r := real(ant.Perf.Z); r < lib.Cfg.Sim.MinZr || r > lib.Cfg.Sim.MaxZr {
			break
		}

		// quit after max number of rounds
		if tries++; tries > maxTries+mdl.Num*lib.Cfg.Sim.MaxRounds {
			break
		}

		// check for improved performance
		if sign, val := cmp.Compare(ant.Perf, mdl.best.Perf); sign == 1 {
			mdl.best = ant
			mdl.Track = append(mdl.Track, &lib.Change{
				Pos:   pos,
				Theta: dw,
			})

			// render geometry (if applicable)
			i = 0
			steps++
			cb(ant, pos, fmt.Sprintf("Step #%d", steps))
			if iter == steps {
				break
			}

			// check progress
			if steps%lib.Cfg.Sim.ProgressCheck == 0 {
				if !math.IsNaN(lastVal) {
					if valChange = (val - lastVal); valChange < lib.Cfg.Sim.MinChange {
						// optimum reached
						break
					}
				}
				lastVal = val
				if tries > maxTries {
					maxTries = tries
				}
				tries = 0
			}
		} else {
			node.AddAngles(-dw, 0)
			pos = -1
		}
	}
	ant = mdl.best
	fmt.Printf("\r\033[0K")
	return
}

// check geometry (bounded to positive x-coordinates)
func (mdl *ModelBend2D) checkGeometry() (ok bool) {
	d := mdl.Nodes[0].Length
	pos := lib.NewVec3(d/2, 0, 0)
	dir := 0.
	for _, node := range mdl.Nodes {
		dir = math.Mod(dir+node.Theta, lib.CircAng)
		end := pos.Move2D(node.Length, dir)
		if end[0] < d/2 {
			return
		}
		pos = end
	}
	ok = true
	return
}

// evaluate performance of antenna geometry
func (mdl *ModelBend2D) eval() (ant *lib.Antenna, err error) {
	ant = lib.BuildAntenna(mdl.Kind, mdl.Spec, mdl.Nodes)
	// ant.DumpNEC(mdl.spec, nil, "./curr.nec")
	err = ant.Eval(mdl.Spec.Source.Freq, mdl.Spec.Wire, mdl.Spec.Ground)
	return
}
