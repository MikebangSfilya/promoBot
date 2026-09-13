package report

import (
	"fmt"
	"strings"
	"unicode/utf16"
)

// maxMessageLen is Telegram's limit for a single text message.
const maxMessageLen = 4096

const (
	// pageSeparator goes between the title and the blocks, and between blocks.
	pageSeparator = "\n\n"
	// pageSeparatorLen is its length in UTF-16 code units. The separator is
	// ASCII, so its byte count already is that, and the compiler folds it.
	pageSeparatorLen = len(pageSeparator)
)

// pageOptions is everything paginate needs besides the content itself.
type pageOptions struct {
	// title heads every page.
	title string
	// suffix turns the title into "Title (page 2)" when there is more than one.
	suffix string
	// truncated announces the pages left out once maxPages is reached.
	truncated string
	// limit is the largest message the chat accepts, in UTF-16 code units.
	limit int
	// maxPages bounds how many messages one report may become.
	maxPages int
}

// block is one self-contained piece of a report: an author and their events, or
// a section and its entries.
type block struct {
	// header is repeated at the top of every page the block continues on, so a
	// block split across messages never loses the name it belongs to.
	header string
	lines  []string
}

// paginate lays the blocks out into pages that each fit within the limit.
//
// A block is kept whole whenever it fits; one that is too long for an entire
// page is split across pages, repeating its header. Every page starts with the
// title, which carries the "(page N)" suffix only when there is more than one
// page.
func paginate(opts pageOptions, blocks []block) []string {
	// The suffix is not known to be needed yet, but it takes space when it is.
	// maxPages is the highest number that can appear, so measuring the header
	// against it reserves exactly enough, and never too little.
	budget := opts.limit - msgLen(fmt.Sprintf(opts.suffix, opts.title, opts.maxPages)) - pageSeparatorLen

	var (
		pages   []string
		current []string
	)

	flush := func() {
		if len(current) > 0 {
			pages = append(pages, strings.Join(current, pageSeparator))
			current = nil
		}
	}

	used := func() int {
		return msgLen(strings.Join(current, pageSeparator))
	}

	for _, b := range blocks {
		for _, part := range splitBlock(b, budget) {
			separator := 0
			if len(current) > 0 {
				separator = pageSeparatorLen
			}

			if used()+separator+msgLen(part) > budget {
				flush()
			}
			current = append(current, part)
		}
	}
	flush()

	if len(pages) == 0 {
		pages = []string{""}
	}

	// Past the cap the rest is dropped, but never silently: the last page says
	// how much was left out.
	if len(pages) > opts.maxPages {
		dropped := len(pages) - opts.maxPages
		pages = pages[:opts.maxPages]

		last := len(pages) - 1
		pages[last] = withNotice(pages[last], fmt.Sprintf(opts.truncated, dropped), budget)
	}

	return withTitles(pages, opts.title, opts.suffix)
}

// splitBlock renders a block, breaking it into several parts along its lines
// when it cannot fit on a page by itself. Each part repeats the header.
func splitBlock(b block, budget int) []string {
	whole := renderBlock(b.header, b.lines)
	if msgLen(whole) <= budget {
		return []string{whole}
	}

	var (
		parts   []string
		current []string
	)

	for _, line := range b.lines {
		candidate := renderBlock(b.header, append(current, line))
		if len(current) > 0 && msgLen(candidate) > budget {
			parts = append(parts, renderBlock(b.header, current))
			current = nil
		}
		current = append(current, line)
	}

	if len(current) > 0 {
		parts = append(parts, renderBlock(b.header, current))
	}

	return parts
}

// withNotice appends the notice to a page, dropping trailing lines until it
// fits, so that saying "something was left out" cannot itself push the message
// over the limit.
func withNotice(page, notice string, budget int) string {
	lines := strings.Split(page, "\n")

	for len(lines) > 1 && msgLen(strings.Join(lines, "\n"))+pageSeparatorLen+msgLen(notice) > budget {
		lines = lines[:len(lines)-1]
	}

	return strings.Join(lines, "\n") + pageSeparator + notice
}

func renderBlock(header string, lines []string) string {
	if len(lines) == 0 {
		return header
	}
	return header + "\n" + strings.Join(lines, "\n")
}

// withTitles puts the heading on every page, adding the "(page N)" suffix only
// when there is more than one.
func withTitles(pages []string, title, suffixFormat string) []string {
	titled := make([]string, 0, len(pages))

	for i, page := range pages {
		heading := title
		if len(pages) > 1 {
			heading = fmt.Sprintf(suffixFormat, title, i+1)
		}
		titled = append(titled, heading+pageSeparator+page)
	}

	return titled
}

// msgLen counts the message the way Telegram does: in UTF-16 code units, so an
// emoji costs two and a Cyrillic letter one. Neither len() nor the rune count
// agrees with that.
func msgLen(s string) int {
	n := 0
	for _, r := range s {
		if size := utf16.RuneLen(r); size > 0 {
			n += size
			continue
		}
		n++ // An unpaired surrogate still occupies one unit.
	}
	return n
}
