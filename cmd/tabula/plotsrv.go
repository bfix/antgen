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
	"embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"text/template"
	"time"

	"github.com/bfix/antgen/lib"
)

//go:embed gui.htpl
var fsys embed.FS

// shared variables with request handlers.
// N.B.: database changes after application start may not be accessable.
var (
	tpl    *template.Template      // HTML templates
	srv    *http.Server            // HTTP server
	prefix string                  // URL prefix (if behind reverse proxy)
	sets   map[string]*lib.PlotSet // list of available plot sets
)

// application entry point
func plotsrv(db *lib.Database, _ string, args []string) {
	// handle command-line arguments
	var (
		listen string // HTTP server listen
		prefix string // HTTP URL prefix
		err    error
	)
	fs := flag.NewFlagSet("srv", flag.ContinueOnError)
	fs.StringVar(&listen, "l", "localhost:12345", "Listen address for web GUI")
	fs.StringVar(&prefix, "p", "", "URL prefix")
	fs.Parse(args)

	// normalize prefix (no trailing slash)
	prefix = strings.TrimRight(prefix, "/")

	// collect sets from database
	if sets, err = db.ListPlotSets(); err != nil {
		log.Fatal("list sets: " + err.Error())
	}
	// read and prepare templates
	tpl = template.New("gui")
	tpl.Funcs(template.FuncMap{
		"msgClass": func(mode string) string {
			cl := "stat-"
			switch mode {
			case "ERROR":
				cl += "err"
			case "WARN":
				cl += "warn"
			case "INFO":
				cl += "info"
			default:
				cl += "norm"
			}
			return cl
		},
		// https://stackoverflow.com/questions/18276173/calling-a-template-with-several-pipeline-parameters
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, errors.New("invalid dict call")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, errors.New("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
		"parRange": func(key string, ps *lib.PlotSet) string {
			var list []float64
			switch key {
			case "k":
				list = ps.Klist
			case "param":
				list = ps.Plist
			default:
				return "n/a"
			}
			trim := func(v float64) string {
				s := strconv.FormatFloat(v, 'f', 6, 64)
				return strings.TrimRight(s, "0.")
			}
			n := len(list)
			return fmt.Sprintf("%s - %s", trim(list[0]), trim(list[n-1]))
		},
	})
	if _, err := tpl.ParseFS(fsys, "gui.htpl"); err != nil {
		log.Fatal("tpl: " + err.Error())
	}

	// define request handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/", plotHandler)

	// prepare HTTP server
	srv = &http.Server{
		Addr:              listen,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       300 * time.Second,
		ReadHeaderTimeout: 20 * time.Second,
		Handler:           mux,
	}
	// run HTTP server in go-routine
	go func() {
		log.Printf("Starting HTTP server at %s...", listen)
		if err := srv.ListenAndServe(); err != nil {
			log.Println("GUI listener: " + err.Error())
		}
	}()

	// handle OS signals
	sigCh := make(chan os.Signal, 5)
	signal.Notify(sigCh)
	for sig := range sigCh {
		switch sig {
		case syscall.SIGKILL, syscall.SIGINT, syscall.SIGTERM:
			log.Printf("Terminating service (on signal '%s')\n", sig)
			return
		case syscall.SIGHUP:
			log.Println("SIGHUP")
		case syscall.SIGURG:
			// TODO: https://github.com/golang/go/issues/37942
		default:
			log.Println("Unhandled signal: " + sig.String())
		}
	}
}
