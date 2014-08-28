// Copyright (c) 2014, Kevin Walsh.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This interface was derived from the code in src/tao/tao_guard.h.

package tao

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"code.google.com/p/goprotobuf/proto"

	"github.com/jlmucb/cloudproxy/tao/auth"
	"github.com/jlmucb/cloudproxy/util"
	"github.com/kevinawalsh/datalog/dlengine"
)

// Signing context for signatures on a set of Tao datalog rules.
const (
	DatalogRulesSigningContext = "Datalog Rules Signing Context V1"
)

// DatalogGuardConfig holds persistent configuration data for a DatalogGuard.
type DatalogGuardConfig struct {
	SignedRulesPath string
}

// DatalogGuard implements a datalog-based policy engine. Rules in this engine
// have the form:
//   (forall X, Y, Z... : F implies G)
// where
//   F is a predicate or a conjunction of predicates
//   G is a predicate
// All predicate arguments must be either concrete terms (Int, Str, Prin, etc.)
// or term-valued variables (TermVar) bound by the quantification. Any variable
// appearing in G must also appear in F. If there are no variables, the
// quantification can be omitted. The implication and its antecedent F can be
// omitted (in which case there can be no variables so the quantification must
// be omitted as well).
//
// TODO(kwalsh) We could easily support a slightly broader class of formulas,
// e.g. by allowing G to be a conjunct of predicates, or by allowing a
// disjunction of conjunctions for F. Anything beyond that seems complicated.
//
// Datalog translation
//
// We assume K speaksfor the guard, where K is the key used to sign the policy
// file. If there is no signing key, a temporary principal (with a bogus key) is
// used for K instead. All deduction takes place within the worldview of Guard.
// Other than this relationship between K and the guard, we don't model the says
// and speaksfor logic within datalog.
//
// Term objects are usually translated to datalog by just printing them. In this
// case, a Prin object must not contain any TermVar objects. TermVar objects
// must be uppercase.
//
// "Term says Pred(...)" is translated to "says(Term, \"Pred\", ...)".
//
// "Pred(...)" alone is translated to "says(K, \"Pred\", ...)".
//
// "forall ... F1 and F2 and ... imp G" is translated to "G :- F1, F2, ...".
type DatalogGuard struct {
	Config DatalogGuardConfig
	Key    *Verifier
	// TODO(kwalsh) maybe use a version number or timestamp inside the file?
	modTime time.Time // Modification time of signed rules file at time of reading.
	db      DatalogRules
	dl      *dlengine.Engine
}

// NewTemporaryDatalogGuard returns a new datalog guard with a fresh, unsigned,
// non-persistent rule set.
func NewTemporaryDatalogGuard() Guard {
	return &DatalogGuard{dl: dlengine.NewEngine()}
}

// NewDatalogGuard returns a new datalog guard that uses a signed, persistent
// signed rule set. ReloadIfModified() should be called to load the rule set.
func NewDatalogGuard(key *Verifier, config DatalogGuardConfig) (*DatalogGuard, error) {
	if key == nil || config.SignedRulesPath == "" {
		return nil, newError("datalog guard missing key or path")
	}
	g := &DatalogGuard{Config: config, Key: key, dl: dlengine.NewEngine()}
	return g, nil
}

// SubprincipalName returns subprincipal DatalogGuard, for temporary guards, or
// DatalogGuard(<key>) for persistent guards.
func (g *DatalogGuard) Subprincipal() auth.SubPrin {
	if g.Key == nil {
		e := auth.PrinExt{Name: "DatalogGuard"}
		return auth.SubPrin{e}
	} else {
		e := auth.PrinExt{Name: "DatalogGuard", Arg: []auth.Term{g.Key.ToPrincipal()}}
		return auth.SubPrin{e}
	}
}

