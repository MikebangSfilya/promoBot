package formatter

import (
	"fmt"
	"strings"
)

const (
	maxCodeLength = 22
	lineOverhead  = 6
	maxLineLength = lineOverhead + maxCodeLength
)

func FormatList[T any](title, ending string, items []T, format func(T) string) string {
	return FormatPage(title, fmt.Sprintf(ending, len(items)), items, 0, format)
}

func FormatPage[T any](title, footer string, items []T, offset int, format func(T) string) string {
	headingEndingLenghts := len(title) + 5 + len(footer)
	bufferSize := headingEndingLenghts + len(items)*maxLineLength

	sb := strings.Builder{}
	sb.Grow(bufferSize)

	sb.WriteString(title)
	sb.WriteString(": \n\n")

	for i, item := range items {
		sb.WriteString(fmt.Sprintf("%d. %s\n", offset+i+1, format(item)))
	}

	sb.WriteString("\n")
	sb.WriteString(footer)

	return sb.String()
}
