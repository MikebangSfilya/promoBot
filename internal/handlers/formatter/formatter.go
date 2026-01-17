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

func FormatList[T fmt.Stringer](title, ending string, items []T) string {

	headingEndingLenghts := len(title) + 5 + len(ending)
	bufferSize := headingEndingLenghts + len(items)*maxLineLength

	sb := strings.Builder{}
	sb.Grow(bufferSize)

	sb.WriteString(title)
	sb.WriteString(": \n\n")

	for i, item := range items {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, item.String()))
	}

	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf(ending, len(items)))

	return sb.String()
}