// ReloadIfModified reads all persistent policy data from disk if the file
// timestamp is more recent than the last time it was read.
func (g *DatalogGuard) ReloadIfModified() error {
	if g.Key == nil {
		return nil
	}
	file, err := os.Open(g.Config.SignedRulesPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// before parsing, check the timestamp
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if !info.ModTime().After(g.modTime) {
		return nil
	}

	serialized, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	var sdb SignedDatalogRules
	if err := proto.Unmarshal(serialized, &sdb); err != nil {
		return err
	}
	if ok, err := g.Key.Verify(sdb.SerializedRules, DatalogRulesSigningContext, sdb.Signature); !ok {
		if err != nil {
			return err
		}
		return newError("datalog rule signature did not verify")
	}
	var db DatalogRules
	if err := proto.Unmarshal(sdb.SerializedRules, &db); err != nil {
		return err
	}
	g.Clear()
	g.modTime = info.ModTime()
	g.db = db
	for _, rule := range db.Rules {
		r, err := auth.UnmarshalForm(rule)
		if err != nil {
			return err
		}
		err = g.assert(r)
		if err != nil {
			return err
		}
	}
	return nil
}

// Save writes all persistent policy data to disk, signed by key.
func (g *DatalogGuard) Save(key *Signer) error {
	if key == nil {
		return newError("datalog temporary ruleset can't be saved")
	}
	rules, err := proto.Marshal(&g.db)
	if err != nil {
		return err
	}
	sig, err := key.Sign(rules, DatalogRulesSigningContext)
	if err != nil {
		return err
	}
	sdb := &SignedDatalogRules{
		SerializedRules: rules,
		Signature:       sig,
	}
	serialized, err := proto.Marshal(sdb)
	if err != nil {
		return err
	}
	if err := util.WritePath(g.Config.SignedRulesPath, serialized, 0777, 0666); err != nil {
		return err
	}
	return nil
}

func setContains(vars []string, v string) bool {
	for _, s := range vars {
		if s == v {
			return true
		}
	}
	return false
}

func setRemove(vars *[]string, v string) {
	if vars == nil {
		return
	}
	for i := 0; i < len(*vars); i++ {
		if (*vars)[i] == v {
			(*vars)[i] = (*vars)[len(*vars)-1]
			*vars = (*vars)[:len(*vars)-1]
			i--
		}
	}
}

func stripQuantifiers(q auth.Form) (f auth.Form, vars []string) {
	for {
		var v string
		switch f := q.(type) {
		case auth.Forall:
			v = f.Var
			q = f.Body
		case *auth.Forall:
			v = f.Var
			q = f.Body
		default:
			return q, vars
		}
		if !setContains(vars, v) {
			vars = append(vars, v)
		}
	}
}

func flattenConjuncts(f ...auth.Form) (conjuncts []auth.Form) {
	for _, f := range f {
		switch f := f.(type) {
		case auth.And:
			conjuncts = append(conjuncts, flattenConjuncts(f.Conjunct...)...)
		case *auth.And:
			conjuncts = append(conjuncts, flattenConjuncts(f.Conjunct...)...)
		default:
			conjuncts = append(conjuncts, f)
		}
	}
	return
}

func stripConditions(f auth.Form) (conds []auth.Form, consequent auth.Form) {
	switch f := f.(type) {
	case auth.Implies:
		conds = flattenConjuncts(f.Antecedent)
		consequent = f.Consequent
	case *auth.Implies:
		conds = flattenConjuncts(f.Antecedent)
		consequent = f.Consequent
	default:
		consequent = f
	}
	return
}

func checkTermVarUsage(vars []string, unusedVars *[]string, e ...auth.Term) error {
	for _, e := range e {
		switch e := e.(type) {
		case auth.TermVar, *auth.TermVar:
			if !setContains(vars, e.String()) {
				return fmt.Errorf("illegal quantification variable: %v\n", e)
			}
			setRemove(unusedVars, e.String())
		case auth.Prin:
			err := checkTermVarUsage(vars, unusedVars, e.Key)
			if err != nil {
				return err
			}
			for _, ext := range e.Ext {
				err := checkTermVarUsage(vars, unusedVars, ext.Arg...)
				if err != nil {
					return err
				}
			}
		case *auth.Prin:
			err := checkTermVarUsage(vars, unusedVars, e.Key)
			if err != nil {
				return err
			}
			for _, ext := range e.Ext {
				err := checkTermVarUsage(vars, unusedVars, ext.Arg...)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func checkFormVarUsage(vars []string, unusedVars *[]string, e ...auth.Form) error {
	for _, e := range e {
		switch e := e.(type) {
		case auth.Pred:
			err := checkTermVarUsage(vars, unusedVars, e.Arg...)
			if err != nil {
				return err
			}
		case *auth.Pred:
			err := checkTermVarUsage(vars, unusedVars, e.Arg...)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *DatalogGuard) stmtToDatalog(f auth.Form, vars []string, unusedVars *[]string) (string, error) {
	speaker := "guard"
	if g.Key != nil {
		speaker = g.Key.ToPrincipal().String()
	}
	stmt, ok := f.(*auth.Says)
	if !ok {
		var val auth.Says
		val, ok = (f).(auth.Says)
		if ok {
			stmt = &val
		}
	}
	if ok {
		err := checkTermVarUsage(vars, unusedVars, stmt.Speaker)
		if err != nil {
			return "", err
		}
		speaker = stmt.Speaker.String()
		f = stmt.Message
	}
	err := checkFormVarUsage(vars, unusedVars, f)
	if err != nil {
		return "", err
	}
	pred, ok := f.(*auth.Pred)
	if !ok {
		var val auth.Pred
		val, ok = f.(auth.Pred)
		if ok {
			pred = &val
		}
	}
	if !ok {
		return "", fmt.Errorf("unsupported datalog statement: %v", f)
	}
	args := []string{fmt.Sprintf("%q", speaker), fmt.Sprintf("%q", pred.Name)}
	for _, arg := range pred.Arg {
		args = append(args, fmt.Sprintf("%q", arg.String()))
	}
	return "says(" + strings.Join(args, ", ") + ")", nil
}

// formToDatalogRule converts (a subset of) auth.Form to datalog syntax.
func (g *DatalogGuard) formToDatalogRule(f auth.Form) (string, error) {
	f, vars := stripQuantifiers(f)
	conditions, consequent := stripConditions(f)
	// vars must be upper-case
	for _, v := range vars {
		if len(v) == 0 || v[0] < 'A' || v[0] > 'Z' {
			return "", fmt.Errorf("illegal quantification variable")
		}
	}
	// convert the conditions
	dcond := make([]string, len(conditions))
	unusedVars := append([]string{}, vars...)
	for i, cond := range conditions {
		var err error
		dcond[i], err = g.stmtToDatalog(cond, vars, &unusedVars)
		if err != nil {
			return "", err
		}
	}
	// check for safety
	if len(unusedVars) > 0 {
		return "", fmt.Errorf("unsafe datalog variable usage: % s", unusedVars)
	}
	goal, err := g.stmtToDatalog(consequent, vars, nil)
	if err != nil {
		return "", err
	}
	if len(dcond) > 0 {
		return goal + " :- " + strings.Join(dcond, ", "), nil
	} else {
		return goal, nil
	}
}

func (g *DatalogGuard) findRule(f auth.Form) (string, int, error) {
	rule, err := g.formToDatalogRule(f)
	if err != nil {
		return "", -1, err
	}
	for i, ser := range g.db.Rules {
		f2, err := auth.UnmarshalForm(ser)
		if err != nil {
			continue
		}
		rule2, err := g.formToDatalogRule(f2)
		if err != nil {
			continue
		}
		if rule == rule2 {
			return rule, i, nil
		}
	}
	return rule, -1, nil
}

func (g *DatalogGuard) assert(f auth.Form) error {
	rule, idx, err := g.findRule(f)
	if err != nil {
		return err
	}
	if idx >= 0 {
		return nil
	}
	err = g.dl.Assert(rule)
	if err != nil {
		return err
	}
	g.db.Rules = append(g.db.Rules, auth.Marshal(f))
	return nil
}

func (g *DatalogGuard) retract(f auth.Form) error {
	rule, idx, err := g.findRule(f)
	if err != nil {
		return err
	}
	if idx < 0 {
		return fmt.Errorf("no such rule")
	}
	err = g.dl.Retract(rule)
	if err != nil {
		return err
	}
	g.db.Rules = append(g.db.Rules[:idx], g.db.Rules[idx+1:]...)
	return nil
}

func (g *DatalogGuard) query(f auth.Form) (bool, error) {
	q, err := g.stmtToDatalog(f, nil, nil)
	if err != nil {
		return false, err
	}
	ans, err := g.dl.Query(q)
	if err != nil {
		return false, err
	}
	return len(ans) > 0, nil
}

func makeDatalogPredicate(p auth.Prin, op string, args []string) auth.Pred {
	a := []interface{}{p, op}
	for _, s := range args {
		a = append(a, s)
	}
	return auth.MakePredicate("Authorized", a...)
}

// Authorize adds an authorization for p to perform op(args).
func (g *DatalogGuard) Authorize(p auth.Prin, op string, args []string) error {
	return g.assert(makeDatalogPredicate(p, op, args))
}

// Retract removes an authorization for p to perform op(args).
func (g *DatalogGuard) Retract(p auth.Prin, op string, args []string) error {
	return g.retract(makeDatalogPredicate(p, op, args))
}

// IsAuthorized checks whether p is authorized to perform op(args).
func (g *DatalogGuard) IsAuthorized(p auth.Prin, op string, args []string) bool {
	ok, _ := g.query(makeDatalogPredicate(p, op, args))
	return ok
}

// AddRule adds a policy rule.
func (g *DatalogGuard) AddRule(rule string) error {
	var r auth.AnyForm
	_, err := fmt.Sscanf("("+rule+")", "%v", &r)
	if err != nil {
		return err
	}
	return g.assert(r.Form)
}

// RetractRule removes a rule previously added via AddRule() or the
// equivalent Authorize() call.
func (g *DatalogGuard) RetractRule(rule string) error {
	err := g.ReloadIfModified()
	if err != nil {
		return err
	}
	var r auth.AnyForm
	_, err = fmt.Sscanf("("+rule+")", "%v", &r)
	if err != nil {
		return err
	}
	return g.retract(r.Form)
}

// Clear removes all rules.
func (g *DatalogGuard) Clear() error {
	g.db.Rules = nil
	g.dl = dlengine.NewEngine()
	return nil
}

// Query the policy. Implementations of this interface should support
// at least queries of the form: Authorized(P, op, args...).
func (g *DatalogGuard) Query(query string) (bool, error) {
	err := g.ReloadIfModified()
	if err != nil {
		return false, err
	}
	var r auth.AnyForm
	_, err = fmt.Sscanf("("+query+")", "%v", &r)
	if err != nil {
		return false, err
	}
	return g.query(r.Form)
}

// RuleCount returns a count of the total number of rules.
func (g *DatalogGuard) RuleCount() int {
	return len(g.db.Rules)
}

// GetRule returns the ith policy rule, if it exists.
func (g *DatalogGuard) GetRule(i int) string {
	if i < 0 || i >= len(g.db.Rules) {
		return ""
	}
	rule := g.db.Rules[i]
	r, err := auth.UnmarshalForm(rule)
	if err != nil {
		return ""
	}
	return r.String()
}

// RuleDebugString returns a debug string for the ith policy rule, if it exists.
func (g *DatalogGuard) RuleDebugString(i int) string {
	if i < 0 || i >= len(g.db.Rules) {
		return ""
	}
	rule := g.db.Rules[i]
	r, err := auth.UnmarshalForm(rule)
	if err != nil {
		return ""
	}
	return r.ShortString()
}

// String returns a string suitable for showing users authorization info.
func (g *DatalogGuard) String() string {
	rules := make([]string, len(g.db.Rules))
	for i := range g.db.Rules {
		rules[i] = g.GetRule(i)
	}
	return "DatalogGuard{\n" + strings.Join(rules, "\n") + "}\n"
}