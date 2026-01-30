package util

import (
	"fmt"
	"strings"
)

// DiffType represents the type of diff operation
type DiffType int

const (
	DiffEqual DiffType = iota
	DiffInsert
	DiffDelete
)

// DiffChunk represents a chunk of diff
type DiffChunk struct {
	Type DiffType
	Text string
}

// UnifiedDiff generates a unified diff between two texts
func UnifiedDiff(oldText, newText, oldName, newName string) string {
	oldLines := strings.Split(oldText, "\n")
	newLines := strings.Split(newText, "\n")

	// Use LCS-based diff algorithm
	chunks := diffLines(oldLines, newLines)

	return formatUnifiedDiff(chunks, oldName, newName, oldLines, newLines)
}

// diffLines computes the diff between two sets of lines
func diffLines(oldLines, newLines []string) []DiffChunk {
	// Simple Myers diff implementation
	m, n := len(oldLines), len(newLines)

	// Build LCS matrix
	lcs := make([][]int, m+1)
	for i := range lcs {
		lcs[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if oldLines[i-1] == newLines[j-1] {
				lcs[i][j] = lcs[i-1][j-1] + 1
			} else {
				if lcs[i-1][j] > lcs[i][j-1] {
					lcs[i][j] = lcs[i-1][j]
				} else {
					lcs[i][j] = lcs[i][j-1]
				}
			}
		}
	}

	// Backtrack to find diff
	var chunks []DiffChunk
	i, j := m, n

	for i > 0 || j > 0 {
		if i > 0 && j > 0 && oldLines[i-1] == newLines[j-1] {
			chunks = append([]DiffChunk{{Type: DiffEqual, Text: oldLines[i-1]}}, chunks...)
			i--
			j--
		} else if j > 0 && (i == 0 || lcs[i][j-1] >= lcs[i-1][j]) {
			chunks = append([]DiffChunk{{Type: DiffInsert, Text: newLines[j-1]}}, chunks...)
			j--
		} else if i > 0 {
			chunks = append([]DiffChunk{{Type: DiffDelete, Text: oldLines[i-1]}}, chunks...)
			i--
		}
	}

	return chunks
}

// formatUnifiedDiff formats chunks as unified diff
func formatUnifiedDiff(chunks []DiffChunk, oldName, newName string, oldLines, newLines []string) string {
	var result strings.Builder

	// Header
	result.WriteString(fmt.Sprintf("--- %s\n", oldName))
	result.WriteString(fmt.Sprintf("+++ %s\n", newName))

	// Group chunks into hunks
	hunks := groupIntoHunks(chunks, 3)

	for _, hunk := range hunks {
		result.WriteString(hunk)
	}

	return result.String()
}

// groupIntoHunks groups diff chunks into hunks with context
func groupIntoHunks(chunks []DiffChunk, contextLines int) []string {
	var hunks []string

	// Find ranges of changes
	type hunkRange struct {
		startOld, endOld int
		startNew, endNew int
		chunks           []DiffChunk
	}

	var ranges []hunkRange
	var currentRange *hunkRange
	oldLineNum, newLineNum := 0, 0

	for i, chunk := range chunks {
		switch chunk.Type {
		case DiffEqual:
			oldLineNum++
			newLineNum++
			if currentRange != nil {
				currentRange.chunks = append(currentRange.chunks, chunk)
				currentRange.endOld = oldLineNum
				currentRange.endNew = newLineNum

				// Check if we should close this range
				nextChangeIdx := -1
				for j := i + 1; j < len(chunks); j++ {
					if chunks[j].Type != DiffEqual {
						nextChangeIdx = j
						break
					}
				}

				if nextChangeIdx == -1 || nextChangeIdx-i > contextLines*2 {
					// Close the range
					ranges = append(ranges, *currentRange)
					currentRange = nil
				}
			}

		case DiffDelete:
			if currentRange == nil {
				// Start new range with context
				start := max(0, oldLineNum-contextLines)
				currentRange = &hunkRange{
					startOld: start,
					startNew: max(0, newLineNum-contextLines),
					chunks:   []DiffChunk{},
				}
				// Add context lines
				for j := start; j < oldLineNum; j++ {
					if j < i {
						currentRange.chunks = append(currentRange.chunks, chunks[j])
					}
				}
			}
			currentRange.chunks = append(currentRange.chunks, chunk)
			currentRange.endOld = oldLineNum + 1
			currentRange.endNew = newLineNum
			oldLineNum++

		case DiffInsert:
			if currentRange == nil {
				start := max(0, oldLineNum-contextLines)
				currentRange = &hunkRange{
					startOld: start,
					startNew: max(0, newLineNum-contextLines),
					chunks:   []DiffChunk{},
				}
				for j := start; j < oldLineNum; j++ {
					if j < i {
						currentRange.chunks = append(currentRange.chunks, chunks[j])
					}
				}
			}
			currentRange.chunks = append(currentRange.chunks, chunk)
			currentRange.endOld = oldLineNum
			currentRange.endNew = newLineNum + 1
			newLineNum++
		}
	}

	// Close final range if open
	if currentRange != nil {
		ranges = append(ranges, *currentRange)
	}

	// Format each range as a hunk
	for _, r := range ranges {
		var hunk strings.Builder

		oldCount := 0
		newCount := 0
		for _, c := range r.chunks {
			switch c.Type {
			case DiffEqual:
				oldCount++
				newCount++
			case DiffDelete:
				oldCount++
			case DiffInsert:
				newCount++
			}
		}

		hunk.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n",
			r.startOld+1, oldCount,
			r.startNew+1, newCount))

		for _, c := range r.chunks {
			switch c.Type {
			case DiffEqual:
				hunk.WriteString(" " + c.Text + "\n")
			case DiffDelete:
				hunk.WriteString("-" + c.Text + "\n")
			case DiffInsert:
				hunk.WriteString("+" + c.Text + "\n")
			}
		}

		hunks = append(hunks, hunk.String())
	}

	return hunks
}

