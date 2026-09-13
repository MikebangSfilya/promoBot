package report

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSuffix = "%s (page %d)"

func testBlock(header string, lines ...string) block {
	return block{header: header, lines: lines}
}

// testPages builds the options these tests share, with a cap high enough to stay
// out of the way except where a test sets its own.
func testPages(title string, limit int) pageOptions {
	return pageOptions{
		title:     title,
		suffix:    testSuffix,
		truncated: "... %d more page(s) left out",
		limit:     limit,
		maxPages:  50,
	}
}

// linesOf returns everything the pages hold apart from the titles, so a test can
// check that nothing was lost or duplicated by the split.
func linesOf(pages []string, header string) []string {
	var lines []string

	for _, page := range pages {
		for i, line := range strings.Split(page, "\n") {
			switch {
			case i == 0, line == "", line == header:
				continue
			}
			lines = append(lines, line)
		}
	}

	return lines
}

func TestPaginate(t *testing.T) {
	t.Run("everything on one page keeps the plain title", func(t *testing.T) {
		pages := paginate(testPages("Title", maxMessageLen), []block{
			testBlock("A", "a1", "a2"),
			testBlock("B", "b1"),
		})

		require.Len(t, pages, 1)
		assert.Equal(t, "Title\n\nA\na1\na2\n\nB\nb1", pages[0])
		assert.NotContains(t, pages[0], "(page")
	})

	t.Run("blocks that do not fit together are spread over pages", func(t *testing.T) {
		// Enough room for one block at a time, not two.
		limit := msgLen("Title (page 50)") + msgLen("\n\n") + 30

		pages := paginate(testPages("Title", limit), []block{
			testBlock("A", strings.Repeat("a", 20)),
			testBlock("B", strings.Repeat("b", 20)),
		})

		require.Len(t, pages, 2)
		assert.True(t, strings.HasPrefix(pages[0], "Title (page 1)\n\n"))
		assert.True(t, strings.HasPrefix(pages[1], "Title (page 2)\n\n"))
		assert.Contains(t, pages[0], "A\n"+strings.Repeat("a", 20))
		assert.Contains(t, pages[1], "B\n"+strings.Repeat("b", 20))
		assert.NotContains(t, pages[0], "B\n")
	})

	t.Run("a block bigger than a page repeats its header", func(t *testing.T) {
		limit := msgLen("Title (page 50)") + msgLen("\n\n") + 40

		var lines []string
		for i := range 12 {
			lines = append(lines, fmt.Sprintf("line-%02d", i))
		}
		pages := paginate(testPages("Title", limit), []block{testBlock("Author", lines...)})

		require.Greater(t, len(pages), 1)
		for i, page := range pages {
			assert.Contains(t, page, "Author", "page %d lost the header", i+1)
		}
		assert.Equal(t, lines, linesOf(pages, "Author"))
	})

	t.Run("every page stays within the limit", func(t *testing.T) {
		limit := msgLen("Title (page 50)") + msgLen("\n\n") + 50

		var blocks []block
		for i := range 10 {
			blocks = append(blocks, testBlock(
				fmt.Sprintf("H%02d", i),
				strings.Repeat("x", 25),
				strings.Repeat("y", 25),
			))
		}

		for _, page := range paginate(testPages("Title", limit), blocks) {
			assert.LessOrEqual(t, msgLen(page), limit)
		}
	})

	t.Run("never produces more pages than the cap", func(t *testing.T) {
		limit := msgLen("Title (page 50)") + msgLen("\n\n") + 40

		var blocks []block
		for i := range 30 {
			blocks = append(blocks, testBlock(fmt.Sprintf("H%02d", i), strings.Repeat("x", 30)))
		}

		opts := testPages("Title", limit)
		opts.maxPages = 3

		pages := paginate(opts, blocks)

		require.Len(t, pages, 3)
		for _, page := range pages {
			assert.LessOrEqual(t, msgLen(page), limit)
		}
		// The reader is told what was dropped rather than left guessing.
		assert.Contains(t, pages[2], "more page(s) left out")
	})

	t.Run("says how many pages were left out", func(t *testing.T) {
		limit := msgLen("Title (page 50)") + msgLen("\n\n") + 40

		var blocks []block
		for i := range 10 {
			blocks = append(blocks, testBlock(fmt.Sprintf("H%02d", i), strings.Repeat("x", 30)))
		}

		opts := testPages("Title", limit)
		opts.maxPages = 50
		uncapped := len(paginate(opts, blocks))
		require.Greater(t, uncapped, 2)

		opts.maxPages = 2
		pages := paginate(opts, blocks)

		require.Len(t, pages, 2)
		assert.Contains(t, pages[1], fmt.Sprintf("... %d more page(s) left out", uncapped-2))
	})

	// Saying "something was left out" must not itself overflow the message.
	t.Run("the notice fits within the limit", func(t *testing.T) {
		limit := msgLen("Title (page 50)") + msgLen("\n\n") + 40

		var blocks []block
		for i := range 20 {
			blocks = append(blocks, testBlock(fmt.Sprintf("H%02d", i), strings.Repeat("x", 35)))
		}

		opts := testPages("Title", limit)
		opts.maxPages = 1
		opts.truncated = strings.Repeat("!", 30) + " %d"

		pages := paginate(opts, blocks)

		require.Len(t, pages, 1)
		assert.LessOrEqual(t, msgLen(pages[0]), limit)
		assert.Contains(t, pages[0], "!!!")
	})

	// Telegram counts UTF-16 code units, so an emoji costs two and must never be
	// cut in half either.
	t.Run("emoji count as two units", func(t *testing.T) {
		assert.Equal(t, 2, msgLen("👤"))
		assert.Equal(t, 1, msgLen("я"))
		assert.Equal(t, 4, msgLen("👤👥"))

		limit := msgLen("👥 Title (page 50)") + msgLen("\n\n") + 30

		var blocks []block
		for i := range 6 {
			blocks = append(blocks, testBlock(fmt.Sprintf("👤 %d", i), strings.Repeat("👥", 8)))
		}

		pages := paginate(testPages("👥 Title", limit), blocks)

		for _, page := range pages {
			assert.LessOrEqual(t, msgLen(page), limit)
			// A halved emoji would show up as a replacement character.
			assert.NotContains(t, page, "�")
		}
	})

	t.Run("no content is lost across pages", func(t *testing.T) {
		limit := msgLen("Title (page 50)") + msgLen("\n\n") + 45

		var (
			blocks []block
			want   []string
		)
		for i := range 8 {
			line := fmt.Sprintf("entry number %02d", i)
			blocks = append(blocks, testBlock("H", line))
			want = append(want, line)
		}

		assert.Equal(t, want, linesOf(paginate(testPages("Title", limit), blocks), "H"))
	})
}
