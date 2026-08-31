package formatter

import (
	"strings"
	"testing"

	"github.com/MikebangSfilya/promoBot/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestFormatPage(t *testing.T) {
	items := []model.ResponseCode{
		{Code: "ONE", BonusLength: 1, Capacity: 2},
		{Code: "TWO", BonusLength: 3, Capacity: 4},
	}

	page := FormatPage("Promo codes", "Page 2/3 · Total: 42", items, 20, func(item model.ResponseCode) string {
		return item.Code
	})

	assert.True(t, strings.HasPrefix(page, "Promo codes: \n\n21. ONE"))
	assert.Contains(t, page, "\n22. TWO")
	assert.True(t, strings.HasSuffix(page, "Page 2/3 · Total: 42"))
}
