// Package osexit содержит анализатор, запрещающий прямой вызов os.Exit
// в функции main пакета main.
package osexit

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Analyzer сообщает о каждом прямом вызове os.Exit внутри функции main
// пакета main. Такой вызов завершает программу мгновенно: отложенные
// функции не выполняются, буферы не сбрасываются, соединения не закрываются.
var Analyzer = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "запрещает прямой вызов os.Exit в функции main пакета main",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		if isGenerated(file) {
			continue
		}

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || fn.Name.Name != "main" {
				continue
			}

			inspectMain(pass, fn)
		}
	}

	return nil, nil
}

// inspectMain обходит тело функции main и сообщает о вызовах os.Exit.
func inspectMain(pass *analysis.Pass, fn *ast.FuncDecl) {
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "Exit" {
			return true
		}

		ident, ok := selector.X.(*ast.Ident)
		if !ok || ident.Name != "os" || ident.Obj != nil {
			return true
		}

		pass.Reportf(call.Pos(), "прямой вызов os.Exit в функции main запрещён")

		return true
	})
}

// isGenerated сообщает, сгенерирован ли файл автоматически. Такие файлы
// анализатор пропускает: например, go test собирает для каждого пакета
// временный main, который вызывает os.Exit напрямую.
func isGenerated(file *ast.File) bool {
	for _, group := range file.Comments {
		for _, comment := range group.List {
			text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
			if strings.HasPrefix(text, "Code generated ") && strings.HasSuffix(text, "DO NOT EDIT.") {
				return true
			}
		}
	}

	return false
}
