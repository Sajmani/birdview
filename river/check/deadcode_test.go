// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package check

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// testOracles are the non-test functions that exist on purpose as reference
// oracles or invariant accessors for BirdView's test suite rather than for
// production callers:
//
//   - internal/palette: WCAG contrast, CIE76 distance, and preset-alternation
//     checks over generated swatches and CSS colors;
//   - internal/boundary: Go reference parser and point-in-polygon lookup that
//     hold the browser's `here.js` lookup to known geographic answers;
//   - internal/dataset: AllMetrics enumerates every presented metric for
//     dataset.provenanceWellFormed, and (*Undrawn).Measured reports whether a
//     404 undrawn verdict carries empirical counts;
//   - internal/region: (Region).Zone exposes unexported r.zone so
//     region_test.go can check the partition invariant without widening
//     Region construction.
//
// Each is treated as an oracle root only after checking that no production path
// already reaches it, so a stale allowlist entry fails the test too.
var testOracles = map[string]bool{
	"birdview/internal/palette.(Swatch).ContrastWith":  true,
	"birdview/internal/palette.DeltaE76":               true,
	"birdview/internal/palette.AlternatesLightAndDark": true,
	"birdview/internal/boundary.Parse":                 true,
	"birdview/internal/boundary.(*File).Lookup":        true,
	"birdview/internal/dataset.(*Species).AllMetrics":  true,
	"birdview/internal/dataset.(*Undrawn).Measured":    true,
	"birdview/internal/region.(Region).Zone":           true,
}

type moduleImporter struct {
	pkgs     map[string]*types.Package
	fallback types.Importer
}

func (m *moduleImporter) Import(path string) (*types.Package, error) {
	if p, ok := m.pkgs[path]; ok {
		return p, nil
	}
	return m.fallback.Import(path)
}

func funcKey(fn *types.Func) string {
	if fn.Pkg() == nil {
		return fn.Name()
	}
	sig := fn.Type().(*types.Signature)
	if recv := sig.Recv(); recv != nil {
		rt := types.TypeString(recv.Type(), func(*types.Package) string { return "" })
		return fn.Pkg().Path() + ".(" + rt + ")." + fn.Name()
	}
	return fn.Pkg().Path() + "." + fn.Name()
}

