package report

import "strings"

// escapeHTML makes a value safe to drop into the reports' HTML markup.
//
// Telegram's HTML parse mode only ascribes meaning to these three characters,
// so escaping exactly them keeps everything else — dates, em dashes, arrows —
// intact. A promo code or an author name holding a stray "&" would otherwise
// break the markup of the whole message, not just its own line.
var escapeHTML = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
).Replace
