package analysis

import (
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strings"
)

// SymbolKind represents the kind of code symbol
type SymbolKind int

const (
	SymbolFunction SymbolKind = iota
	SymbolMethod
	SymbolType
	SymbolStruct
	SymbolInterface
	SymbolVariable
	SymbolConstant
	SymbolField
	SymbolImport
	SymbolPackage
)

// String returns the string representation of a symbol kind
func (k SymbolKind) String() string {
	switch k {
	case SymbolFunction:
		return "function"
	case SymbolMethod:
		return "method"
	case SymbolType:
		return "type"
	case SymbolStruct:
		return "struct"
	case SymbolInterface:
		return "interface"
	case SymbolVariable:
		return "variable"
	case SymbolConstant:
		return "constant"
	case SymbolField:
		return "field"
	case SymbolImport:
		return "import"
	case SymbolPackage:
		return "package"
	default:
		return "unknown"
	}
}

// Symbol icon returns an icon for the symbol kind
func (k SymbolKind) Icon() string {
	switch k {
	case SymbolFunction:
		return "ƒ"
	case SymbolMethod:
		return "m"
	case SymbolType:
		return "T"
	case SymbolStruct:
		return "S"
	case SymbolInterface:
		return "I"
	case SymbolVariable:
		return "v"
	case SymbolConstant:
		return "c"
	case SymbolField:
		return "."
	case SymbolImport:
		return "→"
	case SymbolPackage:
		return "P"
	default:
		return "?"
	}
}

// Symbol represents a code symbol (function, type, variable, etc.)
type Symbol struct {
	Name       string
	Kind       SymbolKind
	Signature  string     // Full signature (for functions/methods)
	Receiver   string     // Receiver type (for methods)
	StartLine  int
	EndLine    int
	StartCol   int
	EndCol     int
	DocComment string     // Associated documentation
	Children   []*Symbol  // Nested symbols (fields in structs, etc.)
	Exported   bool       // Whether the symbol is exported
	FilePath   string     // File containing the symbol
}

// GoParser parses Go source code and extracts symbols
type GoParser struct {
	fset *token.FileSet
}

// NewGoParser creates a new Go parser
func NewGoParser() *GoParser {
	return &GoParser{
		fset: token.NewFileSet(),
	}
}

// ParseFile parses a Go file and returns its symbols
func (p *GoParser) ParseFile(filename string, src []byte) ([]*Symbol, error) {
	file, err := parser.ParseFile(p.fset, filename, src, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var symbols []*Symbol

	// Extract package
	if file.Name != nil {
		symbols = append(symbols, &Symbol{
			Name:      file.Name.Name,
			Kind:      SymbolPackage,
			StartLine: p.fset.Position(file.Name.Pos()).Line,
			FilePath:  filename,
		})
	}

	// Extract imports
	for _, imp := range file.Imports {
		name := ""
		if imp.Name != nil {
			name = imp.Name.Name
		} else {
			// Get the last part of the import path
			path := strings.Trim(imp.Path.Value, "\"")
			parts := strings.Split(path, "/")
			name = parts[len(parts)-1]
		}

		symbols = append(symbols, &Symbol{
			Name:      name,
			Kind:      SymbolImport,
			Signature: imp.Path.Value,
			StartLine: p.fset.Position(imp.Pos()).Line,
			FilePath:  filename,
		})
	}

	// Walk the AST
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			sym := p.extractFunction(node, filename)
			symbols = append(symbols, sym)

		case *ast.GenDecl:
			syms := p.extractGenDecl(node, filename)
			symbols = append(symbols, syms...)
		}

		return true
	})

	// Sort by line number
	sort.Slice(symbols, func(i, j int) bool {
		return symbols[i].StartLine < symbols[j].StartLine
	})

	return symbols, nil
}

