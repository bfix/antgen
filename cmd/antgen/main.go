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
	_ "embed"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strings"

	"github.com/bfix/antgen/lib"
)

//go:generate sh -c "printf %s $(git describe --tags) > _version"
//go:embed _version
var Version string

//go:generate sh -c "printf %s $(date +%F) > _date"
//go:embed _date
var Date string

// Dipole optimization:
//
// Optimize a dipole for a given frequency ('-freq'); if a frequency range is
// specified, optimize for the center frequency. The range info (if available)
// is used to generate a matching "FR" card for NEC2.
//
//	----------------_----------------          ^ Z
//	|<--   λ/k   -->|<--   λ/k   -->|          |
//	                ^- Excitation            Y x--> X
//
// The antenna is made out of a wire with specific properties ('-wire') and is
// possibly mounted over ground ('-ground'). The half-length of the dipole is
// specified as a fraction ('-k') of the wavelength of the (center) frequency.
// The dipole is center-fed from a source ('-source') with defined impedance
// and output power.
//
// The initial (pre-optimization) geometry of the antenna is assembled by a
// generator ('-gen'); a generator can be volatile (meaning the geometry is
// based on some kind of seeded randomization '-seed') or static (like a
// straight line or a V-shaped dipole). The generator creates only one half
// of the dipole; the other half is mirrored on the YZ plane.
//
// The initial geometry is optimized by an optimization model for the
// specified target ('-opt') using an optimization model ('-model').
//
// Optimizations are written into files in the output directory ('-out').

