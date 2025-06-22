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

const (
	TRK_MARK   = -1
	TRK_SHORT  = -2
	TRK_LENGTH = -3
)

type Change struct {
	Pos   int     `json:"pos"`
	Angle float64 `json:"angle"`
}

func Changes(nodes []Node) []*Change {
	changes := make([]*Change, 0)
	for i, n := range nodes {
		node := n.(*Node2D)
		if !IsNull(node.angle) {
			changes = append(changes, &Change{Pos: i, Angle: node.angle})
		}
	}
	return changes
}

type TrackList struct {
	Cmts   []string  `json:"comments"`
	SegL   float64   `json:"segL"`
	Num    int       `json:"num"`
	Wire   Wire      `json:"wire"`
	Height float64   `json:"height"`
	Track  []*Change `json:"track"`
}

func (tl *TrackList) Nodes() []Node {

	// build initial geometry
	nodes := make([]Node, tl.Num)
	for i := range nodes {
		nodes[i] = NewNode2D(tl.SegL, 0)
	}

	// iterate over changes
	for _, chg := range tl.Track {
		if chg.Pos == -1 {
			continue
		}
		// apply change
		n := nodes[chg.Pos].(*Node2D)
		n.angle += chg.Angle
	}
	return nodes
}