// SimpleDiff generates a simple side-by-side comparison
func SimpleDiff(oldText, newText string) string {
	oldLines := strings.Split(oldText, "\n")
	newLines := strings.Split(newText, "\n")

	var result strings.Builder

	maxLines := len(oldLines)
	if len(newLines) > maxLines {
		maxLines = len(newLines)
	}

	for i := 0; i < maxLines; i++ {
		var oldLine, newLine string
		if i < len(oldLines) {
			oldLine = oldLines[i]
		}
		if i < len(newLines) {
			newLine = newLines[i]
		}

		if oldLine == newLine {
			result.WriteString(fmt.Sprintf("  %s\n", oldLine))
		} else {
			if oldLine != "" {
				result.WriteString(fmt.Sprintf("- %s\n", oldLine))
			}
			if newLine != "" {
				result.WriteString(fmt.Sprintf("+ %s\n", newLine))
			}
		}
	}

	return result.String()
}

// CountChanges counts the number of added and removed lines
func CountChanges(diff string) (added, removed int) {
	lines := strings.Split(diff, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			added++
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			removed++
		}
	}
	return
}

// HasChanges returns true if the diff contains any changes
func HasChanges(diff string) bool {
	added, removed := CountChanges(diff)
	return added > 0 || removed > 0
}

// ApplyPatch applies a unified diff patch to text
// This is a simplified implementation that handles basic cases
func ApplyPatch(original, patch string) (string, error) {
	lines := strings.Split(original, "\n")
	patchLines := strings.Split(patch, "\n")

	var result []string
	lineNum := 0
	patchIdx := 0

	// Skip header lines
	for patchIdx < len(patchLines) {
		if strings.HasPrefix(patchLines[patchIdx], "@@") {
			break
		}
		patchIdx++
	}

	for patchIdx < len(patchLines) {
		line := patchLines[patchIdx]

		if strings.HasPrefix(line, "@@") {
			// Parse hunk header
			var oldStart, oldCount int
			fmt.Sscanf(line, "@@ -%d,%d", &oldStart, &oldCount)

			// Copy lines up to hunk start
			for lineNum < oldStart-1 && lineNum < len(lines) {
				result = append(result, lines[lineNum])
				lineNum++
			}
			patchIdx++
			continue
		}

		if strings.HasPrefix(line, "+") {
			// Add new line
			result = append(result, line[1:])
		} else if strings.HasPrefix(line, "-") {
			// Skip removed line
			lineNum++
		} else if strings.HasPrefix(line, " ") {
			// Context line
			result = append(result, line[1:])
			lineNum++
		}

		patchIdx++
	}

	// Copy remaining lines
	for lineNum < len(lines) {
		result = append(result, lines[lineNum])
		lineNum++
	}

	return strings.Join(result, "\n"), nil
}

// DiffStats holds statistics about a diff
type DiffStats struct {
	FilesChanged int
	Additions    int
	Deletions    int
}

// FormatDiffStats formats diff statistics as a string
func (s DiffStats) String() string {
	return fmt.Sprintf("%d file(s) changed, %d insertion(s)(+), %d deletion(s)(-)",
		s.FilesChanged, s.Additions, s.Deletions)
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
