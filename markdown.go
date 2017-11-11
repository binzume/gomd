package markdown

import (
	"bufio"
	"regexp"
	"strings"
)

type Source struct {
	scanner *bufio.Scanner
	retry   bool
}

func (s *Source) Text() string {
	return s.scanner.Text()
}

func (s *Source) Scan() bool {
	if s.retry {
		s.retry = false
		return true
	}
	return s.scanner.Scan()
}

func (s *Source) Retry() {
	s.retry = true
}

type Matcher interface {
	Prefix() string
	TryMatch(line string) (int, []string)
	Render(params []string, scanner *Source, writer DocWriter)
}

type SimpleInlineMatcher struct {
	Start      string
	End        string
	RenderFunc func(content string, writer DocWriter, matcher *SimpleInlineMatcher)
}

func (m *SimpleInlineMatcher) Prefix() string {
	return m.Start
}

func (m *SimpleInlineMatcher) TryMatch(line string) (int, []string) {
	line = line[len(m.Start):]
	p := strings.Index(line, m.End)
	if p < 0 {
		return -1, nil
	}
	return p + len(m.End) + len(m.Start), []string{line[:p]}
}

func (m *SimpleInlineMatcher) Render(params []string, scanner *Source, writer DocWriter) {
	m.RenderFunc(params[0], writer, m)
}

type RegexMatcher struct {
	Re         *regexp.Regexp
	RenderFunc func(matches []string, writer DocWriter, scanner *Source, matcher *RegexMatcher)
	mode       int
}

func (m *RegexMatcher) Prefix() string {
	return ""
}

func (m *RegexMatcher) TryMatch(text string) (int, []string) {
	match := m.Re.FindStringSubmatch(text)
	if len(match) < 1 {
		return -1, nil
	}
	return len(match[0]), match
}

func (m *RegexMatcher) Render(params []string, scanner *Source, writer DocWriter) {
	m.RenderFunc(params, writer, scanner, m)
}

type LinkInlineMatcher struct{}

func (m *LinkInlineMatcher) Prefix() string {
	return "["
}

func (m *LinkInlineMatcher) TryMatch(line string) (int, []string) {
	p1 := strings.Index(line, "](")
	if p1 < 0 {
		return -1, nil
	}
	p2 := strings.Index(line, ")")
	if p2 < 0 || p2 < p1 {
		return -1, nil
	}
	return p2 + 1, []string{line[1:p1], line[p1+2 : p2]}
}

func (m *LinkInlineMatcher) Render(params []string, scanner *Source, writer DocWriter) {
	n := writer.Link(strings.Split(params[1], " ")[0], "", 0)
	writer.Write(params[0])
	writer.End(n)
}

func strike(text string, writer DocWriter, markup *SimpleInlineMatcher) {
	a := writer.Strike()
	inline(text, writer)
	writer.End(a)
}

func strong(text string, writer DocWriter, markup *SimpleInlineMatcher) {
	a := writer.Strong()
	inline(text, writer)
	writer.End(a)
}

func bold(text string, writer DocWriter, markup *SimpleInlineMatcher) {
	a := writer.Bold()
	inline(text, writer)
	writer.End(a)
}

func emphasis(text string, writer DocWriter, markup *SimpleInlineMatcher) {
	a := writer.Emphasis()
	inline(text, writer)
	writer.End(a)
}

func icode(text string, writer DocWriter, markup *SimpleInlineMatcher) {
	a := writer.Code()
	writer.Write(text)
	writer.End(a)
}

func autolink(params []string, writer DocWriter, scanner *Source, markup *RegexMatcher) {
	n := writer.Link(params[0], "", 0)
	writer.Write(params[0])
	writer.End(n)
}

func list(params []string, writer DocWriter, scanner *Source, markup *RegexMatcher) {
	var mode int
	switch params[2] {
	case "*", "-", "+":
		mode = 0
	default:
		mode = 1
	}
	nt := writer.List(mode)
	defer writer.End(nt)

	ni := writer.ListItem()
	indent := len(params[1])
	inline(params[3], writer)
	writer.End(ni)

	for scanner.Scan() {
		text := scanner.Text()
		params = markup.Re.FindStringSubmatch(text)
		if len(params) < 1 || len(params[1]) < indent {
			scanner.Retry()
			break
		} else if len(params[1]) > indent {
			list(params, writer, scanner, markup)
			continue
		}
		indent = len(params[1])

		ni := writer.ListItem()
		inline(params[3], writer)
		writer.End(ni)
	}
}

func code(params []string, writer DocWriter, scanner *Source, markup *RegexMatcher) {
	lang := params[1]
	n := writer.CodeBlock(lang, params[2])
	defer writer.End(n)

	for scanner.Scan() {
		text := scanner.Text()
		m := markup.Re.FindStringSubmatch(text)
		if len(m) > 0 {
			break
		}
		tokenizer := NewTokenizer(lang)
		tokenizer.Code(text)
		for typ, s := tokenizer.Read(); typ != CODE_EOF; typ, s = tokenizer.Read() {
			switch typ {
			case CODE_Keyword:
				writer.WriteStyle(s, "code_key", "", 0)
			case CODE_Number:
				writer.WriteStyle(s, "code_num", "", 0)
			case CODE_String:
				writer.WriteStyle(s, "code_str", "", 0)
			case CODE_Comment:
				writer.WriteStyle(s, "code_comment", "", 0)
			case CODE_Ident:
				writer.WriteStyle(s, "code_ident", "", 0)
			default:
				writer.Write(s)
			}
		}
		writer.Write("\n")
	}
}

