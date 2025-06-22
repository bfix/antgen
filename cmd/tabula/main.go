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
	"log"
	"os"

	"github.com/bfix/antgen/lib"
)

// shared variables with request handlers.
// N.B.: database changes after application start may not be accessable.
var (
	db *lib.Database // reference to (opened) database
)

// application entry point
func main() {
	// handle command-line arguments
	args := os.Args[1:]
	var dbName, in string
	fs := flag.NewFlagSet("main", flag.ContinueOnError)
	fs.StringVar(&dbName, "db", "./out/results.db", "result database")
	fs.StringVar(&in, "in", "./out", "model base directory")
	fs.Parse(args)
	args = fs.Args()

	// open database
	if len(dbName) == 0 {
		flag.Usage()
		log.Fatal("no database specified")
	}
	var err error
	if db, err = lib.OpenDatabase(dbName); err != nil {
		log.Fatal("open db: " + err.Error())
	}
	defer db.Close()

	// execute command
	switch args[0] {
	case "import":
		importFromDirectory(db, in, args[1:])
	case "plot-srv":
		plotsrv(db, in, args[1:])
	case "plot-file":
		plotToFile(db, in, args[1:])
	case "show-best":
		showBest(db, in, args[1:])
	case "stats":
		stats := db.Stats()
		log.Println("Database statistics:")
		log.Printf("       Number of antennas: %10d", stats.NumAnt)
		log.Printf("  Number of optimizations: %10d", stats.NumSteps)
		log.Printf("    Number of simulations: %10d", stats.NumSims)
		log.Printf("             Elapsed time: %s", stats.Duration)
	}
}
