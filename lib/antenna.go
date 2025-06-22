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
	"log"
	"os"

	necpp "github.com/ctdk/go-libnecpp"
)

// Antenna geometry, parameter and performance
type Antenna struct {
	kind   string      // kind of antenna
	segs   []Line      // antenna geometry (build from segments)
	excite int         // position of exitation segment
	Lambda float64     // wavelength at operating frequency
	Perf   Performance // antenna performance
}

// NewAntenna instantiates a new kind of antenna
func NewAntenna(kind string) *Antenna {
	return &Antenna{
		kind: kind,
		segs: make([]Line, 0),
	}
}

// BuildAntenna from given geometry
func BuildAntenna(kind string, spec *Specification, nodes []Node) (ant *Antenna) {
	ant = NewAntenna(kind)
	ant.Lambda = spec.Source.Lambda()
	d := nodes[0].Len()
	pos := NewVec3(d/2, 0, spec.Ground.Height)
	dir := 0.
	ant.Add(NewSegment(pos.MirrorX(), pos, spec.Wire.Diameter))
	ant.excite = 0
	for _, node := range nodes {
		length, angle := node.Polar()
		dir += angle
		end := pos.Move2D(length, dir)
		ant.Add(NewSegment(pos, end, spec.Wire.Diameter))
		ant.Add(NewSegment(end.MirrorX(), pos.MirrorX(), spec.Wire.Diameter))
		pos = end
	}
	ant.FixGeometry(2 * d)
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
func (a *Antenna) Add(s *Segment) {
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
		s := seg.(*Segment)
		start, end := seg.Start(), seg.End()
		if err = ctx.Wire(i+1, k, start[0], start[1], start[2], end[0], end[1], end[2], s.dia/2, 1, 1); err != nil {
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
	if err = ctx.ExCard(necpp.VoltageApplied, a.excite, 1, 0, Cfg.Sim.ExciteU, 0, 0, 0, 0, 0); err != nil {
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
			n := a.segs[i].(*Segment)
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

// DumpNEC writes an antenna simulation card deck to file.
func (a *Antenna) DumpNEC(spec *Specification, comments []string, fName string) {
	wrt, err := os.Create(fName)
	if err != nil {
		log.Fatal(err)
	}
	defer wrt.Close()

	for _, cmt := range comments {
		fmt.Fprintf(wrt, "CM %s\n", cmt)
	}
	fmt.Fprintln(wrt, "CE Model output")
	for i, seg := range a.segs {
		s := seg.(*Segment)
		l := s.end.Add(s.start.Neg()).Length()
		n := int(min(100, max(1, l/0.01)))
		fmt.Fprintf(wrt, "GW %d %d %e %e %e %e %e %e %e\n", i+1, n,
			s.start[0], s.start[1], s.start[2],
			s.end[0], s.end[1], s.end[2],
			s.dia/2,
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
	fmt.Fprintf(wrt, "EX 0 %d 1 0 %f\n", a.excite, volt)
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

//----------------------------------------------------------------------

// Segment in an antenna geometry (straight piece of wire)
// Implements the Line interface.
type Segment struct {
	start Vec3    // start position
	end   Vec3    // end position
	dia   float64 // wire diameter
}

// NewSegment from given parameters
func NewSegment(s, e Vec3, d float64) *Segment {
	return &Segment{
		start: s,
		end:   e,
		dia:   d,
	}
}

// Start position of wire
func (s *Segment) Start() Vec3 {
	return s.start
}

// End position of wire
func (s *Segment) End() Vec3 {
	return s.end
}

// String returns a human-readable segment text
func (s *Segment) String() string {
	return fmt.Sprintf("{%s-%s, %f}", s.start, s.end, s.dia)
}

// Length of segment
func (s *Segment) Length() float64 {
	return s.Dir().Length()
}

// Dir of segment (absolute direction)
func (s *Segment) Dir() Vec3 {
	return s.end.Add(s.start.Neg())
}

// Distance between two segments
func (s *Segment) Distance(sj Line) (d float64) {
	li := NewLine3(s.start, s.end)
	lj := NewLine3(sj.Start(), sj.End())
	return li.Distance(lj)
}

// Intersect checks if two segments intersect and returns
// the intersection point if they do.
func (s *Segment) Intersect(sj Line) (p Vec3, cross bool) {
	li := NewLine3(s.start, s.end)
	lj := NewLine3(sj.Start(), sj.End())
	return li.Intersect(lj)
}
