package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/siddharth-bhatnagar/anvil/internal/analysis"
	"github.com/siddharth-bhatnagar/anvil/pkg/schema"
)

// AnalyzeFileTool analyzes a source file and returns its symbols
type AnalyzeFileTool struct {
	BaseTool
	parser      *analysis.GoParser
	highlighter *analysis.SyntaxHighlighter
}

// NewAnalyzeFileTool creates a new analyze file tool
func NewAnalyzeFileTool() *AnalyzeFileTool {
	return &AnalyzeFileTool{
		BaseTool: NewBaseTool(
			"analyze_file",
			"Analyze a source file to extract symbols (functions, types, variables)",
			[]schema.ToolParameter{
				{
					Name:        "path",
					Description: "Path to the file to analyze",
					Type:        "string",
					Required:    true,
				},
				{
					Name:        "include_private",
					Description: "Include non-exported symbols",
					Type:        "boolean",
					Required:    false,
					Default:     true,
				},
			},
		),
		parser:      analysis.NewGoParser(),
		highlighter: analysis.NewSyntaxHighlighter(),
	}
}

// Execute analyzes a file
func (t *AnalyzeFileTool) Execute(ctx context.Context, args map[string]any) (*schema.ToolResult, error) {
	pathVal, ok := args["path"]
	if !ok {
		return &schema.ToolResult{
			Success: false,
			Error:   "missing required parameter: path",
		}, fmt.Errorf("missing required parameter: path")
	}

	path := fmt.Sprintf("%v", pathVal)

	includePrivate := true
	if val, ok := args["include_private"]; ok {
		if boolVal, ok := val.(bool); ok {
			includePrivate = boolVal
		}
	}

	// Check if file exists
	info, err := os.Stat(path)
	if err != nil {
		return &schema.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("cannot access file: %v", err),
		}, err
	}

	if info.IsDir() {
		return &schema.ToolResult{
			Success: false,
			Error:   "path is a directory, not a file",
		}, fmt.Errorf("path is a directory")
	}

	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		return &schema.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to read file: %v", err),
		}, err
	}

	// Check if it's a Go file
	ext := filepath.Ext(path)
	if ext != ".go" {
		// For non-Go files, return basic info
		return &schema.ToolResult{
			Success: true,
			Output:  fmt.Sprintf("File: %s\nType: %s\nSize: %d bytes\n\n(Symbol analysis only available for Go files)", path, analysis.GetFileType(path), len(content)),
			Data: map[string]any{
				"path":     path,
				"type":     analysis.GetFileType(path),
				"size":     len(content),
				"language": analysis.GetLanguage(path),
			},
		}, nil
	}

	// Parse Go file
	symbols, err := t.parser.ParseFile(path, content)
	if err != nil {
		return &schema.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to parse file: %v", err),
		}, err
	}

	// Filter symbols if needed
	if !includePrivate {
		symbols = analysis.ExportedSymbols(symbols)
	}

	// Build output
	var output strings.Builder
	output.WriteString(fmt.Sprintf("File: %s\n", path))
	output.WriteString(fmt.Sprintf("Type: %s\n", analysis.GetFileType(path)))
	output.WriteString(fmt.Sprintf("Symbols: %d\n\n", len(symbols)))

	// Group symbols by kind
	functions := analysis.FilterSymbols(symbols, analysis.SymbolFunction)
	methods := analysis.FilterSymbols(symbols, analysis.SymbolMethod)
	types := analysis.FilterSymbols(symbols, analysis.SymbolStruct, analysis.SymbolInterface, analysis.SymbolType)
	vars := analysis.FilterSymbols(symbols, analysis.SymbolVariable, analysis.SymbolConstant)

	if len(functions) > 0 {
		output.WriteString("Functions:\n")
		for _, fn := range functions {
			output.WriteString(fmt.Sprintf("  %s %s (line %d)\n", fn.Kind.Icon(), fn.Name, fn.StartLine))
			if fn.Signature != "" {
				output.WriteString(fmt.Sprintf("    %s\n", fn.Signature))
			}
		}
		output.WriteString("\n")
	}

	if len(methods) > 0 {
		output.WriteString("Methods:\n")
		for _, m := range methods {
			receiver := ""
			if m.Receiver != "" {
				receiver = fmt.Sprintf("(%s) ", m.Receiver)
			}
			output.WriteString(fmt.Sprintf("  %s %s%s (line %d)\n", m.Kind.Icon(), receiver, m.Name, m.StartLine))
		}
		output.WriteString("\n")
	}

	if len(types) > 0 {
		output.WriteString("Types:\n")
		for _, typ := range types {
			output.WriteString(fmt.Sprintf("  %s %s (line %d)\n", typ.Kind.Icon(), typ.Name, typ.StartLine))
			for _, child := range typ.Children {
				output.WriteString(fmt.Sprintf("    %s %s: %s\n", child.Kind.Icon(), child.Name, child.Signature))
			}
		}
		output.WriteString("\n")
	}

	if len(vars) > 0 {
		output.WriteString("Variables/Constants:\n")
		for _, v := range vars {
			output.WriteString(fmt.Sprintf("  %s %s (line %d)\n", v.Kind.Icon(), v.Name, v.StartLine))
		}
	}

	// Convert symbols to serializable format
	symbolData := make([]map[string]any, len(symbols))
	for i, sym := range symbols {
		symbolData[i] = map[string]any{
			"name":      sym.Name,
			"kind":      sym.Kind.String(),
			"line":      sym.StartLine,
			"exported":  sym.Exported,
			"signature": sym.Signature,
		}
	}

	return &schema.ToolResult{
		Success: true,
		Output:  output.String(),
		Data: map[string]any{
			"path":    path,
			"symbols": symbolData,
		},
	}, nil
}

