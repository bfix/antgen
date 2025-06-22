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
	"flag"
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"github.com/bfix/antgen/lib"
)

// import performance data from model files
func importFromDirectory(db *lib.Database, in string, args []string) {
	// handle command-line arguments
	var (
		set string // only import set with given prefix
	)
	fls := flag.NewFlagSet("import", flag.ContinueOnError)
	fls.StringVar(&set, "set", "", "set prefix")
	fls.Parse(args)

	// traverse directory and import model files
	num := 0
	if err := filepath.Walk(in, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".nec") {
			if len(set) > 0 && !strings.HasPrefix(path, in+"/"+set) {
				return nil
			}
			log.Printf(">>> %s", path)

			// extract information from model file
			p, ok, err := lib.ParseMdlParams(path, in)
			if err != nil {
				log.Printf("ERROR: %s", err.Error())
				return nil
			}
			if ok {
				if err = db.Insert(p); err != nil {
					log.Printf("ERROR: %s", err.Error())
					return nil
				}
				num++
			}
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}
	log.Printf("Done: %d models imported.", num)
}
