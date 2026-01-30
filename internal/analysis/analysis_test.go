package analysis

import (
	"strings"
	"testing"
)

// Test sample Go code
const sampleGoCode = `package main

import "fmt"

// Greet prints a greeting message
func Greet(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}

type Person struct {
	Name string
	Age  int
}

func (p *Person) String() string {
	return p.Name
}
`

// TestNewSyntaxHighlighter tests highlighter creation
func TestNewSyntaxHighlighter(t *testing.T) {
	highlighter := NewSyntaxHighlighter()
	if highlighter == nil {
		t.Fatal("expected highlighter, got nil")
	}

	if highlighter.style == nil {
		t.Error("expected style to be set")
	}

	if highlighter.formatter == nil {
		t.Error("expected formatter to be set")
	}
}

// TestNewSyntaxHighlighterWithStyle tests highlighter creation with custom style
func TestNewSyntaxHighlighterWithStyle(t *testing.T) {
	styles := []string{"monokai", "github", "invalid-style"}

	for _, style := range styles {
		t.Run(style, func(t *testing.T) {
			highlighter := NewSyntaxHighlighterWithStyle(style)
			if highlighter == nil {
				t.Fatal("expected highlighter, got nil")
			}

			// Should use fallback for invalid styles
			if highlighter.style == nil {
				t.Error("expected style to be set (fallback if invalid)")
			}
		})
	}
}

// TestHighlight tests basic syntax highlighting
func TestHighlight(t *testing.T) {
	highlighter := NewSyntaxHighlighter()

	tests := []struct {
		name     string
		code     string
		filename string
	}{
		{
			name:     "Go code",
			code:     `package main\nfunc main() {}`,
			filename: "test.go",
		},
		{
			name:     "Python code",
			code:     `def hello():\n    print("hello")`,
			filename: "test.py",
		},
		{
			name:     "JavaScript code",
			code:     `function hello() { console.log("hello"); }`,
			filename: "test.js",
		},
		{
			name:     "Unknown file",
			code:     `some text`,
			filename: "test.unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := highlighter.Highlight(tt.code, tt.filename)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result == "" {
				t.Error("expected non-empty highlighted output")
			}
		})
	}
}

// TestHighlightWithLanguage tests highlighting with explicit language
func TestHighlightWithLanguage(t *testing.T) {
	highlighter := NewSyntaxHighlighter()

	tests := []struct {
		language string
		code     string
	}{
		{"go", `package main`},
		{"python", `def hello(): pass`},
		{"javascript", `function hello() {}`},
		{"rust", `fn main() {}`},
		{"invalid", `some code`}, // Should fallback
	}

	for _, tt := range tests {
		t.Run(tt.language, func(t *testing.T) {
			result, err := highlighter.HighlightWithLanguage(tt.code, tt.language)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result == "" {
				t.Error("expected non-empty highlighted output")
			}
		})
	}
}

// TestHighlightLines tests highlighting specific lines
func TestHighlightLines(t *testing.T) {
	highlighter := NewSyntaxHighlighter()
	code := "line 1\nline 2\nline 3\nline 4\nline 5"

	tests := []struct {
		name      string
		startLine int
		endLine   int
		wantError bool
	}{
		{"valid range", 2, 4, false},
		{"single line", 3, 3, false},
		{"start before 1", 0, 2, false},    // Should adjust to 1
		{"end after max", 3, 100, false},   // Should adjust to max
		{"inverted range", 4, 2, false},    // Should return empty
		{"entire file", 1, 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := highlighter.HighlightLines(code, "test.go", tt.startLine, tt.endLine)
			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Result can be empty for inverted ranges
			if tt.startLine <= tt.endLine && result == "" && tt.startLine >= 1 {
				t.Error("expected non-empty result for valid range")
			}
		})
	}
}

