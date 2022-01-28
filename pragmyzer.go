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
	"strings"

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

func run(pass *analysis.Pass) (interface{}, error) {
	if !imports(pass.Pkg, "embed") {
		return nil, nil
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.File)(nil),
	}

	fset := pass.Fset
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		// 1. Reject anything that doesn't import "_ embed"
		// Look for the last comment before the variable and see if it has go:embed
		// Find the comments for each node.
		f := n.(*ast.File)
		cmap := ast.NewCommentMap(fset, f, f.Comments)
		// Our target is to go looking for all comments.
		// Search for all variables.
		for _, cgL := range cmap {
			for _, cg := range cgL {
				for _, comment := range cg.List {
					if !strings.HasPrefix(comment.Text, " ") {
						continue
					}

					trimmed := strings.TrimSpace(comment.Text)
					if !strings.HasPrefix(trimmed, "go:") {
						continue
					}
					pass.ReportRangef(comment, "pragmas should NOT have a space")
				}
			}
		}
	})

	return nil, nil
}
