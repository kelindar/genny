package parse

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/scanner"
	"go/token"
	"io"
	"os"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/imports"
)

var header = `
// Code generated with https://github.com/kelindar/genny DO NOT EDIT.
// Any changes will be lost if this file is regenerated.


`

const (
	debug = false
)

var (
	packageKeyword = []byte("package")
	importKeyword  = []byte("import")
	openBrace      = []byte("(")
	closeBrace     = []byte(")")
	genericPackage = "generic"
	genericType    = "generic.Type"
	genericNumber  = "generic.Number"
	linefeed       = "\r\n"
)
var unwantedLinePrefixes = [][]byte{
	[]byte("//go:generate genny "),
	[]byte("//go:generate $GOPATH/bin/genny "),
}

func subIntoLiteral(prefix, lit, typeTemplate, specificType string) string {
	// print("l >> %s ... tt >> %s", lit, typeTemplate)
	if lit == typeTemplate {
		return specificType
	}

	if !containsFold(lit, typeTemplate) {
		return lit
	}

	specificLg := wordify(specificType, true)
	specificSm := wordify(specificType, false)

	var replacer string
	switch {
	case isExported(typeTemplate):
		replacer = specificLg
	default:
		replacer = specificSm
	}

	// result := lit //replaceBoundary(lit, typeTemplate, specificType)
	typeregex := regexp.MustCompile("\\b" + typeTemplate + "\\b")
	result := typeregex.ReplaceAllString(lit, specificType)
	result = strings.Replace(result, typeTemplate, replacer, -1)

	if strings.HasPrefix(result, specificLg) && !isExported(lit) {
		result = strings.Replace(result, specificLg, replacer, 1)
	}

	// Two special cases for number functions
	switch {
	case isNumber(specificType) && strings.HasSuffix(prefix, "func "):
		result = replaceBoundary(result, typeTemplate, specificLg)
	case isNumber(specificType) && strings.HasSuffix(prefix, "."):
		result = replaceBoundary(result, typeTemplate, specificLg)
	case isNumber(specificType) && strings.HasSuffix(prefix, "// "):
		result = replaceBoundary(result, typeTemplate, specificLg)
	default:
		result = replaceBoundary(result, typeTemplate, specificSm)
	}

	return result
}

func subTypeIntoComment(line, typeTemplate, specificType string) string {
	var subbed string
	for _, w := range strings.Fields(line) {
		subbed = subbed + subIntoLiteral("// ", w, typeTemplate, specificType) + " "
	}
	return subbed
}

// Does the heavy lifting of taking a line of our code and
// sbustituting a type into there for our generic type
func subTypeIntoLine(line, typeTemplate, specificType string) string {
	src := []byte(line)
	var s scanner.Scanner
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(src))
	s.Init(file, src, nil, scanner.ScanComments)
	output := ""
	for {
		position, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		// print("%s -> %s", lit, tok)

		switch {
		case tok == token.COMMENT:
			subbed := subTypeIntoComment(lit, typeTemplate, specificType)
			output = output + subbed + " "
		case tok.IsLiteral():
			//println("LITERAL", line, lit, typeTemplate, specificType)
			prefix := line[:position-1]
			subbed := subIntoLiteral(prefix, lit, typeTemplate, specificType)
			output = output + subbed + " "
		default:
			output = output + tok.String() + " "
		}

	}
	return output
}

