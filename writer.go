package main

type DocWriter interface {
	Heading(text string, level int) int
	Link(url string, title string, options int) int
	Image(url string, title string, options int) int
	Strike() int
	Strong() int
	Bold() int
	Italic() int
	Code() int
	Paragraph() int
	List(mode int) int
	ListItem() int
	Table() int
	TableRow() int
	TableCell(flags int) int
	CodeBlock(lang string) int
	End(lv int)
	Write(text string)
	WriteStyle(text string, className string, color string, flags int)
	Close()
}
