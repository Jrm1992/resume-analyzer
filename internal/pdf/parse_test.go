package pdf

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

// makeSimplePDF writes a minimal single-page PDF containing the given text.
// Hand-rolled because we don't want a render dependency in parse tests.
func makeSimplePDF(t *testing.T, text string) []byte {
	t.Helper()
	// Minimal 1-page PDF with a single Tj text operator.
	// Uses the 14 standard Type 1 fonts (Helvetica) so no font embed is needed.
	content := "BT /F1 12 Tf 72 720 Td (" + text + ") Tj ET"
	stream := []byte(content)
	pdf := &bytes.Buffer{}
	pdf.WriteString("%PDF-1.4\n")
	// object 1: catalog
	obj1 := "1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n"
	// object 2: pages
	obj2 := "2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n"
	// object 3: page
	obj3 := "3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] " +
		"/Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>\nendobj\n"
	// object 4: content stream
	obj4 := "4 0 obj\n<< /Length " + itoa(len(stream)) + " >>\nstream\n" +
		string(stream) + "\nendstream\nendobj\n"
	// object 5: font
	obj5 := "5 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n"

	offsets := []int{0}
	write := func(s string) {
		offsets = append(offsets, pdf.Len())
		pdf.WriteString(s)
	}
	write(obj1)
	write(obj2)
	write(obj3)
	write(obj4)
	write(obj5)

	xrefStart := pdf.Len()
	pdf.WriteString("xref\n0 6\n")
	pdf.WriteString("0000000000 65535 f \n")
	for _, off := range offsets[1:] {
		pdf.WriteString(pad10(off) + " 00000 n \n")
	}
	pdf.WriteString("trailer\n<< /Size 6 /Root 1 0 R >>\nstartxref\n")
	pdf.WriteString(itoa(xrefStart))
	pdf.WriteString("\n%%EOF\n")
	return pdf.Bytes()
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

func pad10(n int) string {
	s := itoa(n)
	for len(s) < 10 {
		s = "0" + s
	}
	return s
}

func TestParse_ExtractsText(t *testing.T) {
	data := makeSimplePDF(t, "Jane Doe Software Engineer")
	text, err := Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !strings.Contains(text, "Jane Doe") {
		t.Errorf("got %q, want substring 'Jane Doe'", text)
	}
}

func TestParse_EmptyReader(t *testing.T) {
	_, err := Parse(bytes.NewReader(nil))
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestParse_BadPDF(t *testing.T) {
	_, err := Parse(bytes.NewReader([]byte("not a pdf")))
	if err == nil {
		t.Fatal("expected error for non-PDF input")
	}
}

// ReaderAt assertion — ledongthuc/pdf requires ReaderAt + size.
var _ io.Reader = (*bytes.Reader)(nil)
