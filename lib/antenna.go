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
	"io"

	necpp "github.com/ctdk/go-libnecpp"
)

// Antenna geometry, parameter and performance
type Antenna struct {
	kind   string       // kind of antenna
	segs   []*Line      // antenna geometry
	dia    float64      // constant wire diameter
	excite int          // position of exitation segment
	Lambda float64      // wavelength at operating frequency
	Perf   *Performance // antenna performance
}

// NewAntenna instantiates a new kind of antenna
func NewAntenna(kind string) *Antenna {
	return &Antenna{
		kind: kind,
		segs: make([]*Line, 0),
		Perf: new(Performance),
	}
}

// BuildAntenna from given geometry
func BuildAntenna(kind string, spec *Specification, nodes []*Node) (ant *Antenna) {
	ant = NewAntenna(kind)
	ant.Lambda = spec.Source.Lambda()
	ant.dia = spec.Wire.Diameter
	d := spec.Feedpt.Gap
	if IsNull(d) {
		d = nodes[0].Length
		spec.Feedpt.Gap = d
	}
	pos := NewVec3(d/2, 0, spec.Ground.Height)
	if ext := spec.Feedpt.Extension; ext > 0.001 {
		posE := pos
		posE[2] = -ext
		ant.Add(NewLine(posE.MirrorX(), posE))
		ant.Add(NewLine(posE, pos))
		ant.Add(NewLine(posE.MirrorX(), pos.MirrorX()))
	} else {
		ant.Add(NewLine(pos.MirrorX(), pos))
	}

	ant.excite = 0
	dir := 0.
	for _, node := range nodes {
		dir += node.Theta
		end := pos.Move2D(node.Length, dir)
		ant.Add(NewLine(pos, end))
		ant.Add(NewLine(end.MirrorX(), pos.MirrorX()))
		pos = end
	}
	ant.FixGeometry(2 * nodes[0].Length)
	return
}

// Type of antenna
func (a *Antenna) Type() string {
	return a.kind
}

// SetExcitation places the feed point on a wire segment
func (a *Antenna) SetExcitation(pos int) {
	a.excite = pos
}

// Add segment to antenna geometry
func (a *Antenna) Add(s *Line) {
	a.segs = append(a.segs, s)
}

// Eval antenna performance at given frequency
func (a *Antenna) Eval(freq int64, wire Wire, ground Ground) (err error) {
	// allocate NEC2 context
	var ctx *necpp.NecppCtx
	if ctx, err = necpp.New(); err != nil {
		return
	}
	defer ctx.Delete()

	// build antenna wire segments
	a.Lambda = C / float64(freq)
	dx := a.Lambda / 100
	for i, seg := range a.segs {
		k := max(1, min(100, int(seg.Length()/dx)))
		start, end := seg.Start(), seg.End()
		if err = ctx.Wire(i+1, k, start[0], start[1], start[2], end[0], end[1], end[2], a.dia/2, 1, 1); err != nil {
			return
		}
	}
	if err = ctx.GeometryComplete(necpp.GeoGroundPlaneFlag(ground.Mode)); err != nil {
		return
	}
	// set ground parameters
	if ground.Mode != 0 {
		if err = ctx.GnCard(necpp.GroundTypeFlag(ground.Type), ground.NRadl, ground.Epse, ground.Sig, 0, 0, 0, 0); err != nil {
			return
		}
	}
	// set material for all segments
	if !IsNull(wire.Conductivity) {
		if err = ctx.LdCard(5, 0, 0, 0, wire.Conductivity, 0, 0); err != nil {
			return
		}
	}
	if !IsNull(wire.Inductance) {
		if err = ctx.LdCard(2, 0, 0, 0, 0, wire.Inductance, 0); err != nil {
			return
		}
	}
	// specify evaluation parameters
	if err = ctx.FrCard(necpp.Linear, 1, float64(freq)/1e6, 0); err != nil {
		return
	}
	if err = ctx.ExCard(necpp.VoltageApplied, a.excite+1, 1, 0, Cfg.Sim.ExciteU, 0, 0, 0, 0, 0); err != nil {
		return
	}

	// radiation pattern requested:
	// Θ (Theta): angle measured between the positive Z semiaxis and the
	//            ground plane XY (elevation angle: π/2 - Θ)
	// Φ (Phi):   angle measured between the positive X semiaxis and the
	//            YZ plane (azimuth = π/2 - Φ)
	nTheta := int(180./Cfg.Sim.ThetaStep) + 1
	nPhi := int(360./Cfg.Sim.PhiStep) + 1
	if err = ctx.RpCard(necpp.Normal, nTheta, nPhi, necpp.MajorMinor, necpp.TotalNormalized,
		necpp.PowerGain, necpp.NoAvg, 0, 0, Cfg.Sim.ThetaStep, Cfg.Sim.PhiStep, 0, 0); err != nil {
		return
	}

	// get simulated preformance result
	a.Perf.Gain = new(Gain)
	if a.Perf.Gain.Max, err = ctx.GainMax(0); err != nil {
		return
	}
	if a.Perf.Gain.Mean, err = ctx.GainMean(0); err != nil {
		return
	}
	if a.Perf.Gain.SD, err = ctx.GainSd(0); err != nil {
		return
	}
	if a.Perf.Z, err = ctx.Impedance(0); err != nil {
		return
	}

	// get radiation pattern
	a.Perf.Rp = new(RadPattern)
	a.Perf.Rp.Max, a.Perf.Rp.Min = 0, 100
	a.Perf.Rp.NPhi = nPhi
	a.Perf.Rp.NTheta = nTheta
	a.Perf.Rp.Values = make([][]float64, nTheta)
	for i := range nTheta {
		a.Perf.Rp.Values[i] = make([]float64, nPhi)
	}
	var val float64
	for theta := range nTheta {
		for phi := range nPhi {
			if val, err = ctx.Gain(0, theta, phi); err != nil {
				return
			}
			a.Perf.Rp.Max = max(a.Perf.Rp.Max, val)
			a.Perf.Rp.Min = min(a.Perf.Rp.Min, val)
			a.Perf.Rp.Values[theta][phi] = val
		}
	}
	return
}

