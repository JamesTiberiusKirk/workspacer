package codelist

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

	filterHighlightStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#FF0000")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true)
)

type styledSegment struct {
	text  string
	style string
}

func highlightFilteredText(text string, searchTerms []string, filterText string) string {
	segments := splitStyled(text)

	for _, term := range searchTerms {
		segments = highlightSegments(segments, term, lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true), false)
	}

	if filterText != "" {
		segments = highlightSegments(segments, filterText, lipgloss.NewStyle().Foreground(lipgloss.Color("red")).Bold(true), true)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, joinSegments(segments)...)
}

func joinSegments(segments []styledSegment) []string {
	var result []string
	for _, seg := range segments {
		if seg.style != "" {
			result = append(result, seg.style+seg.text)
		} else {
			result = append(result, seg.text)
		}
	}
	return result
}

func splitStyled(text string) []styledSegment {
	var segments []styledSegment
	var currentStyle, currentText strings.Builder
	inStyle := false

	for _, r := range text {
		if r == '\x1b' {
			if currentText.Len() > 0 {
				segments = append(segments, styledSegment{currentText.String(), currentStyle.String()})
				currentText.Reset()
			}
			currentStyle.Reset()
			currentStyle.WriteRune(r)
			inStyle = true
		} else if inStyle {
			currentStyle.WriteRune(r)
			if r == 'm' {
				inStyle = false
			}
		} else {
			currentText.WriteRune(r)
		}
	}

	if currentText.Len() > 0 {
		segments = append(segments, styledSegment{currentText.String(), currentStyle.String()})
	}

	return segments
}

func highlightSegments(segments []styledSegment, term string, highlightStyle lipgloss.Style, isFilter bool) []styledSegment {
	var result []styledSegment
	re := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(term))

	for _, seg := range segments {
		if seg.style != "" && isFilter {
			result = append(result, seg)
			continue
		}

		indices := re.FindAllStringIndex(seg.text, -1)
		lastIndex := 0
		for _, idx := range indices {
			if idx[0] > lastIndex {
				result = append(result, styledSegment{seg.text[lastIndex:idx[0]], seg.style})
			}
			highlightedText := highlightStyle.Render(seg.text[idx[0]:idx[1]])
			result = append(result, styledSegment{highlightedText, ""})
			lastIndex = idx[1]
		}
		if lastIndex < len(seg.text) {
			result = append(result, styledSegment{seg.text[lastIndex:], seg.style})
		}
	}

	return result
}

func highlightCode(code, language string) (string, []int) {
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

func wrapText(text string, width int) string {
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
