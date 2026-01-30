package components

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// MarkdownRenderer renders markdown content for terminal display
type MarkdownRenderer struct {
	width     int
	codeStyle lipgloss.Style
	h1Style   lipgloss.Style
	h2Style   lipgloss.Style
	h3Style   lipgloss.Style
	boldStyle lipgloss.Style
	linkStyle lipgloss.Style
	listStyle lipgloss.Style
	quoteStyle lipgloss.Style
}

// NewMarkdownRenderer creates a new markdown renderer
func NewMarkdownRenderer(width int) *MarkdownRenderer {
	return &MarkdownRenderer{
		width: width,
		codeStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("252")).
			Padding(0, 1),
		h1Style: lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true).
			Underline(true),
		h2Style: lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true),
		h3Style: lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true),
		boldStyle: lipgloss.NewStyle().
			Bold(true),
		linkStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Underline(true),
		listStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		quoteStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true).
			BorderLeft(true).
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color("240")).
			PaddingLeft(1),
	}
}

// SetWidth sets the rendering width
func (m *MarkdownRenderer) SetWidth(width int) {
	m.width = width
}

// Render renders markdown content
func (m *MarkdownRenderer) Render(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	inCodeBlock := false
	codeBlockContent := []string{}
	codeBlockLang := ""

	for _, line := range lines {
		// Handle code blocks
		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				// End of code block
				result = append(result, m.renderCodeBlock(codeBlockContent, codeBlockLang))
				codeBlockContent = []string{}
				codeBlockLang = ""
				inCodeBlock = false
			} else {
				// Start of code block
				inCodeBlock = true
				codeBlockLang = strings.TrimPrefix(line, "```")
			}
			continue
		}

		if inCodeBlock {
			codeBlockContent = append(codeBlockContent, line)
			continue
		}

		// Handle headers
		if strings.HasPrefix(line, "### ") {
			result = append(result, m.h3Style.Render(strings.TrimPrefix(line, "### ")))
			continue
		}
		if strings.HasPrefix(line, "## ") {
			result = append(result, m.h2Style.Render(strings.TrimPrefix(line, "## ")))
			continue
		}
		if strings.HasPrefix(line, "# ") {
			result = append(result, m.h1Style.Render(strings.TrimPrefix(line, "# ")))
			continue
		}

		// Handle blockquotes
		if strings.HasPrefix(line, "> ") {
			result = append(result, m.quoteStyle.Render(strings.TrimPrefix(line, "> ")))
			continue
		}

		// Handle bullet lists
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			bullet := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("•")
			text := strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* ")
			result = append(result, "  "+bullet+" "+m.renderInline(text))
			continue
		}

		// Handle numbered lists
		if matched, _ := regexp.MatchString(`^\d+\. `, line); matched {
			re := regexp.MustCompile(`^(\d+)\. (.*)$`)
			matches := re.FindStringSubmatch(line)
			if len(matches) == 3 {
				num := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render(matches[1] + ".")
				result = append(result, "  "+num+" "+m.renderInline(matches[2]))
				continue
			}
		}

		// Handle horizontal rules
		if line == "---" || line == "***" || line == "___" {
			result = append(result, strings.Repeat("─", m.width-4))
			continue
		}

		// Regular paragraph
		result = append(result, m.renderInline(line))
	}

	// Handle unclosed code block
	if inCodeBlock && len(codeBlockContent) > 0 {
		result = append(result, m.renderCodeBlock(codeBlockContent, codeBlockLang))
	}

	return strings.Join(result, "\n")
}

// renderCodeBlock renders a code block
func (m *MarkdownRenderer) renderCodeBlock(lines []string, language string) string {
	content := strings.Join(lines, "\n")

	// Add language indicator if present
	var header string
	if language != "" {
		langStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)
		header = langStyle.Render(language) + "\n"
	}

	codeStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("252")).
		Padding(0, 1).
		Width(m.width - 4)

	return header + codeStyle.Render(content)
}

// renderInline handles inline markdown formatting
func (m *MarkdownRenderer) renderInline(text string) string {
	// Handle inline code
	codeRe := regexp.MustCompile("`([^`]+)`")
	text = codeRe.ReplaceAllStringFunc(text, func(match string) string {
		code := strings.Trim(match, "`")
		return m.codeStyle.Render(code)
	})

	// Handle bold
	boldRe := regexp.MustCompile(`\*\*([^*]+)\*\*`)
	text = boldRe.ReplaceAllStringFunc(text, func(match string) string {
		content := strings.Trim(match, "*")
		return m.boldStyle.Render(content)
	})

	// Handle bold with __
	boldRe2 := regexp.MustCompile(`__([^_]+)__`)
	text = boldRe2.ReplaceAllStringFunc(text, func(match string) string {
		content := strings.Trim(match, "_")
		return m.boldStyle.Render(content)
	})

	// Handle italic
	italicRe := regexp.MustCompile(`\*([^*]+)\*`)
	text = italicRe.ReplaceAllStringFunc(text, func(match string) string {
		content := strings.Trim(match, "*")
		return lipgloss.NewStyle().Italic(true).Render(content)
	})

	// Handle links [text](url)
	linkRe := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	text = linkRe.ReplaceAllStringFunc(text, func(match string) string {
		matches := linkRe.FindStringSubmatch(match)
		if len(matches) == 3 {
			return m.linkStyle.Render(matches[1])
		}
		return match
	})

	return text
}

// RenderSimple renders content with minimal formatting
func (m *MarkdownRenderer) RenderSimple(content string) string {
	// Just handle basic inline formatting
	return m.renderInline(content)
}

// CodeBlock creates a formatted code block
func CodeBlock(code, language string, width int) string {
	renderer := NewMarkdownRenderer(width)
	return renderer.renderCodeBlock(strings.Split(code, "\n"), language)
}

// InlineCode formats inline code
func InlineCode(code string) string {
	style := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("252")).
		Padding(0, 1)
	return style.Render(code)
}

// Bold formats text as bold
func Bold(text string) string {
	return lipgloss.NewStyle().Bold(true).Render(text)
}

// Italic formats text as italic
func Italic(text string) string {
	return lipgloss.NewStyle().Italic(true).Render(text)
}

// Header formats a header with the specified level
func Header(text string, level int) string {
	var style lipgloss.Style
	switch level {
	case 1:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true).
			Underline(true)
	case 2:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true)
	case 3:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true)
	default:
		style = lipgloss.NewStyle().Bold(true)
	}
	return style.Render(text)
}

// BulletList creates a formatted bullet list
func BulletList(items []string) string {
	var result []string
	bulletStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))

	for _, item := range items {
		result = append(result, "  "+bulletStyle.Render("•")+" "+item)
	}

	return strings.Join(result, "\n")
}

// NumberedList creates a formatted numbered list
func NumberedList(items []string) string {
	var result []string
	numStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))

	for i, item := range items {
		result = append(result, "  "+numStyle.Render(string(rune('1'+i))+".")+" "+item)
	}

	return strings.Join(result, "\n")
}

// Quote formats a blockquote
func Quote(text string) string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true).
		BorderLeft(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color("240")).
		PaddingLeft(1)
	return style.Render(text)
}
