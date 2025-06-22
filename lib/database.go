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
	"database/sql"
	"errors"
	"fmt"
	"math"
	"math/cmplx"
	"path/filepath"
	"sort"

	_ "github.com/mattn/go-sqlite3"
)

// Index to performance record.
// Either 'k' or 'param' can be NaN (skipped), but not both.
type Index struct {
	k     float64
	param float64
}

// NewIndex from k and param
func NewIndex(k, param float64) Index {
	return Index{k, param}
}

// Match returns true if k and param match the index values.
// Handles NaN values as "match always"
func (i Index) Match(j Index) bool {
	if !math.IsNaN(i.k) && !math.IsNaN(j.k) && i.k != j.k {
		return false
	}
	if !math.IsNaN(i.param) && !math.IsNaN(j.param) && i.param != j.param {
		return false
	}
	return true
}

func (i Index) K() float64 {
	return i.k
}

func (i Index) Param() float64 {
	return i.param
}

//----------------------------------------------------------------------

// IndexList is a list of unique indices (usually over a set of data)
type IndexList struct {
	list []Index
	dup  map[string]bool
}

// NewIndexList creates an empty list of indices
func NewIndexList() *IndexList {
	return &IndexList{
		list: make([]Index, 0),
		dup:  make(map[string]bool),
	}
}

// Add index to list
func (il *IndexList) Add(i Index) {
	s := fmt.Sprintf("%f-%f", i.k, i.param)
	if _, ok := il.dup[s]; ok {
		return
	}
	il.list = append(il.list, i)
	il.dup[s] = true
}

// Sorted returns a sorted list of indices
func (il *IndexList) Sorted() (out []Index) {
	// sort indices (k primary, param secondary; ascending)
	sort.Slice(il.list, func(i, j int) bool {
		if !math.IsNaN(il.list[i].k) {
			if il.list[i].k < il.list[j].k {
				return true
			}
			if !IsNull(il.list[i].k) {
				return false
			}
		}
		return il.list[i].param < il.list[j].param
	})
	out = il.list
	return
}

//----------------------------------------------------------------------

// Row in the performance table
type Row struct {
	id    int64   // database record id
	idx   Index   // record index (k, param)
	gmax  float64 // maximum gain
	gmean float64 // mean gain
	sd    float64 // gain std. deviation
	zr    float64 // antenna resistance
	zi    float64 // antenna reactance
	fdir  string  // file path
	ftag  string  // file tag
}

// Reference to database entry and related model file
func (r *Row) Reference() (id int64, fdir, ftag string) {
	return r.id, r.fdir, r.ftag
}

// Index of record (k,param)
func (r *Row) Index() Index {
	return r.idx
}

// Value of a named performance parameter is returned
func (r *Row) Value(name string) float64 {
	// values from database
	switch name {
	case "k":
		return r.idx.k
	case "param":
		return r.idx.param
	case "Gmax":
		return r.gmax
	case "Gmean":
		return r.gmean
	case "SD":
		return r.sd
	case "Zr":
		return r.zr
	case "Zi":
		return r.zi

	// derived values
	case "Geff":
		// Gmax of a matched antenna
		z := complex(r.zr, r.zi)
		pf := real(z) / cmplx.Abs(z)
		return r.gmax + 10*math.Log10(pf)
	case "Loss":
		// Loss due to unmatched antenna
		z := complex(r.zr, r.zi)
		z0 := complex(50, 0)
		g := cmplx.Abs((z - z0) / (z + z0))
		s := (1 + g) / (1 - g)
		return 10 * math.Log10(4*s/Sqr(s+1))
	case "PwrFac":
		// Loss due to phase shift
		z := complex(r.zr, r.zi)
		pf := real(z) / cmplx.Abs(z)
		return 10 * math.Log10(pf)
	}
	return math.NaN()
}

// Record in the database
type Record struct {
	Freq  int64       // operating frequency
	Wire  Wire        // wire spec
	Gnd   Ground      // ground spec
	K     float64     // k (dipole wing length)
	Param float64     // free parameter (generator)
	Perf  Performance // final performance
	Mdl   string      // antenna model
	Gen   string      // antenna generator (initial geometry)
	Opt   string      // optimizer
	Seed  int64       // random seed
	Stats Stats       // optimization stats
	Path  string      // relative path
	Tag   string      // model tag
}

//----------------------------------------------------------------------

// Set of performance data from one run; either 'k' or 'param' (or both!) are
// varied across a defined span. The set is identified by 'fdir' (optimization
// results are grouped in distinct directories).
type Set struct {
	kVar bool          // varying 'k'
	pVar bool          // varying 'param'
	vals Index         // constant 'k','param'
	data []*Row        // rows in set
	idx  map[Index]int // row index
}

