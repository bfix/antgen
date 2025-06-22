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
	"math"
	"sort"
)

// Geometry of 2D-bended antenna
type Geometry struct {
	Cmts   []string  `json:"comments"` // optimization info/comments
	SegL   float64   `json:"segL"`     // segment length in wings
	Num    int       `json:"num"`      // number of segments
	Wire   Wire      `json:"wire"`     // wire parameters
	Height float64   `json:"height"`   // height of antenna
	Bends  []float64 `json:"bends"`    // list of 2D bends
}

//----------------------------------------------------------------------

func Smooth2D(nodes []Node, rng int) (out []Node) {
	num := len(nodes)
	out = make([]Node, num)
	for i := range out {
		l, _ := nodes[i].Polar()
		out[i] = NewNode2D(l, 0)
	}
	for i, n := range nodes {
		_, ang := n.Polar()
		s, e := -rng, rng
		if i+s < 0 {
			s = -i
		}
		if i+e > num-1 {
			e = num - 1 - i
		}
		f := 0.
		for j := s; j <= e; j++ {
			f += 1 / math.Exp2(math.Abs(float64(j)))
		}
		ang /= f
		for j := s; j <= e; j++ {
			nj := out[i+j].(*Node2D)
			nj.angle += ang / math.Exp2(math.Abs(float64(j)))
		}
	}
	return
}

//----------------------------------------------------------------------

type BoundingBox struct {
	Xmin, Xmax float64
	Ymin, Ymax float64
	Zmin, Zmax float64
}

func NewBoundingBox() *BoundingBox {
	limit := math.MaxFloat32
	return &BoundingBox{
		Xmin: limit,
		Xmax: -limit,
		Ymin: limit,
		Ymax: -limit,
		Zmin: limit,
		Zmax: -limit,
	}
}

func (b *BoundingBox) Include(v Vec3) {
	b.Xmin = min(v[0], b.Xmin)
	b.Xmax = max(v[0], b.Xmax)
	b.Ymin = min(v[1], b.Ymin)
	b.Ymax = max(v[1], b.Ymax)
	b.Zmin = min(v[2], b.Zmin)
	b.Zmax = max(v[2], b.Zmax)
}

//----------------------------------------------------------------------

// Node interface for 3D line/segments
type Node interface {

	// Dir returns the direction of the node as vector
	Dir() Vec3

	// Len returns the length of the segment
	Len() float64

	// Polar coordinates (2D) of the node
	Polar() (r float64, ang float64)
}

// Node2D impements a 2D node
type Node2D struct {
	length float64 // length of segment
	angle  float64 // angle in XY plane
}

// NewNode2D creates a new 2D node
func NewNode2D(l, a float64) (n *Node2D) {
	return &Node2D{
		length: l,
		angle:  a,
	}
}

// Len returns the length of the segment
func (n *Node2D) Len() float64 {
	return n.length
}

// Dir returns the direction of the node as vector
func (n *Node2D) Dir() (v Vec3) {
	v[0] = math.Cos(n.angle)
	v[1] = math.Sin(n.angle)
	v[2] = 0
	return
}

// Polar coordinates (2D) of the node
func (n *Node2D) Polar() (float64, float64) {
	return n.length, n.angle
}

// SetAngle of a node
func (n *Node2D) SetAngle(a float64) {
	n.angle = a
}

// AddAngle to the current direction of a node
func (n *Node2D) AddAngle(da float64) {
	n.angle += da
}

// GetAngle returns the current node angle
func (n *Node2D) GetAngle() float64 {
	return n.angle
}

//----------------------------------------------------------------------

// Vec3 is a 3D vector
type Vec3 [3]float64

// NewVec3 creates a new 3D vector
func NewVec3(x, y, z float64) (v Vec3) {
	v[0], v[1], v[2] = x, y, z
	return
}

// String returns a human-readable vector
func (v Vec3) String() string {
	return fmt.Sprintf("[%f,%f,%f]", v[0], v[1], v[2])
}

// Length of the vector
func (v Vec3) Length() float64 {
	x, y, z := v[0], v[1], v[2]
	return math.Sqrt(x*x + y*y + z*z)
}

// Norm returns a normalized vector
func (v Vec3) Norm() (u Vec3) {
	l := v.Length()
	return v.Mult(1 / l)
}

// Add two vectors
func (v Vec3) Add(u Vec3) (d Vec3) {
	d[0] = v[0] + u[0]
	d[1] = v[1] + u[1]
	d[2] = v[2] + u[2]
	return
}

// Sub (substract) two vectors
func (v Vec3) Sub(u Vec3) (d Vec3) {
	d[0] = v[0] - u[0]
	d[1] = v[1] - u[1]
	d[2] = v[2] - u[2]
	return
}

// Mult returns the multiplication of a vector with a scalar k
func (v Vec3) Mult(k float64) (d Vec3) {
	d[0] = k * v[0]
	d[1] = k * v[1]
	d[2] = k * v[2]
	return
}

// Neg returns the negative vector
func (v Vec3) Neg() (d Vec3) {
	d[0] = -v[0]
	d[1] = -v[1]
	d[2] = -v[2]
	return
}

// Prod returns the cross product between two vectors
func (v Vec3) Prod(u Vec3) (d Vec3) {
	d[0] = v[1]*u[2] - v[2]*u[1]
	d[1] = v[2]*u[0] - v[0]*u[2]
	d[2] = v[0]*u[1] - v[1]*u[0]
	return
}

// Dot returns the dot product between two vectors
func (v Vec3) Dot(u Vec3) float64 {
	return v[0]*u[0] + v[1]*u[1] + v[2]*u[2]
}