// typeSet looks like "KeyType: int, ValueType: string"
func generateSpecific(filename string, in io.ReadSeeker, typeSet map[string]string) ([]byte, error) {

	// ensure we are at the beginning of the file
	in.Seek(0, os.SEEK_SET)

	// parse the source file
	fs := token.NewFileSet()
	file, err := parser.ParseFile(fs, filename, in, 0)
	if err != nil {
		return nil, &errSource{Err: err}
	}

	// make sure every generic.Type is represented in the types
	// argument.
	for _, decl := range file.Decls {
		switch it := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range it.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				switch tt := ts.Type.(type) {
				case *ast.SelectorExpr:
					if name, ok := tt.X.(*ast.Ident); ok {
						if name.Name == genericPackage {
							if _, ok := typeSet[ts.Name.Name]; !ok {
								return nil, &errMissingSpecificType{GenericType: ts.Name.Name}
							}
						}
					}
				}
			}
		}
	}

	in.Seek(0, os.SEEK_SET)

	var buf bytes.Buffer

	comment := ""
	scanner := bufio.NewScanner(in)
	reInterfaceBegin := regexp.MustCompile(`^\s*type\s+\w+\s+interface\s*\{`)
	reInterfaceEnd := regexp.MustCompile(`^\s*\}`)
	var interfaceLines []string
	interfaceContainsType := false
	for scanner.Scan() {

		line := scanner.Text()

		if reInterfaceBegin.MatchString(line) {
			interfaceLines = []string{""}
		}

		if len(interfaceLines) > 0 && reInterfaceEnd.MatchString(line) {
			if !interfaceContainsType {
				for _, li := range append(interfaceLines, line)[1:] {
					buf.WriteString(li + "\n")
				}
			}
			interfaceLines, interfaceContainsType = nil, false
			continue
		}

		// does this line contain generic.Type?
		if strings.Contains(line, genericType) || strings.Contains(line, genericNumber) {
			comment = ""
			if len(interfaceLines) > 0 {
				interfaceContainsType = true
			}
			continue
		}

		for t, specificType := range typeSet {
			if containsFold(line, t) {
				newLine := subTypeIntoLine(line, t, specificType)
				line = newLine
			}
		}

		if comment != "" {
			buf.WriteString(makeLine(comment))
			comment = ""
		}

		// is this line a comment?
		// TODO: should we handle /* */ comments?
		if strings.HasPrefix(line, "//") {
			// record this line to print later
			comment = line
			continue
		}

		// write the line
		if len(interfaceLines) > 0 {
			interfaceLines = append(interfaceLines, line)
		} else {
			buf.WriteString(makeLine(line))
		}
	}

	// write trailing comment, if any
	if comment != "" {
		buf.WriteString(makeLine(comment))
		comment = ""
	}

	// write it out
	return buf.Bytes(), nil
}

// Generics parses the source file and generates the bytes replacing the
// generic types for the keys map with the specific types (its value).
func Generics(filename, pkgName string, in io.ReadSeeker, typeSets []map[string]string, importPaths []string, stripTag string, useAstImpl bool) ([]byte, error) {
	localUnwantedLinePrefixes := [][]byte{}
	localUnwantedLinePrefixes = append(localUnwantedLinePrefixes, unwantedLinePrefixes...)

	if stripTag != "" {
		localUnwantedLinePrefixes = append(localUnwantedLinePrefixes, []byte(fmt.Sprintf("// +build %s", stripTag)))
	}

	totalOutput := [][]byte{}

	for _, typeSet := range typeSets {

		// generate the specifics
		var parsed []byte
		var err error
		if useAstImpl {
			parsed, err = generateSpecificAst(filename, in, typeSet)
		} else {
			parsed, err = generateSpecific(filename, in, typeSet)
		}
		if err != nil {
			return nil, err
		}

		totalOutput = append(totalOutput, parsed)
	}

	// clean up the code line by line

	packageFound := false
	// Whether to wait for the "genny:start" comment to start copying. This will be set to true
	// after we have went through the first generated type, so subsequent generated types will
	// not copy anything before that line
	fileHasGennyStart := false
	importLineIndex := -1
	var collectedImports stringArraySet
	cleanOutputLines := []string{header}
	for fileIndex, transformedOutput := range totalOutput {
		insideImportBlock := false
		packageFoundForFile := false
		scanner := bufio.NewScanner(bytes.NewReader(transformedOutput))
		pastGennyStart := false

	FORSCAN:
		for scanner.Scan() {

			if bytes.HasPrefix(scanner.Bytes(), []byte("//genny:start")) {
				pastGennyStart = true
				fileHasGennyStart = true
				continue
			}

			// end of imports block?
			if insideImportBlock {
				if bytes.HasSuffix(scanner.Bytes(), closeBrace) {
					insideImportBlock = false
					// cleanOutputLines = append(cleanOutputLines, fmt.Sprintln(")"))
				} else {
					collectedImports = collectedImports.append(makeLine(scanner.Text()))
					// cleanOutputLines = append(cleanOutputLines, makeLine(scanner.Text()))
				}
				continue
			}

			if bytes.HasPrefix(scanner.Bytes(), packageKeyword) {
				packageFoundForFile = true
				if !packageFound {
					packageFound = true
					cleanOutputLines = append(cleanOutputLines, makeLine(scanner.Text()))
				}
				continue
			} else if bytes.HasPrefix(scanner.Bytes(), importKeyword) {
				if importLineIndex == -1 {
					importLineIndex = len(cleanOutputLines)
				}
				if bytes.HasSuffix(scanner.Bytes(), openBrace) {
					insideImportBlock = true
					// cleanOutputLines = append(cleanOutputLines, fmt.Sprintln("import ("))
				} else {
					importLine := strings.TrimSpace(makeLine(scanner.Text()))
					importLine = strings.TrimSpace(importLine[6:])
					collectedImports = collectedImports.append(importLine)
					// cleanOutputLines = append(cleanOutputLines, importLine)
				}

				continue
			}

			if fileIndex != 0 && !packageFoundForFile {
				continue
			}

			if fileHasGennyStart && !pastGennyStart {
				continue
			}

			// check all unwantedLinePrefixes - and skip them
			for _, prefix := range localUnwantedLinePrefixes {
				if bytes.HasPrefix(scanner.Bytes(), prefix) {
					continue FORSCAN
				}
			}

			cleanOutputLines = append(cleanOutputLines, makeLine(scanner.Text()))
		}
	}

	var linesWithImport []string
	linesWithImport = append(linesWithImport, cleanOutputLines[:importLineIndex]...)
	linesWithImport = append(linesWithImport, fmt.Sprintln("import ("))
	linesWithImport = append(linesWithImport, collectedImports...)
	linesWithImport = append(linesWithImport, fmt.Sprintln(")"))
	linesWithImport = append(linesWithImport, cleanOutputLines[importLineIndex+1:]...)

	cleanOutput := strings.Join(linesWithImport, "")

	output := []byte(cleanOutput)

	// change package name
	if pkgName != "" {
		output = changePackage(bytes.NewReader([]byte(output)), pkgName)
	}
	if len(importPaths) > 0 {
		output = addImports(bytes.NewReader(output), importPaths)
	}
	// fix the imports
	var err error
	output, err = imports.Process(filename, output, nil)
	if err != nil {
		return nil, &errImports{Err: err}
	}

	return output, nil
}

