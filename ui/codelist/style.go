package codelist

import (
	"regexp"
	"sort"
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

func highlightWords2(text string, words []string, highlightStyle lipgloss.Style) string {
	// Sort words by length in descending order to handle longer phrases first
	sort.Slice(words, func(i, j int) bool {
		return len(words[i]) > len(words[j])
	})

	// Create a regular expression to match any of the words
	wordPatterns := make([]string, len(words))
	for i, word := range words {
		wordPatterns[i] = `\b(` + regexp.QuoteMeta(word) + `)\b`
	}
	re := regexp.MustCompile(`(?i)` + strings.Join(wordPatterns, "|"))

	// Split the text into segments (matched words and non-matched parts)
	segments := re.Split(text, -1)
	matches := re.FindAllString(text, -1)

	// Combine segments and matches, applying highlight style to matches
	var result strings.Builder
	for i, segment := range segments {
		result.WriteString(segment)
		if i < len(matches) {
			result.WriteString(highlightStyle.Render(matches[i]))
		}
	}

	return result.String()
}

func highlightWords(text string, highlightTerms []string, style lipgloss.Style) string {
	for _, t := range highlightTerms {
		re := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(t))
		text = re.ReplaceAllStringFunc(text, func(s string) string {
			return style.Render(s)
		})
	}

	return text
}

type styledSegment struct {
	text  string
	style string
}

func highlightFilteredText(text string, searchTerms []string, filterText string) string {
	segments := splitStyled(text)

	for _, term := range searchTerms {
		segments = highlightSegments(segments, term,
			lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true), false)
	}

	if filterText != "" {
		segments = highlightSegments(segments, filterText,
			lipgloss.NewStyle().Foreground(lipgloss.Color("red")).Bold(true), true)
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
		if isFilter && seg.style != "" { // Skip already-styled segments for filters
			result = append(result, seg)
			continue
		}

		matches := re.FindAllStringIndex(seg.text, -1)
		lastIndex := 0

		for _, match := range matches {
			if match[0] > lastIndex {
				result = append(result, styledSegment{text: seg.text[lastIndex:match[0]], style: seg.style})
			}
			result = append(result, styledSegment{text: highlightStyle.Render(seg.text[match[0]:match[1]]), style: ""})
			lastIndex = match[1]
		}

		if lastIndex < len(seg.text) {
			result = append(result, styledSegment{text: seg.text[lastIndex:], style: seg.style})
		}
	}

	return result
}

func highlightCode(code, language string) string {
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
		return code
	}
	var buf strings.Builder
	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		return code
	}

	return buf.String()
}