// TestGetLanguage tests language detection
func TestGetLanguage(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"test.go", "Go"},
		{"test.py", "Python"},
		{"test.js", "JavaScript"},
		{"test.ts", "TypeScript"},
		{"test.rs", "Rust"},
		{"test.java", "Java"},
		{"test.unknown", "plaintext"},
		{"", "plaintext"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := GetLanguage(tt.filename)
			if got != tt.want {
				t.Errorf("GetLanguage(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

// TestGetLanguageByExtension tests language detection by extension
func TestGetLanguageByExtension(t *testing.T) {
	tests := []struct {
		ext  string
		want string
	}{
		{".go", "Go"},
		{"go", "Go"},     // Without dot
		{".py", "Python"},
		{".js", "JavaScript"},
		{".unknown", "plaintext"},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			got := GetLanguageByExtension(tt.ext)
			if got != tt.want {
				t.Errorf("GetLanguageByExtension(%q) = %q, want %q", tt.ext, got, tt.want)
			}
		})
	}
}

// TestSupportedLanguages tests getting list of supported languages
func TestSupportedLanguages(t *testing.T) {
	languages := SupportedLanguages()

	if len(languages) == 0 {
		t.Error("expected non-empty list of supported languages")
	}

	// Check for some common languages
	expectedLanguages := []string{"Go", "Python", "JavaScript"}
	for _, expected := range expectedLanguages {
		found := false
		for _, lang := range languages {
			if lang == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected to find language %q in supported languages", expected)
		}
	}
}

// TestAvailableStyles tests getting list of available styles
func TestAvailableStyles(t *testing.T) {
	styles := AvailableStyles()

	if len(styles) == 0 {
		t.Error("expected non-empty list of available styles")
	}

	// Should contain at least some common styles
	commonStyles := []string{"monokai", "github"}
	for _, expected := range commonStyles {
		found := false
		for _, style := range styles {
			if style == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected to find style %q in available styles", expected)
		}
	}
}

// TestIsCodeFile tests code file detection
func TestIsCodeFile(t *testing.T) {
	tests := []struct {
		filename string
		want     bool
	}{
		{"test.go", true},
		{"test.py", true},
		{"test.js", true},
		{"test.xyz123", false}, // Unknown extension
		{"README", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := IsCodeFile(tt.filename)
			if got != tt.want {
				t.Errorf("IsCodeFile(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

// TestGetFileType tests file type description
func TestGetFileType(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"test.go", "Go"},
		{"test.py", "Python"},
		{"test.js", "JavaScript"},
		{"test.jsx", "JavaScript (React)"},
		{"test.ts", "TypeScript"},
		{"test.tsx", "TypeScript (React)"},
		{"test.md", "Markdown"},
		{"test.json", "JSON"},
		{"test.yaml", "YAML"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := GetFileType(tt.filename)
			if got != tt.want {
				t.Errorf("GetFileType(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

// TestNewGoParser tests Go parser creation
func TestNewGoParser(t *testing.T) {
	parser := NewGoParser()
	if parser == nil {
		t.Fatal("expected parser, got nil")
	}

	if parser.fset == nil {
		t.Error("expected file set to be initialized")
	}
}

// TestParseFile tests Go file parsing and symbol extraction
func TestParseFile(t *testing.T) {
	parser := NewGoParser()

	symbols, err := parser.ParseFile("test.go", []byte(sampleGoCode))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if len(symbols) == 0 {
		t.Fatal("expected symbols to be extracted")
	}

	// Check for package symbol
	pkg := FindSymbol(symbols, "main")
	if pkg == nil || pkg.Kind != SymbolPackage {
		t.Error("expected to find package symbol 'main'")
	}

	// Check for function
	fn := FindSymbol(symbols, "Greet")
	if fn == nil || fn.Kind != SymbolFunction {
		t.Error("expected to find function 'Greet'")
	}
	if fn != nil && !fn.Exported {
		t.Error("expected 'Greet' to be exported")
	}
	if fn != nil && fn.DocComment == "" {
		t.Error("expected doc comment for 'Greet'")
	}

	// Check for struct
	person := FindSymbol(symbols, "Person")
	if person == nil || person.Kind != SymbolStruct {
		t.Error("expected to find struct 'Person'")
	}

	// Check for struct fields
	if person != nil {
		if len(person.Children) != 2 {
			t.Errorf("expected 2 fields in Person, got %d", len(person.Children))
		}

		nameField := FindSymbol(person.Children, "Name")
		if nameField == nil || nameField.Kind != SymbolField {
			t.Error("expected to find field 'Name'")
		}
	}

	// Check for method
	stringMethod := FindSymbol(symbols, "String")
	if stringMethod == nil || stringMethod.Kind != SymbolMethod {
		t.Error("expected to find method 'String'")
	}
	if stringMethod != nil && stringMethod.Receiver != "*Person" {
		t.Errorf("expected receiver '*Person', got %q", stringMethod.Receiver)
	}
}

// TestParseFileWithInterface tests interface parsing
func TestParseFileWithInterface(t *testing.T) {
	code := `package main

type Reader interface {
	Read(p []byte) (n int, err error)
	io.Closer
}
`

	parser := NewGoParser()
	symbols, err := parser.ParseFile("test.go", []byte(code))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	reader := FindSymbol(symbols, "Reader")
	if reader == nil || reader.Kind != SymbolInterface {
		t.Fatal("expected to find interface 'Reader'")
	}

	// Should have one method + one embedded interface
	if len(reader.Children) != 2 {
		t.Errorf("expected 2 children, got %d", len(reader.Children))
	}
}

// TestParseFileWithConstants tests constant and variable parsing
func TestParseFileWithConstants(t *testing.T) {
	code := `package main

const Pi = 3.14
const MaxSize = 100

var GlobalVar string
var counter int = 0
`

	parser := NewGoParser()
	symbols, err := parser.ParseFile("test.go", []byte(code))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Check for constants
	pi := FindSymbol(symbols, "Pi")
	if pi == nil || pi.Kind != SymbolConstant {
		t.Error("expected to find constant 'Pi'")
	}

	// Check for variables
	globalVar := FindSymbol(symbols, "GlobalVar")
	if globalVar == nil || globalVar.Kind != SymbolVariable {
		t.Error("expected to find variable 'GlobalVar'")
	}
}

// TestParseFileWithInvalidCode tests parsing invalid Go code
func TestParseFileWithInvalidCode(t *testing.T) {
	invalidCode := `package main

func invalid syntax here {
`

	parser := NewGoParser()
	_, err := parser.ParseFile("test.go", []byte(invalidCode))
	if err == nil {
		t.Error("expected parse error for invalid code")
	}
}

// TestSymbolKindString tests SymbolKind string representation
func TestSymbolKindString(t *testing.T) {
	tests := []struct {
		kind SymbolKind
		want string
	}{
		{SymbolFunction, "function"},
		{SymbolMethod, "method"},
		{SymbolType, "type"},
		{SymbolStruct, "struct"},
		{SymbolInterface, "interface"},
		{SymbolVariable, "variable"},
		{SymbolConstant, "constant"},
		{SymbolField, "field"},
		{SymbolImport, "import"},
		{SymbolPackage, "package"},
		{SymbolKind(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.kind.String()
			if got != tt.want {
				t.Errorf("SymbolKind.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestSymbolKindIcon tests SymbolKind icon representation
func TestSymbolKindIcon(t *testing.T) {
	tests := []struct {
		kind SymbolKind
		want string
	}{
		{SymbolFunction, "ƒ"},
		{SymbolMethod, "m"},
		{SymbolType, "T"},
		{SymbolStruct, "S"},
		{SymbolInterface, "I"},
		{SymbolVariable, "v"},
		{SymbolConstant, "c"},
		{SymbolField, "."},
		{SymbolImport, "→"},
		{SymbolPackage, "P"},
		{SymbolKind(999), "?"},
	}

	for _, tt := range tests {
		t.Run(tt.kind.String(), func(t *testing.T) {
			got := tt.kind.Icon()
			if got != tt.want {
				t.Errorf("SymbolKind.Icon() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestFindSymbol tests symbol search functionality
func TestFindSymbol(t *testing.T) {
	symbols := []*Symbol{
		{Name: "foo", Kind: SymbolFunction},
		{
			Name: "bar",
			Kind: SymbolStruct,
			Children: []*Symbol{
				{Name: "field1", Kind: SymbolField},
				{Name: "field2", Kind: SymbolField},
			},
		},
	}

	// Find top-level symbol
	foo := FindSymbol(symbols, "foo")
	if foo == nil || foo.Name != "foo" {
		t.Error("expected to find symbol 'foo'")
	}

	// Find nested symbol
	field := FindSymbol(symbols, "field1")
	if field == nil || field.Name != "field1" {
		t.Error("expected to find nested symbol 'field1'")
	}

	// Search for non-existent symbol
	notFound := FindSymbol(symbols, "nonexistent")
	if notFound != nil {
		t.Error("expected nil for non-existent symbol")
	}
}

// TestFilterSymbols tests symbol filtering by kind
func TestFilterSymbols(t *testing.T) {
	symbols := []*Symbol{
		{Name: "func1", Kind: SymbolFunction},
		{Name: "func2", Kind: SymbolFunction},
		{Name: "var1", Kind: SymbolVariable},
		{Name: "const1", Kind: SymbolConstant},
		{Name: "type1", Kind: SymbolType},
	}

	// Filter functions only
	functions := FilterSymbols(symbols, SymbolFunction)
	if len(functions) != 2 {
		t.Errorf("expected 2 functions, got %d", len(functions))
	}

	// Filter multiple kinds
	vars := FilterSymbols(symbols, SymbolVariable, SymbolConstant)
	if len(vars) != 2 {
		t.Errorf("expected 2 variables/constants, got %d", len(vars))
	}

	// Filter with no matches
	methods := FilterSymbols(symbols, SymbolMethod)
	if len(methods) != 0 {
		t.Errorf("expected 0 methods, got %d", len(methods))
	}
}

// TestExportedSymbols tests filtering for exported symbols
func TestExportedSymbols(t *testing.T) {
	symbols := []*Symbol{
		{Name: "Public", Exported: true},
		{Name: "private", Exported: false},
		{Name: "AnotherPublic", Exported: true},
		{Name: "anotherPrivate", Exported: false},
	}

	exported := ExportedSymbols(symbols)
	if len(exported) != 2 {
		t.Errorf("expected 2 exported symbols, got %d", len(exported))
	}

	for _, sym := range exported {
		if !sym.Exported {
			t.Errorf("found non-exported symbol %q in exported list", sym.Name)
		}
	}
}

// TestFormatType tests type formatting
func TestFormatType(t *testing.T) {
	parser := NewGoParser()
	code := `package main

type MyType struct {
	ptr    *string
	slice  []int
	arr    [10]int
	mapVal map[string]int
}
`

	symbols, err := parser.ParseFile("test.go", []byte(code))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	myType := FindSymbol(symbols, "MyType")
	if myType == nil {
		t.Fatal("expected to find MyType")
	}

	// Check field type formatting
	tests := []struct {
		fieldName    string
		wantContains string
	}{
		{"ptr", "*"},
		{"slice", "[]"},
		{"mapVal", "map["},
	}

	for _, tt := range tests {
		field := FindSymbol(myType.Children, tt.fieldName)
		if field == nil {
			t.Errorf("expected to find field %q", tt.fieldName)
			continue
		}

		if !strings.Contains(field.Signature, tt.wantContains) {
			t.Errorf("field %q signature %q should contain %q",
				tt.fieldName, field.Signature, tt.wantContains)
		}
	}
}