// NewSet returns a new (empty) set
func NewSet() *Set {
	return &Set{
		kVar: false,
		pVar: false,
		vals: Index{0, 0},
		data: make([]*Row, 0),
		idx:  make(map[Index]int),
	}
}

// Add row to set
func (s *Set) Add(r *Row) {
	idx := r.Index()
	sn := len(s.data)
	if sn == 0 {
		s.vals = idx
	}
	s.data = append(s.data, r)
	s.idx[idx] = sn

	if !math.IsNaN(idx.k) && !math.IsNaN(s.vals.k) && idx.k != s.vals.k {
		s.kVar = true
	}
	if !math.IsNaN(idx.param) && !math.IsNaN(s.vals.param) && idx.param != s.vals.param {
		s.pVar = true
	}
}

// Varying flags
const (
	VaryK = 1
	VaryP = 2
)

// Varying returns the var. kind and adds indices to the sweep list
func (s *Set) Varying(sweep *IndexList) (f int) {
	if s.kVar {
		f += VaryK
	}
	if s.pVar {
		f += VaryP
	}
	for _, r := range s.data {
		sweep.Add(r.Index())
	}
	return
}

// Value returns a named column value from the set for a given index
func (s *Set) Value(idx Index, name string) float64 {
	for _, r := range s.data {
		if idx.Match(r.Index()) {
			return r.Value(name)
		}
	}
	return math.NaN()
}

// Values returns the values for named columns at a given index
func (s *Set) Values(idx Index, names []string) map[string]float64 {
	res := make(map[string]float64)
	for _, key := range names {
		res[key] = s.Value(idx, key)
	}
	return res
}

//----------------------------------------------------------------------

// Table data for post-processing (plot)
type Table struct {
	Name   string   // name of table (plot name)
	NumIdx int      // number of indices (parameters)
	Dims   []string // column names
	Vals   [][]any  // values (float64 or complex128)
	Refs   []int    // column references (to selection set)
}

func TblValue[T any](tbl *Table, row, col int) (v T) {
	var ok bool
	var res T
	if res, ok = tbl.Vals[row][col].(T); ok {
		v = res
	}
	return
}

//----------------------------------------------------------------------

// database initialization statements
var ini = `
create table performance (
    id      integer primary key,    -- database record id
	freq    integer not null,       -- operating frequency
	mat     varchar(15) not null,   -- wire material
	dia     float not null,         -- wire diameter
	height  float not null,         -- antenna height
	ground  integer not null,       -- ground type
	gType   integer not null,       -- ground mode
    k       float not null,         -- wing span in lambda
    param   float default null,     -- free parameter
    Gmax    float not null,         -- maximum gain
    Gmean   float not null,         -- mean gain
    SD      float not null,         -- gain std. deviation
    Zr      float not null,         -- antenna resistance
    Zi      float not null,         -- antenna reactance
	mdl     varchar(63) default '', -- model
	opt     varchar(63) default '', -- optimization
	gen     varchar(63) default '', -- generator
    fdir    varchar(255) not null,  -- model path
    ftag    varchar(31) not null,   -- model tag
    seed    integer not null,       -- randomizer seed
    mthds   integer default 0,      -- number of opt methods
    steps   integer default 0,      -- number of steps
    sims    integer default 0,      -- number of simulations
    elapsed integer default 0       -- elapsed time in seconds
);
create unique index idx_file on performance(fdir,ftag);
`

// Database for optimization results
type Database struct {
	inst *sql.DB
}

// Open SQLite3 database from file
func OpenDatabase(fname string) (db *Database, err error) {
	db = new(Database)
	if db.inst, err = sql.Open("sqlite3", fname); err == nil {
		var num int64
		row := db.inst.QueryRow("select count(*) from performance")
		if err = row.Scan(&num); err != nil {
			// initialize database
			_, err = db.inst.Exec(ini)
		}
	}
	return
}

// Close database
func (db *Database) Close() error {
	if db.inst == nil {
		return errors.New("database not opened")
	}
	return db.inst.Close()
}

// Insert model parameters into database
func (db *Database) Insert(rec *Record) error {
	stmt := "replace into performance(fdir,ftag,mdl,gen,opt,seed,freq,mat,dia," +
		"height,ground,gType,k,param,Gmax,Gmean,SD,Zr,Zi,mthds,steps,sims,elapsed)" +
		" values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
	_, err := db.inst.Exec(stmt,
		rec.Path, rec.Tag, rec.Mdl, rec.Gen, rec.Opt, rec.Seed, rec.Freq,
		rec.Wire.Material, rec.Wire.Diameter, rec.Gnd.Height, rec.Gnd.Mode,
		rec.Gnd.Type, rec.K, rec.Param, rec.Perf.Gain.Max, rec.Perf.Gain.Mean,
		rec.Perf.Gain.SD, real(rec.Perf.Z), imag(rec.Perf.Z), rec.Stats.NumMthds,
		rec.Stats.NumSteps, rec.Stats.NumSims, int(rec.Stats.Elapsed.Seconds()),
	)
	return err
}

