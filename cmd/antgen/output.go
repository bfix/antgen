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
	"log"
	"math"
	"os"
	"strings"

	"github.com/bfix/antgen/lib"
)

func OutputModel(
	k float64,
	param float64,
	spec *lib.Specification,
	ant *lib.Antenna,
	gen lib.Generator,
	ini lib.Performance,
	mdl lib.Model,
	opt string,
	seed int64,
	total lib.Stats,
	steps []string,
	tag string,
	outDir string,
	outPrf string,
) {

	// assemble comments
	cmts := comments(k, param, spec, ini, ant.Perf, mdl.Info(), gen.Info(), opt, seed, tag, total)

	// handle output prefix
	if len(outPrf) > 0 && !strings.HasSuffix(outPrf, "_") {
		outPrf += "_"
	}
	// get model filename
	fName := fmt.Sprintf("%s/%smodel-%s.nec", outDir, outPrf, tag)
	ant.DumpNEC(spec, cmts, fName)
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

// assemble comments
func comments(
	k float64,
	param float64,
	spec *lib.Specification,
	ini, perf lib.Performance,
	mdl, gen, opt string,
	seed int64,
	tag string,
	total lib.Stats,
) (cmts []string) {

	// intro
	cmts = append(cmts, fmt.Sprintf("AntGen %s (%s) - Copyright 2024-present Bernd Fix   >Y<", Version, Date))

	// specification (source, wire, ground)
	cmts = append(cmts, ">>>>> Source: freq:Zr:Zi")
	cmt := fmt.Sprintf("Source: %d:%f:%f",
		spec.Source.Freq, spec.Source.Z.R, spec.Source.Z.X,
	)
	cmts = append(cmts, cmt)
	cmts = append(cmts, ">>>>> Wire: dia:material:conductivity:inductance")
	cmt = fmt.Sprintf("Wire: %.3f:%s:%.3e:%.3e",
		spec.Wire.Diameter, spec.Wire.Material, spec.Wire.Conductivity, spec.Wire.Inductance,
	)
	cmts = append(cmts, cmt)
	cmts = append(cmts, ">>>>> Ground: height:mode:type:nradl:epse:sig")
	cmt = fmt.Sprintf("Ground: %.3f:%d:%d:%d:%f:%f",
		spec.Ground.Height, spec.Ground.Mode, spec.Ground.Type, spec.Ground.NRadl, spec.Ground.Epse, spec.Ground.Sig,
	)
	cmts = append(cmts, cmt)

	// model parameters
	cmts = append(cmts, ">>>>> Param: k:param:tag")
	ps := ""
	if !math.IsNaN(param) {
		ps = fmt.Sprintf("%f", param)
	}
	cmt = fmt.Sprintf("Param: %f:%s:%s", k, ps, tag)
	cmts = append(cmts, cmt)

	// optimization parameters
	cmts = append(cmts, ">>>>> Mode: model:generator:seed:optimizer")
	cmt = fmt.Sprintf("Mode: %s:%s:%d:%s", mdl, gen, seed, opt)
	cmts = append(cmts, cmt)

	// initial performance
	cmts = append(cmts, ">>>>> Init: Gmax:Gmean:SD:Zr:Zi")
	cmt = fmt.Sprintf("Init: %f:%f:%f:%f:%f",
		ini.Gain.Max, ini.Gain.Mean, ini.Gain.SD,
		real(ini.Z), imag(ini.Z),
	)
	cmts = append(cmts, cmt)

	// final performance
	cmts = append(cmts, ">>>>> Result: Gmax:Gmean:SD:Zr:Zi")
	cmt = fmt.Sprintf("Result: %f:%f:%f:%f:%f",
		perf.Gain.Max, perf.Gain.Mean, perf.Gain.SD,
		real(perf.Z), imag(perf.Z),
	)
	cmts = append(cmts, cmt)

	// statistics
	cmts = append(cmts, ">>>>> Stats: Mthds:Steps:Sims:Elapsed")
	cmt = fmt.Sprintf("Stats: %d:%d:%d:%d",
		total.NumMthds, total.NumSteps, total.NumSims, int(total.Elapsed.Seconds()),
	)
	cmts = append(cmts, cmt)

	return
}