// extractFunction extracts a function or method symbol
func (p *GoParser) extractFunction(fn *ast.FuncDecl, filename string) *Symbol {
	sym := &Symbol{
		Name:      fn.Name.Name,
		Kind:      SymbolFunction,
		StartLine: p.fset.Position(fn.Pos()).Line,
		EndLine:   p.fset.Position(fn.End()).Line,
		StartCol:  p.fset.Position(fn.Pos()).Column,
		EndCol:    p.fset.Position(fn.End()).Column,
		Exported:  ast.IsExported(fn.Name.Name),
		FilePath:  filename,
	}

	// Check if it's a method
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		sym.Kind = SymbolMethod
		// Extract receiver type
		if recv := fn.Recv.List[0]; recv.Type != nil {
			sym.Receiver = formatType(recv.Type)
		}
	}

	// Build signature
	sym.Signature = p.buildFunctionSignature(fn)

	// Extract doc comment
	if fn.Doc != nil {
		sym.DocComment = fn.Doc.Text()
	}

	return sym
}

// extractGenDecl extracts symbols from a general declaration (type, var, const)
func (p *GoParser) extractGenDecl(decl *ast.GenDecl, filename string) []*Symbol {
	var symbols []*Symbol

	// Get doc comment for the whole declaration
	declDoc := ""
	if decl.Doc != nil {
		declDoc = decl.Doc.Text()
	}

	for _, spec := range decl.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			sym := &Symbol{
				Name:       s.Name.Name,
				StartLine:  p.fset.Position(s.Pos()).Line,
				EndLine:    p.fset.Position(s.End()).Line,
				Exported:   ast.IsExported(s.Name.Name),
				FilePath:   filename,
				DocComment: declDoc,
			}

			// Determine the kind based on the type
			switch t := s.Type.(type) {
			case *ast.StructType:
				sym.Kind = SymbolStruct
				sym.Children = p.extractStructFields(t, filename)
			case *ast.InterfaceType:
				sym.Kind = SymbolInterface
				sym.Children = p.extractInterfaceMethods(t, filename)
			default:
				sym.Kind = SymbolType
				sym.Signature = formatType(s.Type)
			}

			// Get spec-level doc if available
			if s.Doc != nil {
				sym.DocComment = s.Doc.Text()
			}

			symbols = append(symbols, sym)

		case *ast.ValueSpec:
			for i, name := range s.Names {
				sym := &Symbol{
					Name:       name.Name,
					StartLine:  p.fset.Position(name.Pos()).Line,
					Exported:   ast.IsExported(name.Name),
					FilePath:   filename,
					DocComment: declDoc,
				}

				if decl.Tok == token.CONST {
					sym.Kind = SymbolConstant
				} else {
					sym.Kind = SymbolVariable
				}

				// Get type if specified
				if s.Type != nil {
					sym.Signature = formatType(s.Type)
				} else if i < len(s.Values) {
					// Try to infer type from value
					sym.Signature = formatExpr(s.Values[i])
				}

				symbols = append(symbols, sym)
			}
		}
	}

	return symbols
}

// extractStructFields extracts field symbols from a struct
func (p *GoParser) extractStructFields(st *ast.StructType, filename string) []*Symbol {
	var fields []*Symbol

	if st.Fields == nil {
		return fields
	}

	for _, field := range st.Fields.List {
		for _, name := range field.Names {
			sym := &Symbol{
				Name:      name.Name,
				Kind:      SymbolField,
				StartLine: p.fset.Position(name.Pos()).Line,
				Exported:  ast.IsExported(name.Name),
				FilePath:  filename,
			}

			if field.Type != nil {
				sym.Signature = formatType(field.Type)
			}

			if field.Doc != nil {
				sym.DocComment = field.Doc.Text()
			} else if field.Comment != nil {
				sym.DocComment = field.Comment.Text()
			}

			fields = append(fields, sym)
		}

		// Handle embedded fields
		if len(field.Names) == 0 && field.Type != nil {
			sym := &Symbol{
				Name:      formatType(field.Type),
				Kind:      SymbolField,
				StartLine: p.fset.Position(field.Pos()).Line,
				Signature: "embedded",
				Exported:  true, // Embedded fields are typically exported types
				FilePath:  filename,
			}
			fields = append(fields, sym)
		}
	}

	return fields
}