func table(params []string, writer DocWriter, scanner *Source, markup *RegexMatcher) {
	align := make(map[int]int, len(params))
	if scanner.Scan() {
		t := markup.Re.FindStringSubmatch(scanner.Text())
		if len(t) > 0 {
			for i, s := range strings.Split(t[1], "|") {
				if strings.Trim(s, " -:|") != "" {
					scanner.Retry()
					break
				}
				if strings.HasPrefix(s, ":") {
					align[i] |= 1
				}
				if strings.HasSuffix(s, ":") {
					align[i] |= 2
				}
			}
		} else {
			scanner.Retry()
		}
	}
	nt := writer.Table()
	h := 4
	for scanner.Scan() {
		text := params[1]
		nr := writer.TableRow()
		for i, s := range strings.Split(text, "|") {
			nc := writer.TableCell(align[i] | h)
			inline(s, writer)
			writer.End(nc)
		}
		writer.End(nr)
		h = 0

		text = scanner.Text()
		params = markup.Re.FindStringSubmatch(text)
		if len(params) < 1 {
			scanner.Retry()
			break
		}
	}
	writer.End(nt)
}

func quote(params []string, writer DocWriter, scanner *Source, markup *RegexMatcher) {
	n := writer.QuoteBlock()
	defer writer.End(n)

	inline(params[1]+"\n", writer)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			break
		}
		params := markup.Re.FindStringSubmatch(text)
		if len(params) > 0 {
			text = params[1]
		}
		inline(text+"\n", writer)
	}
}

func heading(params []string, writer DocWriter, scanner *Source, markup *RegexMatcher) {
	writer.Heading(params[2], len(params[1]))
}

func comment(params []string, writer DocWriter, scanner *Source, markup *RegexMatcher) {
}

func hr(params []string, writer DocWriter, scanner *Source, markup *RegexMatcher) {
	writer.Hr()
}

func checkbox(params []string, writer DocWriter, scanner *Source, markup *RegexMatcher) {
	writer.CheckBox(params[1] == "x")
	inline(params[2], writer)
}

func pluginBlock(params []string, writer DocWriter, scanner *Source, markup *RegexMatcher) {
	// TODO
	n := writer.CodeBlock("", "")
	for scanner.Scan() {
		text := scanner.Text()
		if text == "}" {
			break
		}
	}
	writer.End(n)
}

var defaultInlineElems []Matcher
var defaultBlockElems []Matcher

func init() {
	defaultInlineElems = []Matcher{
		&SimpleInlineMatcher{"~~", "~~", strike},
		&SimpleInlineMatcher{"**", "**", strong},
		&SimpleInlineMatcher{"*", "*", emphasis},
		&SimpleInlineMatcher{"``", "``", icode},
		&SimpleInlineMatcher{"`", "`", icode},
		&SimpleInlineMatcher{"__", "__", strong},
		&LinkInlineMatcher{},
		&RegexMatcher{regexp.MustCompile(`^https?:[^\s\"\'\)<>]+`), autolink, 0},
	}
	defaultBlockElems = []Matcher{
		&RegexMatcher{regexp.MustCompile(`^(#{1,4})\s*(.*)`), heading, 0},
		&RegexMatcher{regexp.MustCompile(`^>+\s?(.*)`), quote, 0},
		&RegexMatcher{regexp.MustCompile("^```\\s*(\\w*)(:.*)?$"), code, 0},
		&RegexMatcher{regexp.MustCompile(`^\|(.+)\|$`), table, 0},
		&RegexMatcher{regexp.MustCompile(`^(\s*)(-|\*|\+|\d+\.)\s(.+)$`), list, 0},
		&RegexMatcher{regexp.MustCompile(`^&(\w+)[{]*$`), pluginBlock, 0},
		&RegexMatcher{regexp.MustCompile(`^([-_]\s?){3,}$`), hr, 0},
		&RegexMatcher{regexp.MustCompile(`^\[([x ])\](.*)`), checkbox, 0}, // todo inline
		&RegexMatcher{regexp.MustCompile(`^//.*`), comment, 0},
	}
}

type Markdown struct {
	inlineElems []Matcher
	blockElems  []Matcher
}

func NewMarkdown() *Markdown {
	return &Markdown{defaultInlineElems, defaultBlockElems}
}

func inline(text string, writer DocWriter) {
	for pos := 0; pos < len(text); pos++ {
		for _, markup := range defaultInlineElems {
			// TODO more fast.
			if strings.HasPrefix(text[pos:], markup.Prefix()) {
				l, params := markup.TryMatch(text[pos:])
				if l > 0 {
					if pos > 0 && text[pos-1] == '\\' {
						// Escaped
						writer.Write(text[:pos-1])
						text = text[pos:]
						pos = len(markup.Prefix())
						continue
					}
					writer.Write(text[:pos])
					markup.Render(params, nil, writer)
					text = text[(pos + l):]
					pos = 0
				}
			}
		}
	}
	writer.Write(text)
}

func (m *Markdown) block(scanner *Source, writer DocWriter) {
	para := 0
	for scanner.Scan() {
		text := scanner.Text()
		for _, matcher := range m.blockElems {
			l, params := matcher.TryMatch(text)
			if l > 0 {
				if para != 0 {
					writer.End(para)
					para = 0
				}
				writer.Write("\n")
				matcher.Render(params, scanner, writer)
				text = ""
				break
			}
		}
		if text == "" {
			if para != 0 {
				writer.End(para)
				para = 0
			}
			continue
		}
		if para == 0 {
			para = writer.Paragraph()
		}
		inline(text, writer)
	}
}

// ToHTML convert md to html
func Convert(scanner0 *bufio.Scanner, writer DocWriter) error {
	NewMarkdown().block(&Source{scanner: scanner0}, writer)
	return scanner0.Err()
}
