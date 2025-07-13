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
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// GenMdlParams assembles model parameters as list of strings.
// The output is parsable with ParseMdlParams().
func GenMdlParams(
	param float64,
	spec *Specification,
	ini, perf *Performance,
	mdl, gen, opt string,
	seed int64,
	tag string,
	total Stats,
) (cmts []string) {

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
	cmts = append(cmts, ">>>>> Feedpoint: gap:extension")
	cmt = fmt.Sprintf("Feedpoint: %.3f:%.3f", spec.Feedpt.Gap, spec.Feedpt.Extension)
	cmts = append(cmts, cmt)
	cmts = append(cmts, ">>>>> Ground: height:mode:type:nradl:epse:sig")
	cmt = fmt.Sprintf("Ground: %.3f:%d:%d:%d:%f:%f",
		spec.Ground.Height, spec.Ground.Mode, spec.Ground.Type,
		spec.Ground.NRadl, spec.Ground.Epse, spec.Ground.Sig,
	)
	cmts = append(cmts, cmt)

	// model parameters
	cmts = append(cmts, ">>>>> Param: k:param:tag")
	ps := ""
	if !math.IsNaN(param) {
		ps = fmt.Sprintf("%f", param)
	}
	cmt = fmt.Sprintf("Param: %f:%s:%s", spec.K, ps, tag)
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

// ParseMdlParams from model file (extract performance parameters)
func ParseMdlParams(cmts []string) (p *Record, ok bool, err error) {
	p = new(Record)
	found := 0
	for _, line := range cmts {
		kind, vals := SplitParam(line)
		switch kind {

		// >>>>> Source: freq:Zr:Zi
		case "Source":
			if p.Freq, err = strconv.ParseInt(vals[0], 10, 64); err != nil {
				return
			}
			found++

		// >>>>> Wire: dia:material:conductivity:inductance
		case "Wire":
			if p.Wire.Diameter, err = strconv.ParseFloat(vals[0], 64); err != nil {
				return
			}
			p.Wire.Material = vals[1]
			if p.Wire.Conductivity, err = strconv.ParseFloat(vals[2], 64); err != nil {
				return
			}
			if p.Wire.Inductance, err = strconv.ParseFloat(vals[3], 64); err != nil {
				return
			}
			found++

		// >>>>> Feedpoint: gap:extension
		case "Feedpoint":
			if p.Feedpt.Gap, err = strconv.ParseFloat(vals[0], 64); err != nil {
				return
			}
			if p.Feedpt.Extension, err = strconv.ParseFloat(vals[1], 64); err != nil {
				return
			}
			found++

		// >>>>> Ground: height:mode:type:...
		case "Ground":
			if p.Gnd.Height, err = strconv.ParseFloat(vals[0], 64); err != nil {
				return
			}
			if p.Gnd.Mode, err = strconv.Atoi(vals[1]); err != nil {
				return
			}
			if p.Gnd.Type, err = strconv.Atoi(vals[2]); err != nil {
				return
			}
			found++

		// >>>>> Param: k:param:tag
		case "Param":
			if p.K, err = strconv.ParseFloat(vals[0], 64); err != nil {
				return
			}
			p.Param = math.NaN()
			if len(vals[1]) > 0 {
				if p.Param, err = strconv.ParseFloat(vals[1], 64); err != nil {
					return
				}
			}
			p.Tag = vals[2]
			found++

		// >>>>> Mode: model:generator:seed:optimizer
		case "Mode":
			p.Mdl = vals[0]
			p.Gen = vals[1]
			if p.Seed, err = strconv.ParseInt(vals[2], 10, 64); err != nil {
				return
			}
			p.Opt = vals[3]
			found++

		// >>>>> Result: Gmax:Gmean:SD:Zr:Zi
		case "Result":
			p.Perf.Gain = new(Gain)

			if p.Perf.Gain.Max, err = strconv.ParseFloat(vals[0], 64); err != nil {
				return
			}
			if p.Perf.Gain.Mean, err = strconv.ParseFloat(vals[1], 64); err != nil {
				return
			}
			if p.Perf.Gain.SD, err = strconv.ParseFloat(vals[2], 64); err != nil {
				return
			}
			var Zr, Zi float64
			if Zr, err = strconv.ParseFloat(vals[3], 64); err != nil {
				return
			}
			if Zi, err = strconv.ParseFloat(vals[4], 64); err != nil {
				return
			}
			p.Perf.Z = complex(Zr, Zi)
			found++

		// >>>>> Stats: mthds:steps:sims:elapsed
		case "Stats":
			if p.Stats.NumMthds, err = strconv.Atoi(vals[0]); err != nil {
				return
			}
			if p.Stats.NumSteps, err = strconv.Atoi(vals[1]); err != nil {
				return
			}
			if p.Stats.NumSims, err = strconv.Atoi(vals[2]); err != nil {
				return
			}
			var t int
			if t, err = strconv.Atoi(vals[3]); err != nil {
				return
			}
			p.Stats.Elapsed = time.Duration(t) * time.Second
			found++
		}
	}
	ok = (found > 0)
	return
}

// ParseMdlParamsFromNEC retrieves model parameters from a NEC2 model file
func ParseMdlParamsFromNEC(fName, dirIn string) (p *Record, ok bool, err error) {
	var fIn *os.File
	if fIn, err = os.Open(fName); err != nil {
		return
	}
	defer fIn.Close()

	var cmts []string
	rdr := bufio.NewReader(fIn)
	var buf []byte
	for {
		if buf, _, err = rdr.ReadLine(); err != nil {
			if err == io.EOF {
				err = nil
				break
			}
			return
		}
		line := string(buf)
		if len(line) > 2 && line[:3] == "CM " {
			cmts = append(cmts, line[3:])
		}
	}
	p, ok, err = ParseMdlParams(cmts)
	if p != nil {
		p.Path = strings.ReplaceAll(filepath.Dir(fName), dirIn+"/", "")
	}
	return
}

// SplitParam dissects a parameter string
func SplitParam(line string) (kind string, vals []string) {
	if strings.HasPrefix(line, "CM ") {
		line = line[:3]
	}
	idx := strings.IndexRune(line, ':')
	if idx == -1 {
		return
	}
	kind = line[:idx]
	vals = strings.Split(line[idx+2:], ":")
	return
}
