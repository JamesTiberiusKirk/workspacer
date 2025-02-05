package util

import (
	"regexp"
	"strings"

	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/charmbracelet/lipgloss"
)

var (
	highlightStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("205")). // Pink background
		Foreground(lipgloss.Color("#ffffff")).
		Bold(true)
)

func HighlightCode(code, language string) (string, []int) {
	lexer := lexers.Get(language)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return code, []int{}
	}
	var buf strings.Builder
	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		return code, []int{}
	}

	highlightedCode := buf.String()
	ansiPositions := []int{}
	for i, r := range highlightedCode {
		if r == '\x1b' {
			ansiPositions = append(ansiPositions, i)
		}
	}

	return highlightedCode, ansiPositions
}

func HighlightSearchTerms(code string, searchTerms []string) string {
	type segment struct {
		text   string
		isANSI bool
	}

	// Split the code into ANSI and non-ANSI segments
	var segments []segment
	ansiRegex := regexp.MustCompile("\x1b\\[[0-9;]*m")
	indices := ansiRegex.FindAllStringIndex(code, -1)
	lastIndex := 0
	for _, idx := range indices {
		if idx[0] > lastIndex {
			segments = append(segments, segment{code[lastIndex:idx[0]], false})
		}
		segments = append(segments, segment{code[idx[0]:idx[1]], true})
		lastIndex = idx[1]
	}
	if lastIndex < len(code) {
		segments = append(segments, segment{code[lastIndex:], false})
	}

	// Highlight search terms in non-ANSI segments
	for i, seg := range segments {
		if seg.isANSI {
			continue
		}
		for _, term := range searchTerms {
			re := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(term))
			seg.text = re.ReplaceAllStringFunc(seg.text, func(match string) string {
				return highlightStyle.Render(match)
			})
		}
		segments[i] = seg
	}

	// Reconstruct the highlighted code
	var result strings.Builder
	for _, seg := range segments {
		result.WriteString(seg.text)
	}

	return result.String()
}

func WrapText(text string, width int) string {
	lines := strings.Split(text, "\n")
	var wrappedLines []string

	for _, line := range lines {
		if len(line) <= width {
			wrappedLines = append(wrappedLines, line)
			continue
		}

		wrappedLines = append(wrappedLines, line[0:width-1])
		wl := line[width:]
		for {
			if len(wl) < width {
				wrappedLines = append(wrappedLines, wl)
				break
			}
			wl = wl[:width-1]
			wrappedLines = append(wrappedLines, wl)
		}
	}

	return strings.Join(wrappedLines, "\n")
}