func makeLine(s string) string {
	return fmt.Sprintln(strings.TrimRight(s, linefeed))
}

// isAlphaNumeric gets whether the rune is alphanumeric or _.
func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// wordify turns a type into a nice word for function and type
// names etc.
// If s matches format `<Title>:<Type>` then <Title> is returned
func wordify(s string, exported bool) string {
	if sepIdx := strings.Index(s, ":"); sepIdx >= 0 {
		s = s[:sepIdx]
	} else {
		s = strings.TrimRight(s, "{}")
		s = strings.TrimLeft(s, "*&")
		s = strings.Replace(s, ".", "", -1)
	}
	if !exported {
		return strings.ToLower(string(s[0])) + s[1:]
	}

	return strings.ToUpper(string(s[0])) + s[1:]
}

// typify gets type name from string.
// if string contains ":" then right part is returned otherwise string itself is returned
func typify(s string) string {
	if sepIdx := strings.Index(s, ":"); sepIdx >= 0 {
		return s[sepIdx+1:]
	}
	return s
}

func changePackage(r io.Reader, pkgName string) []byte {
	var out bytes.Buffer
	sc := bufio.NewScanner(r)
	done := false

	for sc.Scan() {
		s := sc.Text()

		if !done && strings.HasPrefix(s, "package") {
			parts := strings.Split(s, " ")
			parts[1] = pkgName
			s = strings.Join(parts, " ")
			done = true
		}

		fmt.Fprintln(&out, s)
	}
	return out.Bytes()
}

func addImports(r io.Reader, importPaths []string) []byte {
	var out bytes.Buffer
	sc := bufio.NewScanner(r)
	done := false

	for sc.Scan() {
		s := sc.Text()

		if !done && strings.HasPrefix(s, "package") {
			fmt.Fprintln(&out, s)
			for _, imp := range importPaths {
				fmt.Fprintf(&out, "import \"%s\"\n", imp)
			}
			done = true
			continue
		}

		fmt.Fprintln(&out, s)
	}
	return out.Bytes()
}

// ===== Start AST related implementation =====

type replaceSpec struct {
	genericType  string
	specificType string
}

func (rs replaceSpec) toType() string {
	return typify(rs.specificType)
}

func (rs replaceSpec) toWord(uppercase bool) string {
	return wordify(rs.specificType, uppercase)
}

func (rs replaceSpec) String() string {
	return fmt.Sprintf("%s -> %s", rs.genericType, rs.specificType)
}

func deleteAllComments(file *ast.File, root ast.Node) {
	ast.Inspect(root, func(n ast.Node) bool {
		if comment, ok := n.(*ast.CommentGroup); ok {
			deleteComment(file, comment)
		}
		return true
	})
}

func deleteComment(file *ast.File, comment *ast.CommentGroup) {
	for i, com := range file.Comments {
		if com == comment {
			file.Comments = append(file.Comments[:i], file.Comments[i+1:]...)
			break
		}
	}
}