// Bulge specifies the number of segments involved in avoiding
// wire intersections.
const Bulge = 100

// FixGeometry makes sure that an antenna geometry can be used for simulations
// (e.g. avoiding wire intersections by "bridging" wire crossings).
func (a *Antenna) FixGeometry(minD float64) {
	probs := CheckDistances(a.segs, minD)
	regions := Regions(probs)
	for _, r := range regions {
		start, end := max(1, r[0]-Bulge), min(len(a.segs)-1, r[1]+Bulge)
		step := minD / float64(max(r[0]-start, end-r[1]))
		for i := start; i <= end; i++ {
			n := a.segs[i]
			if i < r[0] {
				dz := minD - float64(r[0]-i)*step
				n.start[2] += dz
				n.end[2] += dz + step
			} else if i > r[1] {
				dz := minD - float64(i-r[1])*step
				n.start[2] += dz + step
				n.end[2] += dz
			} else {
				n.start[2] += minD
				n.end[2] += minD
			}
		}
	}
}

// DumpNEC writes an antenna simulation card deck to writer.
func (a *Antenna) DumpNEC(wrt io.Writer, spec *Specification, comments []string) {
	for _, cmt := range comments {
		fmt.Fprintf(wrt, "CM %s\n", cmt)
	}
	fmt.Fprintln(wrt, "CE Model output")
	for i, s := range a.segs {
		l := s.end.Add(s.start.Neg()).Length()
		n := int(min(100, max(1, l/0.01)))
		fmt.Fprintf(wrt, "GW %d %d %e %e %e %e %e %e %e\n", i+1, n,
			s.start[0], s.start[1], s.start[2],
			s.end[0], s.end[1], s.end[2],
			a.dia/2,
		)
	}
	volt := 1. // math.Sqrt(spec.FeedP * real(spec.FeedZ))

	fmt.Fprintf(wrt, "GE %d\n", spec.Ground.Mode)
	if !IsNull(spec.Wire.Inductance) {
		fmt.Fprintf(wrt, "LD 2 0 0 0 0 %e 0\n", spec.Wire.Inductance)
	}
	if !IsNull(spec.Wire.Conductivity) {
		fmt.Fprintf(wrt, "LD 5 0 0 0 %e\n", spec.Wire.Conductivity)
	}
	fmt.Fprintf(wrt, "EX 0 %d 1 0 %f\n", a.excite+1, volt)
	f := float64(spec.Source.Freq) / 1e6
	if spec.Source.Span > 0 {
		fh := float64(spec.Source.Span) / 1e6
		fmt.Fprintf(wrt, "FR 0 101 0 0 %f %f\n", f-fh, 2*fh/100)
	} else {
		fmt.Fprintf(wrt, "FR 0 1 0 0 %f 0\n", f)
	}
	fmt.Fprintln(wrt, "RP 0 37 73 1000 0 0 5 5 0 0")
	fmt.Fprintln(wrt, "EN")
}
