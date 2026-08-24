// Утилита reset генерирует методы Reset() для структур, помеченных
// комментарием // generate:reset.
//
// # Запуск
//
// Из корня проекта:
//
//	go run ./cmd/reset
//
// Утилита сканирует текущую директорию и все вложенные, находит структуры
// с комментарием // generate:reset и для каждого пакета создаёт файл
// reset.gen.go с методами Reset().
//
// Директорию можно указать явно первым аргументом:
//
//	go run ./cmd/reset ./internal
//
// # Правила сброса
//
// Метод Reset() приводит поля структуры к начальному состоянию:
//
//   - примитивы получают нулевое значение: 0, "", false;
//   - слайсы обрезаются по длине, но не зануляются: s = s[:0];
//   - мапы очищаются встроенной функцией clear;
//   - вложенные структуры сбрасываются своим методом Reset(), если он есть;
//   - указатели, отличные от nil, сбрасываются по тем же правилам.
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// generateMark — комментарий, помечающий структуры для генерации.
const generateMark = "generate:reset"

// generatedFile — имя файла, в который попадают сгенерированные методы.
const generatedFile = "reset.gen.go"

// structInfo описывает структуру, для которой генерируется метод Reset.
type structInfo struct {
	spec *ast.StructType
	name string
}

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}

	if err := run(root); err != nil {
		log.Fatal(err)
	}
}

// run обходит директории начиная с root и генерирует методы Reset.
func run(root string) error {
	dirs, err := packageDirs(root)
	if err != nil {
		return err
	}

	for _, dir := range dirs {
		if err := generateForDir(dir); err != nil {
			return fmt.Errorf("%s: %w", dir, err)
		}
	}

	return nil
}

// packageDirs собирает директории с Go-файлами, пропуская служебные.
func packageDirs(root string) ([]string, error) {
	var dirs []string

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !entry.IsDir() {
			return nil
		}

		name := entry.Name()
		if path != root && (strings.HasPrefix(name, ".") || name == "testdata" || name == "vendor") {
			return filepath.SkipDir
		}

		dirs = append(dirs, path)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan %s: %w", root, err)
	}

	return dirs, nil
}

// generateForDir разбирает файлы директории и пишет reset.gen.go,
// если в них есть помеченные структуры.
func generateForDir(dir string) error {
	packages, err := parseDir(dir)
	if err != nil {
		return err
	}

	for name, files := range packages {
		structs := markedStructs(files)
		if len(structs) == 0 {
			continue
		}

		source, err := render(name, structs)
		if err != nil {
			return err
		}

		path := filepath.Join(dir, generatedFile)
		if err := os.WriteFile(path, source, 0644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}

		log.Printf("%s: сгенерировано методов — %d", path, len(structs))
	}

	return nil
}

// parseDir разбирает Go-файлы директории и группирует их по имени пакета.
// Ранее сгенерированные файлы и тесты пропускаются.
func parseDir(dir string) (map[string][]*ast.File, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read dir: %w", err)
	}

	fset := token.NewFileSet()
	packages := map[string][]*ast.File{}

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") ||
			name == generatedFile || strings.HasSuffix(name, "_test.go") {
			continue
		}

		file, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", name, err)
		}

		pkg := file.Name.Name
		packages[pkg] = append(packages[pkg], file)
	}

	return packages, nil
}

// markedStructs возвращает структуры пакета, помеченные комментарием.
func markedStructs(files []*ast.File) []structInfo {
	var structs []structInfo

	for _, file := range files {
		for _, decl := range file.Decls {
			decl, ok := decl.(*ast.GenDecl)
			if !ok || decl.Tok != token.TYPE {
				continue
			}

			for _, spec := range decl.Specs {
				spec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				structType, ok := spec.Type.(*ast.StructType)
				if !ok || !marked(decl, spec) {
					continue
				}

				structs = append(structs, structInfo{name: spec.Name.Name, spec: structType})
			}
		}
	}

	sort.Slice(structs, func(i, j int) bool { return structs[i].name < structs[j].name })

	return structs
}

