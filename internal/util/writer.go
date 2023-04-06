package util

import (
	"fmt"
	"strings"
)

type indent_type uint8

const (
	indent_line indent_type = iota
	indent_no_line
)

type IndentedWriter struct {
	indents []indent_type
	builder strings.Builder
}

func NewIndentWirter() *IndentedWriter {
	return &IndentedWriter{
		indents: make([]indent_type, 0, 10),
	}
}

func (i *IndentedWriter) Indent(show_vertical_line bool) {
	if show_vertical_line {
		i.indents = append(i.indents, indent_line)
	} else {

		i.indents = append(i.indents, indent_no_line)
	}
}
func (i *IndentedWriter) UnIndent() {
	if len(i.indents) > 0 {
		i.indents = i.indents[:len(i.indents)-1]
	}
}
func (i *IndentedWriter) WriteCloseIndent(msg string) (n int, err error) {
	i.builder.Grow((len(i.indents) * 3) + len(msg) + 2)
	i.addIndentsToBuilder(i.indents)
	i.builder.WriteRune('┗')
	i.builder.WriteRune('╸')
	i.builder.WriteString(msg)
	res, err := i.WriteNoIndent("")
	return res, err

}
func (i *IndentedWriter) WriteFirst(msg string) (n int, err error) {
	i.builder.Grow((len(i.indents) * 3) + len(msg) + 2)
	i.addIndentsToBuilder(i.indents)
	i.builder.WriteRune('┏')
	i.builder.WriteRune('╸')
	i.builder.WriteString(msg)
	return i.WriteNoIndent("")
}
func (i *IndentedWriter) WriteNormal(msg string) (n int, err error) {
	i.builder.Grow((len(i.indents) * 3) + len(msg) + 2)
	i.addIndentsToBuilder(i.indents)
	i.builder.WriteRune('┣')
	i.builder.WriteRune('╸')
	i.builder.WriteString(msg)
	return i.WriteNoIndent("")
}
func (i *IndentedWriter) WriteNormalLines(msgs []string) (n int, err error) {
	cnt := 0
	for _, v := range msgs {
		n, err := i.WriteNormal(v)
		cnt += n
		if err != nil {
			return cnt, err
		}
	}
	return cnt, nil
}
func (i *IndentedWriter) addIndentsToBuilder(indents []indent_type) {
	for _, v := range indents {
		if v == indent_line {
			i.builder.WriteRune('┃')
		} else {
			i.builder.WriteRune(' ')
		}
		i.builder.WriteRune(' ')
		i.builder.WriteRune(' ')
	}
}

func (i *IndentedWriter) WriteNoIndent(msg string) (n int, err error) {
	if msg == "" {
		msg = i.builder.String()
		i.builder.Reset()
	}
	return fmt.Print(msg)
}
