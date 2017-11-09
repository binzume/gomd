package main

import (
	"io"
)

// PlainWriter : impl for DocWriter
type PlainWriter struct {
	writer io.Writer
}

func NewPlainWriter(writer io.Writer) *PlainWriter {
	return &PlainWriter{writer}
}

func (w *PlainWriter) Heading(text string, level int) int {
	io.WriteString(w.writer, "\n")
	return 0
}

func (w *PlainWriter) Paragraph() int {
	io.WriteString(w.writer, "\n")
	return 0
}

func (w *PlainWriter) Link(url string, opt int) int {
	io.WriteString(w.writer, url)
	return 0
}

func (w *PlainWriter) Image(url string, opt int) int {
	io.WriteString(w.writer, url)
	return 0
}

func (w *PlainWriter) List() int {
	io.WriteString(w.writer, "\n")
	return 0
}

func (w *PlainWriter) ListItem() int {
	return 0
}

func (w *PlainWriter) Table() int {
	io.WriteString(w.writer, "\n")
	return 0
}

func (w *PlainWriter) TableRow() int {
	io.WriteString(w.writer, "\n")
	return 0
}

func (w *PlainWriter) TableCell() int {
	io.WriteString(w.writer, "\t")
	return 0
}

func (w *PlainWriter) Strike() int {
	return 0
}

func (w *PlainWriter) Bold() int {
	return 0
}

func (w *PlainWriter) Code() int {
	return 0
}

func (w *PlainWriter) CodeBlock(lang string) int {
	io.WriteString(w.writer, "\t")
	return 0
}

func (w *PlainWriter) WriteStyle(text string, className string, color string, flags int) {
	w.Write(text)
}

func (w *PlainWriter) Write(text string) {
	w.writer.Write([]byte(text))
}

func (w *PlainWriter) End(lv int) {
}

func (w *PlainWriter) Close() {
	w.End(0)
}