// Set returns a set of performance records for a given directory
func (db *Database) Set(fdir string, filter Index) (set *Set, err error) {
	// perform query
	tpl := "select id,k,param,Gmax,Gmean,SD,Zr,Zi,ftag from performance where fdir='%s' order by k,param asc"
	stmt := fmt.Sprintf(tpl, fdir)
	var rows *sql.Rows
	if rows, err = db.inst.Query(stmt); err != nil {
		return
	}
	defer rows.Close()

	// read data
	set = NewSet()
	var param sql.NullFloat64
	for rows.Next() {
		// read record from database
		r := new(Row)
		if err = rows.Scan(&r.id, &r.idx.k, &param, &r.gmax, &r.gmean, &r.sd, &r.zr, &r.zi, &r.ftag); err != nil {
			return
		}
		r.idx.param = math.NaN()
		if param.Valid {
			r.idx.param = param.Float64
		}
		r.fdir = fdir
		// check if record matches filter
		if filter.Match(r.idx) {
			// add record to set
			set.Add(r)
		}
	}
	return
}

// ListPlotSets returns a list of names for available plot sets
func (db *Database) ListPlotSets() (sets map[string]*PlotSet, err error) {
	// perform query
	var rows *sql.Rows
	if rows, err = db.inst.Query("select distinct(fdir) from performance"); err != nil {
		return
	}
	// read data
	var s string
	var list []string
	for rows.Next() {
		if err = rows.Scan(&s); err != nil {
			return
		}
		list = append(list, s)
	}
	// close query
	if err = rows.Close(); err != nil {
		return
	}
	// create map of plot sets
	sets = make(map[string]*PlotSet)
	for _, dir := range list {
		ps := NewPlotSet(dir)
		if ps.Klist, ps.Plist, err = db.VarLists(dir); err != nil {
			return
		}
		ps.Tag = filepath.Dir(dir)
		sets[dir] = ps
	}
	return
}

// VarLists returns a list of (unique) 'k' and 'param' values for a dataset.
// If 'set' is empty, the values represent parameters in the whole database.
func (db *Database) VarLists(set string) (kList, pList []float64, err error) {
	if kList, err = db.varList(set, "k"); err != nil {
		return
	}
	pList, err = db.varList(set, "param")
	return
}

// varList returns a list of named parameter values for a dataset.
// If 'set' is empty, the values represent values of a parameter in
// the whole database.
func (db *Database) varList(set, par string) (list []float64, err error) {
	clause := ""
	if len(set) > 0 {
		clause = fmt.Sprintf("where fdir = '%s'", set)
	}
	stmt := fmt.Sprintf("select distinct(%s) from performance %s order by %s asc", par, clause, par)
	rows, err := db.inst.Query(stmt)
	if err != nil {
		return
	}
	var val sql.NullFloat64
	for rows.Next() {
		if err = rows.Scan(&val); err != nil {
			return
		}
		if val.Valid {
			list = append(list, val.Float64)
		}
	}
	return
}

// GetRows from the database with given where clause and ordering
func (db *Database) GetRows(clause, order string) (list []*Row, err error) {
	// assemble query statement
	stmt := "select Gmax,Gmean,SD,Zr,Zi,fdir,ftag from performance"
	if len(clause) > 0 {
		stmt += " where " + clause
	}
	if len(order) > 0 {
		stmt += " order by " + order
	}
	// perform query
	var rows *sql.Rows
	if rows, err = db.inst.Query(stmt); err != nil {
		return
	}
	defer rows.Close()

	// assemble result list
	for rows.Next() {
		r := new(Row)
		if err = rows.Scan(&r.gmax, &r.gmean, &r.sd, &r.zr, &r.zi, &r.fdir, &r.ftag); err != nil {
			return
		}
		list = append(list, r)
	}
	return
}

// DbStats holds database statistics
type DbStats struct {
	NumAnt   int64  // number of antennas
	NumSteps int64  // number of optimization steps
	NumSims  int64  // number of simulations
	Elapsed  int64  // elapsed simulation time (seconds)
	Duration string // human-readble duration
}

// Stats returns database statistics
func (db *Database) Stats() (stats *DbStats) {
	qInt := func(q string) (v int64) {
		row := db.inst.QueryRow("select " + q + " from performance")
		_ = row.Scan(&v)
		return
	}
	stats = new(DbStats)
	stats.NumAnt = qInt("count(*)")
	stats.NumSteps = qInt("sum(steps)")
	stats.NumSims = qInt("sum(sims)")
	stats.Elapsed = qInt("sum(elapsed)")
	stats.Duration = FormatDuration(stats.Elapsed)
	return
}
