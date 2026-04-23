package pdf

import (
	"fmt"
	"strings"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/pagesize"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

// Render produces an ATS-friendly single-column PDF from a RewrittenResume.
// Uses AddAutoRow for text that may wrap so row heights scale with content
// and lines don't overlap.
func Render(r RewrittenResume) ([]byte, error) {
	cfg := config.NewBuilder().
		WithPageSize(pagesize.A4).
		WithLeftMargin(15).
		WithTopMargin(15).
		WithRightMargin(15).
		WithBottomMargin(15).
		Build()
	m := maroto.New(cfg)

	m.AddAutoRow(text.NewCol(12, r.Name, props.Text{
		Size:  18,
		Style: fontstyle.Bold,
		Align: align.Left,
		Top:   1,
	}))
	if c := formatContact(r.Contact); c != "" {
		m.AddAutoRow(text.NewCol(12, c, props.Text{
			Size:  10,
			Align: align.Left,
			Top:   1,
		}))
	}
	if r.Summary != "" {
		addSection(m, "SUMMARY")
		m.AddAutoRow(text.NewCol(12, r.Summary, props.Text{Size: 10, Top: 1}))
	}
	if len(r.Skills) > 0 {
		addSection(m, "SKILLS")
		m.AddAutoRow(text.NewCol(12, joinSkills(r.Skills), props.Text{Size: 10, Top: 1}))
	}
	if len(r.Experience) > 0 {
		addSection(m, "EXPERIENCE")
		for _, e := range r.Experience {
			header := e.Role
			if e.Company != "" {
				header = e.Role + " — " + e.Company
			}
			if e.Dates != "" {
				header += " (" + e.Dates + ")"
			}
			m.AddAutoRow(text.NewCol(12, header, props.Text{
				Size:  11,
				Style: fontstyle.Bold,
				Top:   2,
			}))
			for _, b := range e.Bullets {
				m.AddAutoRow(text.NewCol(12, "• "+b, props.Text{
					Size: 10,
					Top:  1,
				}))
			}
		}
	}
	if len(r.Education) > 0 {
		addSection(m, "EDUCATION")
		for _, e := range r.Education {
			line := e.Degree
			if e.Institution != "" {
				line += ", " + e.Institution
			}
			if e.Dates != "" {
				line += " (" + e.Dates + ")"
			}
			m.AddAutoRow(text.NewCol(12, line, props.Text{Size: 10, Top: 1}))
		}
	}

	doc, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("pdf: render: %w", err)
	}
	return doc.GetBytes(), nil
}

func addSection(m core.Maroto, title string) {
	m.AddRow(3) // spacer
	m.AddAutoRow(text.NewCol(12, title, props.Text{
		Size:  11,
		Style: fontstyle.Bold,
		Align: align.Left,
		Top:   1,
	}))
}

func formatContact(c ContactInfo) string {
	all := [...]string{c.Email, c.Phone, c.Location, c.LinkedIn}
	parts := make([]string, 0, len(all))
	for _, p := range all {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return strings.Join(parts, " • ")
}

func joinSkills(s []string) string { return strings.Join(s, ", ") }