// RequiresApproval returns false for analysis operations
func (t *AnalyzeFileTool) RequiresApproval(args map[string]any) bool {
	return false
}

// FindSymbolTool finds a symbol definition in the codebase
type FindSymbolTool struct {
	BaseTool
	parser *analysis.GoParser
}

// NewFindSymbolTool creates a new find symbol tool
func NewFindSymbolTool() *FindSymbolTool {
	return &FindSymbolTool{
		BaseTool: NewBaseTool(
			"find_symbol",
			"Find where a symbol (function, type, variable) is defined",
			[]schema.ToolParameter{
				{
					Name:        "name",
					Description: "Name of the symbol to find",
					Type:        "string",
					Required:    true,
				},
				{
					Name:        "path",
					Description: "Directory to search in (defaults to current directory)",
					Type:        "string",
					Required:    false,
					Default:     ".",
				},
				{
					Name:        "kind",
					Description: "Kind of symbol to find (function, method, type, struct, interface, variable, constant)",
					Type:        "string",
					Required:    false,
				},
			},
		),
		parser: analysis.NewGoParser(),
	}
}

// Execute finds a symbol
func (t *FindSymbolTool) Execute(ctx context.Context, args map[string]any) (*schema.ToolResult, error) {
	nameVal, ok := args["name"]
	if !ok {
		return &schema.ToolResult{
			Success: false,
			Error:   "missing required parameter: name",
		}, fmt.Errorf("missing required parameter: name")
	}

	name := fmt.Sprintf("%v", nameVal)

	searchPath := "."
	if pathVal, ok := args["path"]; ok {
		searchPath = fmt.Sprintf("%v", pathVal)
	}

	var kindFilter *analysis.SymbolKind
	if kindVal, ok := args["kind"]; ok {
		kindStr := fmt.Sprintf("%v", kindVal)
		kind := parseSymbolKind(kindStr)
		if kind >= 0 {
			kindFilter = &kind
		}
	}

	var results []*analysis.Symbol

	// Walk through Go files
	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip directories and non-Go files
		if info.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}

		// Skip vendor and test files
		if strings.Contains(path, "/vendor/") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Read and parse file
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		symbols, err := t.parser.ParseFile(path, content)
		if err != nil {
			return nil
		}

		// Search for matching symbol
		for _, sym := range symbols {
			if sym.Name == name {
				if kindFilter == nil || sym.Kind == *kindFilter {
					results = append(results, sym)
				}
			}
			// Also check children
			for _, child := range sym.Children {
				if child.Name == name {
					if kindFilter == nil || child.Kind == *kindFilter {
						results = append(results, child)
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		return &schema.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("search failed: %v", err),
		}, err
	}

	if len(results) == 0 {
		return &schema.ToolResult{
			Success: true,
			Output:  fmt.Sprintf("No symbol found matching '%s'", name),
			Data: map[string]any{
				"found": false,
			},
		}, nil
	}

	// Build output
	var output strings.Builder
	output.WriteString(fmt.Sprintf("Found %d match(es) for '%s':\n\n", len(results), name))

	for _, sym := range results {
		output.WriteString(fmt.Sprintf("%s %s (%s)\n", sym.Kind.Icon(), sym.Name, sym.Kind.String()))
		output.WriteString(fmt.Sprintf("  File: %s:%d\n", sym.FilePath, sym.StartLine))
		if sym.Signature != "" {
			output.WriteString(fmt.Sprintf("  Signature: %s\n", sym.Signature))
		}
		if sym.DocComment != "" {
			doc := strings.TrimSpace(sym.DocComment)
			if len(doc) > 200 {
				doc = doc[:200] + "..."
			}
			output.WriteString(fmt.Sprintf("  Doc: %s\n", doc))
		}
		output.WriteString("\n")
	}

	// Convert to serializable format
	resultData := make([]map[string]any, len(results))
	for i, sym := range results {
		resultData[i] = map[string]any{
			"name":      sym.Name,
			"kind":      sym.Kind.String(),
			"file":      sym.FilePath,
			"line":      sym.StartLine,
			"signature": sym.Signature,
			"exported":  sym.Exported,
		}
	}

	return &schema.ToolResult{
		Success: true,
		Output:  output.String(),
		Data: map[string]any{
			"found":   true,
			"count":   len(results),
			"results": resultData,
		},
	}, nil
}

// RequiresApproval returns false for find operations
func (t *FindSymbolTool) RequiresApproval(args map[string]any) bool {
	return false
}

// parseSymbolKind converts a string to SymbolKind
func parseSymbolKind(s string) analysis.SymbolKind {
	switch strings.ToLower(s) {
	case "function", "func":
		return analysis.SymbolFunction
	case "method":
		return analysis.SymbolMethod
	case "type":
		return analysis.SymbolType
	case "struct":
		return analysis.SymbolStruct
	case "interface":
		return analysis.SymbolInterface
	case "variable", "var":
		return analysis.SymbolVariable
	case "constant", "const":
		return analysis.SymbolConstant
	case "field":
		return analysis.SymbolField
	default:
		return -1
	}
}
