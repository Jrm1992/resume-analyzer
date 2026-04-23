package pdf

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	ledong "github.com/ledongthuc/pdf"
)

var ErrEmptyText = errors.New("pdf: no extractable text (scanned or encrypted?)")

// Parse extracts plain text from a PDF. Buffers the reader to satisfy
// ledongthuc/pdf's ReaderAt requirement.
func Parse(r io.Reader) (string, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("pdf: read: %w", err)
	}
	if len(buf) == 0 {
		return "", errors.New("pdf: empty input")
	}
	reader, err := ledong.NewReader(bytes.NewReader(buf), int64(len(buf)))
	if err != nil {
		return "", fmt.Errorf("pdf: open: %w", err)
	}
	var out strings.Builder
	totalPages := reader.NumPage()
	for i := 1; i <= totalPages; i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			return "", fmt.Errorf("pdf: page %d: %w", i, err)
		}
		out.WriteString(text)
		out.WriteString("\n")
	}
	result := strings.TrimSpace(out.String())
	if result == "" {
		return "", ErrEmptyText
	}
	return result, nil
}