func transformText(text string, spec replaceSpec) string {
	reExact := regexp.MustCompile("\\b" + spec.genericType + "\\b")
	text = reExact.ReplaceAllString(text, spec.specificType)
	return replaceBoundaryFunc(text, spec.genericType, func(match string) string {
		return spec.toWord(unicode.IsUpper(rune(match[0])))
	})
}

func transformType(ident *ast.Ident, spec replaceSpec, log string) *ast.Ident {
	if ident.Name != spec.genericType {
		// If they are not identical, the type is something like genericQueue. Perform normal text
		// transformation
		return transformIdentifier(ident, spec, log)
	}
	output := *ident
	output.Name = spec.toType()
	return &output
}

func transformIdentifier(ident *ast.Ident, spec replaceSpec, log string) *ast.Ident {
	transformed := transformText(ident.Name, spec)

	output := *ident
	output.Name = transformed
	return &output
}

func generateSpecificType(fs *token.FileSet, file *ast.File, spec replaceSpec) {
	astutil.Apply(file,
		func(c *astutil.Cursor) bool {
			switch v := c.Node().(type) {
			case *ast.File:
				for _, commentGroup := range v.Comments {
					for _, cmt := range commentGroup.List {
						// Replace the comments
						cmt.Text = transformText(cmt.Text, spec)
					}
				}
			case *ast.Ident:
				var newIdent *ast.Ident
				if containsFold(v.Name, spec.genericType) {
					// print("    >>>>>", v.Name, spec, reflect.TypeOf(c.Parent()))
					switch p := c.Parent().(type) {
					case *ast.ArrayType:
						// []generic
						// []genericValue
						newIdent = transformType(v, spec, "ARRAY TYPE")
					case *ast.ValueSpec:
						if v == p.Type {
							// var something generic
							// var something genericValue
							newIdent = transformType(v, spec, "VALUE TYPE")
						} else {
							// var generic string
							// var genericVariable string
							newIdent = transformIdentifier(v, spec, "VARNAME")
						}
					case *ast.AssignStmt:
						// myGeneric := something
						// something := myGeneric
						newIdent = transformType(v, spec, "ASSIGN")
					case *ast.CallExpr:
						if v == p.Fun {
							// generic(something), a.k.a. type conversion
							// genericSomething(something)
							newIdent = transformType(v, spec, "TYPE CONVERSION")
						} else {
							// myfunc(generic)
							// myfunc(genericVariable)
							newIdent = transformIdentifier(v, spec, "ARG VARNAME")
						}
					case *ast.TypeSpec:
						if v == p.Name {
							// type generic someType
							// type genericValue someType
							newIdent = transformIdentifier(v, spec, "TYPE NAME")
						} else {
							// type newType generic
							// type newType genericSomething
							newIdent = transformType(v, spec, "TYPE VALUE")
						}
					case *ast.Field:
						if v == p.Type {
							// func a(g generic) or func a(g genericValue)
							newIdent = transformType(v, spec, "FIELD TYPE")
						} else {
							// func a(genericSomething someType)
							newIdent = transformIdentifier(v, spec, "ARG DECL NAME")
						}
					case *ast.FuncDecl:
						// func PrintGeneric()
						newIdent = transformIdentifier(v, spec, "FUNC NAME")
					case *ast.SelectorExpr:
						// a.PrintMyType()
						newIdent = transformIdentifier(v, spec, "SELECTOR")
					case *ast.StarExpr:
						// *generic or *somethingGeneric
						newIdent = transformType(v, spec, "STAR EXPR")
					case *ast.CompositeLit:
						if v == p.Type {
							// myGen := generic{field1: 1, field2: 2}
							newIdent = transformType(v, spec, "COMPOSITE LITERAL")
						}
					case *ast.BinaryExpr:
						// myGeneric == something
						newIdent = transformType(v, spec, "BINARY")
					case *ast.MapType:
						// var m map[generic]myvalue
						// var m map[mykey]generic
						newIdent = transformType(v, spec, "MAP TYPE")
					case *ast.KeyValueExpr:
						// MyStruct{ field: genericVal }
						// MyStruct{ genericVal: field }
						newIdent = transformType(v, spec, "KEY VALUE EXPR")
					case *ast.BranchStmt:
						// ignore
					case *ast.TypeAssertExpr:
						// a.(generic)
						newIdent = transformType(v, spec, "TYPE ASSERT EXPR")
					case *ast.File:
						if debug {
							print("UNRESOLVED???", v.Name, spec, reflect.TypeOf(c.Parent()))
						}
					default:
						if debug {
							print(">>>>>", v.Name, spec, reflect.TypeOf(c.Parent()))
						}
					}
				}
				if newIdent != nil {
					c.Replace(newIdent)
				}
			case *ast.TypeSpec:
				if isGenericTypeDefinition(v) {
					deleteAllComments(file, v)
					c.Delete()
				}
			default:
				print("UNKNOWN:", v, reflect.TypeOf(v))
			}
			return true
		},
		func(c *astutil.Cursor) bool {
			switch v := c.Node().(type) {
			case *ast.GenDecl:
				// If the declaration became empty after removing `type myType generic.Type`,
				// remove the declaration as well
				if len(v.Specs) == 0 {
					deleteComment(file, v.Doc)
					c.Delete()
				}
			}
			return true
		})
}