func main() {
	// handle command-line
	var (
		spec   = new(lib.Specification) // Antenna specifications
		config string                   // configuration file

		freqS   string // 'freq' option: either single freq or freq range
		wireS   string // wire specification
		groundS string // ground specification
		sourceS string // source parameters (without frequency)
		feedptS string // feedpoint parameters

		param float64 // free parameter
		seed  int64   // seed for deterministic randomization
		gen   string  // generator model to use

		model  string // optimization model to use (incl. parameters)
		target string // optimize for target [Gmax, GMean, SD, none]
		iter   int    // number of iterations; 0=no limit
		vis    bool   // visualize optimizations
		logr   bool   // log iteration results
		warn   bool   // emit warnings

		tag     string // tag for output filename
		outDir  string // directory for optimization output
		outPrf  string // filename prefix
		verbose int    // verbose output

		ant *lib.Antenna
		err error
	)
	flag.StringVar(&config, "config", "", "configuration file")
	flag.StringVar(&freqS, "freq", "", "Frequency (default: 430M-440M)")
	flag.Float64Var(&spec.K, "k", lib.Cfg.Def.K, "side extend kλ (default: 0.25λ)")
	flag.StringVar(&wireS, "wire", "", "wire parameter")
	flag.StringVar(&groundS, "ground", "", "antenna height")
	flag.StringVar(&sourceS, "source", "", "feed parameters")
	flag.StringVar(&feedptS, "feedpt", "", "feed point")

	flag.StringVar(&gen, "gen", "stroll", "generator for initial geometry")

	flag.StringVar(&model, "model", "bend2d", "model selection")
	flag.StringVar(&target, "opt", "Gmax", "optimization target (default: Gmax)")

	flag.Int64Var(&seed, "seed", 1000, "model seed")
	flag.IntVar(&iter, "iter", 0, "optimization iterations")

	flag.Float64Var(&param, "param", math.NaN(), "free parameter")
	flag.StringVar(&tag, "tag", "", "output name tag")
	flag.StringVar(&outDir, "out", "./out", "output directory")
	flag.StringVar(&outPrf, "prefix", "", "output prefix")

	flag.IntVar(&verbose, "verbose", 1, "verbosity")
	flag.BoolVar(&vis, "vis", false, "visualize iterations")
	flag.BoolVar(&logr, "log", false, "log iterations")
	flag.BoolVar(&warn, "warn", false, "emit warning")
	flag.Parse()

	// handle optional configuration file
	if len(config) > 0 {
		if err = lib.ReadConfig(config); err != nil {
			log.Fatal(err)
		}
	}

	// handle wire parameters
	if spec.Wire, err = lib.ParseWire(wireS, warn); err != nil {
		log.Fatal(err)
	}

	// handle source parameters
	if spec.Source, err = lib.ParseSource(sourceS, warn); err != nil {
		log.Fatal(err)
	}

	// handle feed point parameters
	if spec.Feedpt, err = lib.ParseFeedpt(feedptS, warn); err != nil {
		log.Fatal(err)
	}

	// change specified source frequency (range)
	if len(freqS) > 0 {
		if spec.Source.Freq, spec.Source.Span, err = lib.GetFrequencyRange(freqS); err != nil {
			log.Fatal(err)
		}
	}

	// handle ground parameters
	if spec.Ground, err = lib.ParseGround(groundS, warn); err != nil {
		log.Fatal(err)
	}

	// get generator model
	g, err := lib.GetGenerator(gen, spec.Source.Lambda())
	if err != nil {
		log.Fatal(err)
	}

	// get optimization model
	mdl, side, err := GetModel(model, spec, g, verbose)
	if err != nil {
		log.Fatal(err)
	}

	// setup comparator
	var cmp *lib.Comparator
	if cmp, err = lib.NewComparator(target, spec); err != nil {
		log.Fatal(err)
	}

	// run optimization in goroutine to allow rendering
	var steps []string
	var step int
	var iniPerf *lib.Performance
	optimize := func(render lib.Canvas) (total lib.Stats) {
		// callback for opt iteration
		cb := func(ant *lib.Antenna, pos int, msg string) {
			if render != nil {
				render.Show(ant, pos, msg)
			}
			step++
			if logr {
				msg := fmt.Sprintf("[%5d] %s", step, ant.Perf.String())
				steps = append(steps, msg)
			}
		}
		// prepare initial geometry
		if ant, err = mdl.Prepare(seed, cb); err != nil {
			log.Printf("Model #%d: %s", seed, err.Error())
			return
		}
		iniPerf = ant.Perf

		// check for optimization
		if target != "none" {
			// optimize antenna (multiple optimizers in sequence possible)
			var stats lib.Stats
			for {
				if ant, stats, err = mdl.Optimize(seed, iter, cmp, cb); err != nil {
					log.Printf("Model #%d: %s", seed, err.Error())
					return
				}
				total.Elapsed += stats.Elapsed
				total.NumMthds += stats.NumMthds
				total.NumSteps += stats.NumSteps
				total.NumSims += stats.NumSims

				// switch to next optimizer
				if !cmp.Next() {
					break
				}
			}
		}
		return
	}

	// setup rendering (if visualization is requested)
	var total lib.Stats
	if vis {
		var render lib.Canvas
		if render, err = lib.GetCanvasFromCfg(lib.Cfg.Render, side); err != nil {
			log.Fatal(err)
		}
		defer render.Close()
		go func() {
			total = optimize(render)
		}()
		render.Run(nil)
	} else {
		total = optimize(nil)
	}
	if ant == nil {
		log.Fatal("Aborted...")
	}

	// output optimization results
	if len(tag) == 0 {
		tag = fmt.Sprintf("%d", seed)
	}
	log.Printf("Model #%s: %s (%d/%d/%d in %s)\n", tag, ant.Perf.String(),
		total.NumMthds, total.NumSteps, total.NumSims, total.Elapsed)
	if !logr {
		steps = nil
	}

	// intro and assemble comments
	var cmts []string
	cmts = append(cmts, fmt.Sprintf("AntGen %s (%s) - Copyright 2024-present Bernd Fix   >Y<", Version, Date))
	cmts = append(cmts, lib.GenMdlParams(param, spec, iniPerf, ant.Perf, model, g.Info(), target, seed, tag, total)...)

	// handle output prefix
	if len(outPrf) > 0 && !strings.HasSuffix(outPrf, "_") {
		outPrf += "_"
	}
	// write model to file
	fName := fmt.Sprintf("%s/%smodel-%s.nec", outDir, outPrf, tag)
	wrt, err := os.Create(fName)
	if err != nil {
		log.Fatal(err)
	}
	defer wrt.Close()
	ant.DumpNEC(wrt, spec, cmts)
	mdl.Finalize(tag, outDir, outPrf, cmts)

	// handle logging
	if len(steps) > 0 {
		fName := fmt.Sprintf("%s/%ssteps-%s.log", outDir, outPrf, tag)
		logF, err := os.Create(fName)
		if err != nil {
			log.Fatal(err)
		}
		for _, line := range steps {
			fmt.Fprintln(logF, line)
		}
		logF.Close()
	}
}