// extractInterfaceMethods extracts method signatures from an interface
func (p *GoParser) extractInterfaceMethods(iface *ast.InterfaceType, filename string) []*Symbol {
	var methods []*Symbol

	if iface.Methods == nil {
		return methods
	}

	for _, method := range iface.Methods.List {
		for _, name := range method.Names {
			sym := &Symbol{
				Name:      name.Name,
				Kind:      SymbolMethod,
				StartLine: p.fset.Position(name.Pos()).Line,
				Exported:  ast.IsExported(name.Name),
				FilePath:  filename,
			}

			if fn, ok := method.Type.(*ast.FuncType); ok {
				sym.Signature = formatFuncType(fn)
			}

			methods = append(methods, sym)
		}

		// Handle embedded interfaces
		if len(method.Names) == 0 && method.Type != nil {
			sym := &Symbol{
				Name:      formatType(method.Type),
				Kind:      SymbolInterface,
				StartLine: p.fset.Position(method.Pos()).Line,
				Signature: "embedded",
				FilePath:  filename,
			}
			methods = append(methods, sym)
		}
	}

	return methods
}

// buildFunctionSignature builds a function signature string
func (p *GoParser) buildFunctionSignature(fn *ast.FuncDecl) string {
	var sig strings.Builder

	sig.WriteString("func ")

	// Add receiver
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		sig.WriteString("(")
		sig.WriteString(formatFieldList(fn.Recv))
		sig.WriteString(") ")
	}

	sig.WriteString(fn.Name.Name)
	sig.WriteString(formatFuncType(fn.Type))

	return sig.String()
}

// formatType formats an AST type expression as a string
func formatType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + formatType(t.X)
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + formatType(t.Elt)
		}
		return "[...]" + formatType(t.Elt)
	case *ast.MapType:
		return "map[" + formatType(t.Key) + "]" + formatType(t.Value)
	case *ast.SelectorExpr:
		return formatType(t.X) + "." + t.Sel.Name
	case *ast.ChanType:
		return "chan " + formatType(t.Value)
	case *ast.FuncType:
		return formatFuncType(t)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.StructType:
		return "struct{}"
	default:
		return "?"
	}
}

// formatFuncType formats a function type
func formatFuncType(fn *ast.FuncType) string {
	var sig strings.Builder

	sig.WriteString("(")
	if fn.Params != nil {
		sig.WriteString(formatFieldList(fn.Params))
	}
	sig.WriteString(")")

	if fn.Results != nil && len(fn.Results.List) > 0 {
		if len(fn.Results.List) == 1 && len(fn.Results.List[0].Names) == 0 {
			sig.WriteString(" ")
			sig.WriteString(formatType(fn.Results.List[0].Type))
		} else {
			sig.WriteString(" (")
			sig.WriteString(formatFieldList(fn.Results))
			sig.WriteString(")")
		}
	}

	return sig.String()
}

// formatFieldList formats a field list
func formatFieldList(fl *ast.FieldList) string {
	if fl == nil || len(fl.List) == 0 {
		return ""
	}

	var parts []string
	for _, field := range fl.List {
		typeStr := formatType(field.Type)

		if len(field.Names) == 0 {
			parts = append(parts, typeStr)
		} else {
			for _, name := range field.Names {
				parts = append(parts, name.Name+" "+typeStr)
			}
		}
	}

	return strings.Join(parts, ", ")
}

// formatExpr formats an expression (for variable initialization)
func formatExpr(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.BasicLit:
		return e.Kind.String()
	case *ast.CompositeLit:
		if e.Type != nil {
			return formatType(e.Type)
		}
		return "composite"
	case *ast.CallExpr:
		return formatType(e.Fun) + "(...)"
	default:
		return "expression"
	}
}

// FindSymbol finds a symbol by name in a list of symbols
func FindSymbol(symbols []*Symbol, name string) *Symbol {
	for _, sym := range symbols {
		if sym.Name == name {
			return sym
		}
		// Check children
		if child := FindSymbol(sym.Children, name); child != nil {
			return child
		}
	}
	return nil
}

// FilterSymbols filters symbols by kind
func FilterSymbols(symbols []*Symbol, kinds ...SymbolKind) []*Symbol {
	kindSet := make(map[SymbolKind]bool)
	for _, k := range kinds {
		kindSet[k] = true
	}

	var filtered []*Symbol
	for _, sym := range symbols {
		if kindSet[sym.Kind] {
			filtered = append(filtered, sym)
		}
	}
	return filtered
}

// ExportedSymbols returns only exported symbols
func ExportedSymbols(symbols []*Symbol) []*Symbol {
	var exported []*Symbol
	for _, sym := range symbols {
		if sym.Exported {
			exported = append(exported, sym)
		}
	}
	return exported
}
