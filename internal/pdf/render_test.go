package pdf

import (
	"bytes"
	"strings"
	"testing"
)

func TestRender_ProducesPDFBytes(t *testing.T) {
	resume := RewrittenResume{
		Name: "Jane Doe",
		Contact: ContactInfo{
			Email: "jane@example.com",
			Phone: "555-0100",
		},
		Summary: "Senior backend engineer with 8 years of Go experience.",
		Skills:  []string{"Go", "PostgreSQL", "gRPC"},
		Experience: []ExperienceEntry{
			{
				Company: "Acme",
				Role:    "Senior Engineer",
				Dates:   "2020–present",
				Bullets: []string{"Built billing system handling $50M/yr."},
			},
		},
		Education: []EducationEntry{
			{Institution: "MIT", Degree: "BS Computer Science", Dates: "2012–2016"},
		},
	}
	data, err := Render(resume)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty output")
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Errorf("output does not start with %%PDF-, got %q", data[:8])
	}
	// Roundtrip: parse back and assert name appears.
	text, err := Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Parse roundtrip: %v", err)
	}
	if !strings.Contains(text, "Jane Doe") {
		t.Errorf("roundtrip missing name, got %q", text)
	}
}

func TestRender_EmptyResumeSucceeds(t *testing.T) {
	// A resume with only a name should still render.
	data, err := Render(RewrittenResume{Name: "Anonymous"})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Error("not a PDF")
	}
}
