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
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand"
	"strings"
)

// handle specified range
func GetRange(s string) (from, to float64, err error) {
	fRange := strings.SplitN(s, "-", 2)
	switch len(fRange) {
	case 1:
		var f float64
		if f, err = ParseNumber(fRange[0]); err != nil {
			return
		}
		from, to = f, f
	case 2:
		if from, err = ParseNumber(fRange[0]); err != nil {
			return
		}
		if to, err = ParseNumber(fRange[1]); err != nil {
			return
		}
	default:
		err = fmt.Errorf("can't handle range '%s'", s)
	}
	return
}

// GetFrequencyRange parses band limits
func GetFrequencyRange(s string) (freq, sw int64, err error) {
	var from, to float64
	if from, to, err = GetRange(s); err == nil {
		freq = int64((from + to) / 2)
		sw = int64(to) - freq
	}
	return
}

//----------------------------------------------------------------------

// Randomizer initialized with seed for deterministic randomization.
func Randomizer(seed int64) *rand.Rand {
	hsh := sha256.New()
	hsh.Write([]byte(fmt.Sprintf(">Y< seed %d", seed)))
	rdr := bytes.NewReader(hsh.Sum(nil))
	v, _ := binary.ReadVarint(rdr)
	return rand.New(rand.NewSource(v))
}

//----------------------------------------------------------------------

// timespan units in ascending order
var timespans = []struct {
	num  int64
	symb rune
}{{60, 's'}, {60, 'm'}, {24, 'h'}, {365, 'd'}, {-1, 'y'}}

// FormatDuration for number of seconds
func FormatDuration(v int64) string {
	out := ""
	var r int64
	for idx := 0; v != 0; idx++ {
		d := timespans[idx].num
		if d < 0 {
			r, v = v, 0
		} else {
			r = v % d
			v /= d
		}
		out = fmt.Sprintf("%d%c ", r, timespans[idx].symb) + out
	}
	return strings.TrimRight(out, " ")
}
