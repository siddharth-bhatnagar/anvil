package analysis

import (
	"bytes"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// SyntaxHighlighter provides syntax highlighting for code
type SyntaxHighlighter struct {
	style     *chroma.Style
	formatter chroma.Formatter
}

// NewSyntaxHighlighter creates a new syntax highlighter
func NewSyntaxHighlighter() *SyntaxHighlighter {
	// Use a terminal-friendly style
	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}

	// Use terminal256 formatter for color output
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	return &SyntaxHighlighter{
		style:     style,
		formatter: formatter,
	}
}

// NewSyntaxHighlighterWithStyle creates a highlighter with a specific style
func NewSyntaxHighlighterWithStyle(styleName string) *SyntaxHighlighter {
	style := styles.Get(styleName)
	if style == nil {
		style = styles.Fallback
	}

	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	return &SyntaxHighlighter{
		style:     style,
		formatter: formatter,
	}
}

// Highlight highlights code and returns the colored output
func (sh *SyntaxHighlighter) Highlight(code string, filename string) (string, error) {
	// Get lexer based on filename
	lexer := lexers.Match(filename)
	if lexer == nil {
		// Try to analyze content
		lexer = lexers.Analyse(code)
	}
	if lexer == nil {
		// Fall back to plaintext
		lexer = lexers.Fallback
	}

	// Coalesce runs of same token type
	lexer = chroma.Coalesce(lexer)

	// Tokenize the code
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return code, err
	}

	// Format the tokens
	var buf bytes.Buffer
	err = sh.formatter.Format(&buf, sh.style, iterator)
	if err != nil {
		return code, err
	}

	return buf.String(), nil
}

// HighlightWithLanguage highlights code with an explicit language
func (sh *SyntaxHighlighter) HighlightWithLanguage(code string, language string) (string, error) {
	lexer := lexers.Get(language)
	if lexer == nil {
		lexer = lexers.Fallback
	}

	lexer = chroma.Coalesce(lexer)

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return code, err
	}

	var buf bytes.Buffer
	err = sh.formatter.Format(&buf, sh.style, iterator)
	if err != nil {
		return code, err
	}

	return buf.String(), nil
}

// HighlightLines highlights specific lines of code
func (sh *SyntaxHighlighter) HighlightLines(code string, filename string, startLine, endLine int) (string, error) {
	lines := strings.Split(code, "\n")

	// Adjust bounds
	if startLine < 1 {
		startLine = 1
	}
	if endLine > len(lines) {
		endLine = len(lines)
	}
	if startLine > endLine {
		return "", nil
	}

	// Extract the requested lines
	selectedLines := lines[startLine-1 : endLine]
	selectedCode := strings.Join(selectedLines, "\n")

	return sh.Highlight(selectedCode, filename)
}

// GetLanguage returns the detected language for a file
func GetLanguage(filename string) string {
	lexer := lexers.Match(filename)
	if lexer == nil {
		return "plaintext"
	}

	config := lexer.Config()
	if config != nil {
		return config.Name
	}

	return "plaintext"
}

// GetLanguageByExtension returns the language for a file extension
func GetLanguageByExtension(ext string) string {
	// Ensure extension has a dot
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	// Create a fake filename with the extension
	filename := "file" + ext
	return GetLanguage(filename)
}

// SupportedLanguages returns a list of supported languages
func SupportedLanguages() []string {
	var languages []string
	for _, lexer := range lexers.GlobalLexerRegistry.Lexers {
		config := lexer.Config()
		if config != nil {
			languages = append(languages, config.Name)
		}
	}
	return languages
}

// AvailableStyles returns a list of available color styles
func AvailableStyles() []string {
	return styles.Names()
}

// IsCodeFile returns true if the file appears to be a code file
func IsCodeFile(filename string) bool {
	// Check if we can find a lexer for this file
	lexer := lexers.Match(filename)
	return lexer != nil && lexer != lexers.Fallback
}

// GetFileType returns a human-readable file type description
func GetFileType(filename string) string {
	ext := filepath.Ext(filename)

	// Common mappings
	typeMap := map[string]string{
		".go":    "Go",
		".py":    "Python",
		".js":    "JavaScript",
		".ts":    "TypeScript",
		".jsx":   "JavaScript (React)",
		".tsx":   "TypeScript (React)",
		".rs":    "Rust",
		".java":  "Java",
		".c":     "C",
		".cpp":   "C++",
		".h":     "C Header",
		".hpp":   "C++ Header",
		".rb":    "Ruby",
		".php":   "PHP",
		".swift": "Swift",
		".kt":    "Kotlin",
		".scala": "Scala",
		".sh":    "Shell",
		".bash":  "Bash",
		".zsh":   "Zsh",
		".fish":  "Fish",
		".sql":   "SQL",
		".html":  "HTML",
		".css":   "CSS",
		".scss":  "SCSS",
		".less":  "Less",
		".json":  "JSON",
		".yaml":  "YAML",
		".yml":   "YAML",
		".toml":  "TOML",
		".xml":   "XML",
		".md":    "Markdown",
		".txt":   "Plain Text",
	}

	if fileType, ok := typeMap[ext]; ok {
		return fileType
	}

	return GetLanguage(filename)
}