func isGenericTypeDefinition(typeSpec *ast.TypeSpec) bool {
	switch t := typeSpec.Type.(type) {
	case *ast.SelectorExpr:
		return isGenericTypeSelector(t)
	case *ast.InterfaceType:
		for _, field := range t.Methods.List {
			// TODO: need to check new specific type also implements the other methods in the
			// interface?
			if selector, ok := field.Type.(*ast.SelectorExpr); ok {
				if isGenericTypeSelector(selector) {
					return true
				}
			}
		}
	}
	return false
}

func isGenericTypeSelector(selector *ast.SelectorExpr) bool {
	if ident, ok := selector.X.(*ast.Ident); ok {
		if ident.Name == "generic" &&
			(selector.Sel.Name == "Type" || selector.Sel.Name == "Number") {
			return true
		}
	}
	return false
}

func generateSpecificAst(filename string, in io.ReadSeeker, typeSet map[string]string) ([]byte, error) {

	// ensure we are at the beginning of the file
	in.Seek(0, os.SEEK_SET)

	// parse the source file
	fs := token.NewFileSet()
	file, err := parser.ParseFile(fs, filename, in, parser.ParseComments)
	if err != nil {
		return nil, &errSource{Err: err}
	}

	// make sure every generic.Type is represented in the types
	// argument.
	for _, decl := range file.Decls {
		switch it := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range it.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				switch tt := ts.Type.(type) {
				case *ast.SelectorExpr:
					if name, ok := tt.X.(*ast.Ident); ok {
						if name.Name == genericPackage {
							if _, ok := typeSet[ts.Name.Name]; !ok {
								return nil, &errMissingSpecificType{GenericType: ts.Name.Name}
							}
						}
					}
				}
			}
		}
	}

	var buf bytes.Buffer
	for t, specificType := range typeSet {
		generateSpecificType(fs, file, replaceSpec{t, specificType})
	}

	err = printer.Fprint(&buf, fs, file)
	return buf.Bytes(), err
}

func containsFold(s, substring string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substring))
}

func indexFold(s, substring string) int {
	return strings.Index(strings.ToLower(s), strings.ToLower(substring))
}

func isExported(lit string) bool {
	if len(lit) == 0 {
		return false
	}
	return unicode.IsUpper(rune(lit[0]))
}

func containsBoundary(s, substring string) bool {
	return indexBoundary(s, substring) >= 0
}

func indexBoundary(s, substring string) int {
	pos := indexFold(s, substring)
	if pos == -1 {
		return -1
	}
	startIsBoundary := pos == 0 || !unicode.IsLetter(rune(s[pos-1])) || !unicode.IsLetter(rune(s[pos])) || unicode.IsUpper(rune(s[pos]))
	// endPos := pos + len(substring)
	endIsBoundary := true
	// TODO: Find a way to deal with "-s", "-ed", etc
	// endIsBoundary := endPos == len(s) || !unicode.IsLetter(rune(s[endPos])) || unicode.IsUpper(rune(s[endPos]))
	if startIsBoundary && endIsBoundary {
		return pos
	}
	recursive := indexBoundary(s[pos+1:], substring)
	if recursive == -1 {
		return -1
	}
	return recursive + pos + 1
}

func replaceBoundary(s, old string, newstring string) string {
	i := 0
	output := ""
	for {
		pos := indexBoundary(s[i:], old)
		if pos == -1 {
			break
		}
		pos += i
		output += s[i:pos]
		output += newstring
		i = pos + len(old)
	}
	output += s[i:]
	return output
}

func replaceBoundaryFunc(s, old string, replace func(string) string) string {
	i := 0
	output := ""
	for {
		pos := indexBoundary(s[i:], old)
		if pos == -1 {
			break
		}
		pos += i
		output += s[i:pos]
		output += replace(s[pos : pos+len(old)])
		i = pos + len(old)
	}
	output += s[i:]
	return output
}

func print(a ...interface{}) {
	if debug {
		fmt.Println(a...)
	}
}
