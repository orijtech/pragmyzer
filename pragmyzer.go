// Copyright 2022 Orijtech, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pragmyzer

import (
	"go/ast"
	"go/types"
	"regexp"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const doc = `pragmyzer reports issues with Go pragma directives such as:
// go:embed file.txt
var data string

which is invalid and will silently be treated as a comment.
`

var Analyzer = &analysis.Analyzer{
	Name:     "pragmyzer",
	Doc:      doc,
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func imports(pkg *types.Package, target string) bool {
	for _, imp := range pkg.Imports() {
		if imp.Path() == target {
			return true
		}
	}
	return false
}

var reBadPragma = regexp.MustCompile(`\s*//\s+go:`)

func run(pass *analysis.Pass) (interface{}, error) {
	if false && !imports(pass.Pkg, "embed") {
		return nil, nil
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.File)(nil),
	}

	fset := pass.Fset
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		// Look for the last comment before the variable and see if it has go:embed
		// Find the comments for each node.
		f := n.(*ast.File)
		cmap := ast.NewCommentMap(fset, f, f.Comments)
		// Our target is to go looking for all comments.
		// Search for all variables.
		for _, cgL := range cmap {
			for _, cg := range cgL {
				for _, comment := range cg.List {
					if !reBadPragma.MatchString(comment.Text) {
						continue
					}
					pass.ReportRangef(comment, "pragmas should NOT have a leading space\nGot:  %q\nWant: %q", comment.Text, reBadPragma.ReplaceAllLiteralString(comment.Text, "//go:"))
				}
			}
		}
	})

	return nil, nil
}