// Equals returns true if two vectors are equal
func (v Vec3) Equals(u Vec3) bool {
	return IsNull(v.Sub(u).Length())
}

// Move2D moves a vector in the XY plane
func (v Vec3) Move2D(r, a float64) (w Vec3) {
	w[0] += v[0] + r*math.Cos(a)
	w[1] += v[1] + r*math.Sin(a)
	w[2] = v[2]
	return
}

// MirrorX mirrors the vector (YZ plane)
func (v Vec3) MirrorX() (w Vec3) {
	w[0] = -v[0]
	w[1] = v[1]
	w[2] = v[2]
	return
}

//----------------------------------------------------------------------

// Line interface
type Line interface {
	// Start point of line (3D)
	Start() Vec3

	// End point of line (3D)
	End() Vec3

	// Length of line
	Length() float64

	// Dir is the direction of the line in 3D
	Dir() Vec3

	// Distance between two lines
	Distance(Line) float64

	// Intersect returns true (and the intersection point)
	// if two lines intersect
	Intersect(Line) (Vec3, bool)

	// String returns the human-readable line
	String() string
}

// Line3 is a 3D implementation of the Line interface
type Line3 struct {
	start Vec3
	end   Vec3
}

// NewLine3 creates a new 3D line
func NewLine3(s, e Vec3) *Line3 {
	return &Line3{
		start: s,
		end:   e,
	}
}

// Start point of line (3D)
func (l *Line3) Start() Vec3 {
	return l.start
}

// End point of line (3D)
func (l *Line3) End() Vec3 {
	return l.end
}

// Length of line
func (l *Line3) Length() float64 {
	return l.Dir().Length()
}

// Dir is the direction of the line in 3D
func (l *Line3) Dir() Vec3 {
	return l.end.Sub(l.start)
}

// String returns the human-readable line
func (l *Line3) String() string {
	return fmt.Sprintf("(%f,%f,%f)-(%f,%f,%f)",
		l.start[0], l.start[1], l.start[2],
		l.end[0], l.end[1], l.end[2],
	)
}

// Distance between two lines
func (li *Line3) Distance(lj Line) (d float64) {
	d = math.MaxFloat64
	if li.Start().Equals(lj.End()) || li.End().Equals(lj.Start()) {
		return
	}
	ei := li.Dir()
	ej := lj.Dir()
	mi := li.Start().Add(ei.Mult(0.5))
	mj := lj.Start().Add(ej.Mult(0.5))
	if mi.Sub(mj).Length() > (ei.Length()+ej.Length())/2 {
		return
	}
	n := ei.Prod(ej)
	na := n.Length()
	if !IsNull(na) {
		g := lj.Start().Sub(li.Start())
		d = n.Dot(g.Neg()) / na
	}
	return
}

// Intersect returns true (and the intersection point) if two lines intersect
func (li *Line3) Intersect(lj Line) (p Vec3, cross bool) {
	var pt [4]Vec3
	d := func(m, n, o, p int) float64 {
		return (pt[m-1][0]-pt[n-1][0])*(pt[o-1][0]-pt[p-1][0]) +
			(pt[m-1][1]-pt[n-1][1])*(pt[o-1][1]-pt[p-1][1]) +
			(pt[m-1][2]-pt[n-1][2])*(pt[o-1][2]-pt[p-1][2])
	}
	pt[0] = li.Start()
	pt[1] = li.End()
	pt[2] = lj.Start()
	pt[3] = lj.End()

	t1 := (d(1, 3, 4, 3)*d(4, 3, 2, 1) - d(1, 3, 2, 1)*d(4, 3, 4, 3)) /
		(d(2, 1, 2, 1)*d(4, 3, 4, 3) - d(4, 3, 2, 1)*d(4, 3, 2, 1))
	if t1 > 0 && t1 < 1 {
		t2 := (d(1, 3, 4, 3) + t1*d(4, 3, 2, 1)) / d(4, 3, 4, 3)
		if t2 > 0 && t1 < 1 {
			p = pt[0].Add(pt[1].Sub(pt[0]).Mult(t1))
			cross = true
			return
		}
	}
	return
}

//------------------------------------------------------------

// Intersects returns a list of segment indices that intersect
// other segments in the list. Only the higher index is reported.
func Intersects(segs []Line) (pos []int) {
	n := len(segs)
	for i := 0; i < n-1; i++ {
		for j := i + 1; j < n; j++ {
			if _, cross := segs[i].Intersect(segs[j]); cross {
				pos = append(pos, j)
			}
		}
	}
	return
}

// CheckDistances returns a list of segment indices where the
// smallest distance of segment to other segments in the list
// is below a given minimum. Only the higher index is reported.
func CheckDistances(segs []Line, minD float64) (pos []int) {
	n := len(segs)
	for i := 0; i < n-1; i++ {
		for j := i + 1; j < n; j++ {
			if d := segs[i].Distance(segs[j]); d < minD {
				if (j - i) > 10 {
					pos = append(pos, j)
				}
			}
		}
	}
	return
}

// Regions condenses a list of indices into regions.
// The list "3 5 6 7 8 12 15 16 19" would be returned
// as "[3,3] [5,8] [12,12] [15,16] [19,19]"
func Regions(pos []int) (r [][2]int) {
	if len(pos) == 0 {
		return
	}
	sort.Slice(pos, func(i, j int) bool {
		return pos[i] < pos[j]
	})
	start, last := pos[0], pos[0]
	for _, idx := range pos[1:] {
		if idx == last {
			continue
		}
		if idx == last+1 {
			last = idx
			continue
		}
		r = append(r, [2]int{start, last})
		start, last = idx, idx
	}
	if last != start {
		r = append(r, [2]int{start, last})
	}
	return
}
