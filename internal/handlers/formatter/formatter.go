package formatter

import (
	"fmt"
	"strings"
)

func FormatList[T fmt.Stringer](title, ending string, items []T) string {
	sb := strings.Builder{}
	sb.WriteString(title)
	sb.WriteString(": \n\n")

	for i, item := range items {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, item.String()))
	}

	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf(ending, len(items)))

	return sb.String()
}
