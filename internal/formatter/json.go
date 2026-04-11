package formatter

import (
	"encoding/json"
	"io"

	"github.com/hidetzu/prism/pkg/prism"
)

// FormatJSON writes a prism.Result as indented JSON to w.
func FormatJSON(w io.Writer, result prism.Result) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(result)
}