// findUnreachableFunctions parses and type-checks every non-test Go package in
// root under modulePath, traces function reachability from every main/init
// entry point, and checks the declared testOracles allowlist.
func findUnreachableFunctions(t *testing.T, root, modulePath string, oracles map[string]bool) (dead, staleOracles, missingOracles []string) {
	t.Helper()
	fset := token.NewFileSet()
	pkgFiles := map[string][]*ast.File{}
	pkgDeps := map[string][]string{}

	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if p != root && (strings.HasPrefix(name, ".") || name == "third_party" || name == "testdata" || name == "fetched") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".go") || strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(fset, p, nil, 0)
		if err != nil {
			t.Fatalf("parsing %s: %v", p, err)
		}
		relDir, _ := filepath.Rel(root, filepath.Dir(p))
		relDir = filepath.ToSlash(relDir)
		impPath := modulePath
		if relDir != "." {
			impPath = modulePath + "/" + relDir
		}
		pkgFiles[impPath] = append(pkgFiles[impPath], f)
		for _, imp := range f.Imports {
			ip, _ := strconv.Unquote(imp.Path.Value)
			if ip == modulePath || strings.HasPrefix(ip, modulePath+"/") {
				pkgDeps[impPath] = append(pkgDeps[impPath], ip)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	imp := &moduleImporter{
		pkgs:     map[string]*types.Package{},
		fallback: importer.ForCompiler(fset, "source", nil),
	}
	infos := map[string]*types.Info{}
	visited := map[string]bool{}
	var checkPkg func(path string)
	checkPkg = func(path string) {
		if visited[path] {
			return
		}
		visited[path] = true
		files := pkgFiles[path]
		if len(files) == 0 {
			return
		}
		for _, dep := range pkgDeps[path] {
			checkPkg(dep)
		}
		info := &types.Info{
			Defs:       map[*ast.Ident]types.Object{},
			Uses:       map[*ast.Ident]types.Object{},
			Selections: map[*ast.SelectorExpr]*types.Selection{},
		}
		cfg := types.Config{Importer: imp}
		tp, err := cfg.Check(path, fset, files, info)
		if err != nil {
			t.Fatalf("type-checking %s: %v", path, err)
		}
		imp.pkgs[path] = tp
		infos[path] = info
	}
	var allPaths []string
	for p := range pkgFiles {
		allPaths = append(allPaths, p)
	}
	sort.Strings(allPaths)
	for _, p := range allPaths {
		checkPkg(p)
	}

	declared := map[string]*types.Func{}
	edges := map[string][]string{}
	var roots []string
	ifaceMethodNames := map[string]bool{}

	// First pass: collect declared functions, package-level roots, and interface method calls.
	for _, p := range allPaths {
		info := infos[p]
		isMain := pkgFiles[p][0].Name.Name == "main"
		for _, f := range pkgFiles[p] {
			for _, decl := range f.Decls {
				switch d := decl.(type) {
				case *ast.FuncDecl:
					obj, _ := info.Defs[d.Name].(*types.Func)
					if obj == nil {
						continue
					}
					k := funcKey(obj)
					declared[k] = obj
					if d.Recv == nil && (d.Name.Name == "init" || (isMain && d.Name.Name == "main")) {
						roots = append(roots, k)
					}
					if d.Body != nil {
						ast.Inspect(d.Body, func(n ast.Node) bool {
							switch x := n.(type) {
							case *ast.Ident:
								if u, ok := info.Uses[x].(*types.Func); ok && u.Pkg() != nil && strings.HasPrefix(u.Pkg().Path(), modulePath) {
									edges[k] = append(edges[k], funcKey(u))
								}
							case *ast.SelectorExpr:
								if sel := info.Selections[x]; sel != nil {
									if u, ok := sel.Obj().(*types.Func); ok {
										if types.IsInterface(sel.Recv()) {
											ifaceMethodNames[u.Name()] = true
										} else if u.Pkg() != nil && strings.HasPrefix(u.Pkg().Path(), modulePath) {
											edges[k] = append(edges[k], funcKey(u))
										}
									}
								}
							}
							return true
						})
					}
				case *ast.GenDecl:
					// Package-level var/const initializers execute at startup.
					ast.Inspect(d, func(n ast.Node) bool {
						if x, ok := n.(*ast.Ident); ok {
							if u, ok := info.Uses[x].(*types.Func); ok && u.Pkg() != nil && strings.HasPrefix(u.Pkg().Path(), modulePath) {
								roots = append(roots, funcKey(u))
							}
						}
						return true
					})
				}
			}
		}
	}

	// Exported methods dispatched via interface calls or Go html/template
	// expressions (e.g. $g.NeckMid in web/static/*.html, goldmark AST
	// Transformer.Transform, http.Handler.ServeHTTP, http.ResponseWriter, error).
	templateField := regexp.MustCompile(`\.[A-Z][A-Za-z0-9_]*\b`)
	tmplFiles, _ := filepath.Glob(filepath.Join(root, "web", "static", "*.html"))
	templateMethods := map[string]bool{}
	for _, tf := range tmplFiles {
		if b, err := os.ReadFile(tf); err == nil {
			for _, m := range templateField.FindAllString(string(b), -1) {
				templateMethods[m[1:]] = true
			}
		}
	}
	stdInterfaceMethods := map[string]bool{
		"Error": true, "String": true, "ServeHTTP": true,
		"Header": true, "Write": true, "WriteHeader": true,
		"Transform": true, "MarshalJSON": true, "UnmarshalJSON": true,
	}
	for k, fn := range declared {
		if sig := fn.Type().(*types.Signature); sig.Recv() != nil {
			if ifaceMethodNames[fn.Name()] || stdInterfaceMethods[fn.Name()] || templateMethods[fn.Name()] {
				// Exported method matching a dispatched interface or template method.
				if fn.Exported() {
					roots = append(roots, k)
				}
			}
		}
	}

	walk := func(seed []string) map[string]bool {
		seen := map[string]bool{}
		q := append([]string(nil), seed...)
		for len(q) > 0 {
			cur := q[0]
			q = q[1:]
			if seen[cur] {
				continue
			}
			seen[cur] = true
			for _, nxt := range edges[cur] {
				if !seen[nxt] {
					q = append(q, nxt)
				}
			}
		}
		return seen
	}

	prodReach := walk(roots)
	var oracleSeeds []string
	for o := range oracles {
		if _, ok := declared[o]; !ok {
			missingOracles = append(missingOracles, o)
			continue
		}
		if prodReach[o] {
			staleOracles = append(staleOracles, o)
			continue
		}
		oracleSeeds = append(oracleSeeds, o)
	}
	sort.Strings(missingOracles)
	sort.Strings(staleOracles)

	allReach := walk(append(roots, oracleSeeds...))
	for k := range declared {
		if !allReach[k] {
			dead = append(dead, k)
		}
	}
	sort.Strings(dead)
	return dead, staleOracles, missingOracles
}

// TestNoUnreachableCode checks that withdrawing a requirement or retiring a
// tool leaves no dead Go functions behind across the module, and that the
// test-oracle allowlist stays exact.
//
// Verifies: tech.bundleIsNormative. Criterion: repo.noUnreachableCode.
func TestNoUnreachableCode(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	dead, stale, missing := findUnreachableFunctions(t, root, "birdview", testOracles)
	if len(missing) > 0 {
		t.Errorf("testOracles names functions that no longer exist: %v", missing)
	}
	if len(stale) > 0 {
		t.Errorf("testOracles names functions now reached from production code (remove from allowlist): %v", stale)
	}
	if len(dead) > 0 {
		t.Errorf("%d unreachable function(s) in non-test code:\n  %s", len(dead), strings.Join(dead, "\n  "))
	}
}