// marked сообщает, стоит ли над объявлением комментарий generate:reset.
func marked(decl *ast.GenDecl, spec *ast.TypeSpec) bool {
	for _, doc := range []*ast.CommentGroup{spec.Doc, decl.Doc} {
		if doc == nil {
			continue
		}

		for _, comment := range doc.List {
			if strings.Contains(comment.Text, generateMark) {
				return true
			}
		}
	}

	return false
}

// render собирает исходный код файла с методами Reset.
func render(pkg string, structs []structInfo) ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteString("// Code generated by cmd/reset. DO NOT EDIT.\n\n")
	buf.WriteString("package " + pkg + "\n")

	for _, item := range structs {
		buf.WriteString("\n// Reset приводит поля " + item.name + " к начальному состоянию.\n")
		buf.WriteString("func (v *" + item.name + ") Reset() {\n")
		buf.WriteString("\tif v == nil {\n\t\treturn\n\t}\n")

		for _, field := range item.spec.Fields.List {
			for _, name := range fieldNames(field) {
				buf.WriteString(resetField("v."+name, field.Type))
			}
		}

		buf.WriteString("}\n")
	}

	source, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format generated code: %w", err)
	}

	return source, nil
}

// fieldNames возвращает имена поля, пропуская встроенные и безымянные.
func fieldNames(field *ast.Field) []string {
	names := make([]string, 0, len(field.Names))

	for _, name := range field.Names {
		names = append(names, name.Name)
	}

	return names
}

// resetField возвращает код сброса одного поля по его типу.
func resetField(target string, fieldType ast.Expr) string {
	switch t := fieldType.(type) {
	case *ast.Ident:
		if zero, ok := basicZero(t.Name); ok {
			return "\t" + target + " = " + zero + "\n"
		}

		return resetNested(target, t.Name)
	case *ast.ArrayType:
		if t.Len != nil {
			return "\t" + target + " = " + zeroLiteral(fieldType) + "\n"
		}

		return "\t" + target + " = " + target + "[:0]\n"
	case *ast.MapType:
		return "\tclear(" + target + ")\n"
	case *ast.StarExpr:
		return resetPointer(target, t.X)
	case *ast.SelectorExpr, *ast.StructType:
		return resetNested(target, "")
	default:
		return "\t" + target + " = " + zeroLiteral(fieldType) + "\n"
	}
}

// resetPointer возвращает код сброса поля-указателя: значение под
// указателем сбрасывается по общим правилам, сам указатель не зануляется.
func resetPointer(target string, pointed ast.Expr) string {
	body := resetField("(*"+target+")", pointed)

	return "\tif " + target + " != nil {\n\t" + strings.ReplaceAll(strings.TrimSuffix(body, "\n"), "\n", "\n\t") + "\n\t}\n"
}

// resetNested возвращает код сброса поля составного типа: если у типа есть
// метод Reset, вызывается он, иначе поле получает нулевое значение.
func resetNested(target, typeName string) string {
	code := "\tif resetter, ok := any(&" + target + ").(interface{ Reset() }); ok {\n" +
		"\t\tresetter.Reset()\n"

	if typeName != "" {
		code += "\t} else {\n\t\t" + target + " = " + typeName + "{}\n\t}\n"
	} else {
		code += "\t}\n"
	}

	return code
}

// basicZero возвращает нулевое значение примитивного типа и признак того,
// что тип действительно примитивный.
func basicZero(typeName string) (string, bool) {
	switch typeName {
	case "string":
		return `""`, true
	case "bool":
		return "false", true
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
		"float32", "float64", "byte", "rune", "complex64", "complex128":
		return "0", true
	case "error", "any":
		return "nil", true
	default:
		return "", false
	}
}

// zeroLiteral возвращает нулевое значение для составного типа.
func zeroLiteral(fieldType ast.Expr) string {
	var buf bytes.Buffer

	if err := format.Node(&buf, token.NewFileSet(), fieldType); err != nil {
		return "nil"
	}

	switch fieldType.(type) {
	case *ast.ArrayType, *ast.StructType:
		return buf.String() + "{}"
	default:
		return "nil"
	}
}
