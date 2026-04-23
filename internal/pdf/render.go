package pdf

import (
	"fmt"

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
func Render(r RewrittenResume) ([]byte, error) {
	cfg := config.NewBuilder().
		WithPageSize(pagesize.A4).
		WithLeftMargin(15).
		WithTopMargin(15).
		WithRightMargin(15).
		WithBottomMargin(15).
		Build()
	m := maroto.New(cfg)

	m.AddRow(12, text.NewCol(12, r.Name, props.Text{
		Size:  18,
		Style: fontstyle.Bold,
		Align: align.Left,
	}))
	if c := formatContact(r.Contact); c != "" {
		m.AddRow(6, text.NewCol(12, c, props.Text{Size: 10, Align: align.Left}))
	}
	if r.Summary != "" {
		addSection(m, "SUMMARY")
		m.AddRow(10, text.NewCol(12, r.Summary, props.Text{Size: 10}))
	}
	if len(r.Skills) > 0 {
		addSection(m, "SKILLS")
		m.AddRow(8, text.NewCol(12, joinSkills(r.Skills), props.Text{Size: 10}))
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
			m.AddRow(8, text.NewCol(12, header, props.Text{Size: 11, Style: fontstyle.Bold}))
			for _, b := range e.Bullets {
				m.AddRow(6, text.NewCol(12, "• "+b, props.Text{Size: 10}))
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
			m.AddRow(6, text.NewCol(12, line, props.Text{Size: 10}))
		}
	}

	doc, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("pdf: render: %w", err)
	}
	return doc.GetBytes(), nil
}

func addSection(m core.Maroto, title string) {
	m.AddRow(2)
	m.AddRow(8, text.NewCol(12, title, props.Text{
		Size:  11,
		Style: fontstyle.Bold,
		Align: align.Left,
	}))
}

func formatContact(c ContactInfo) string {
	parts := []string{}
	for _, p := range []string{c.Email, c.Phone, c.Location, c.LinkedIn} {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return joinSep(parts, " • ")
}

func joinSkills(s []string) string  { return joinSep(s, ", ") }

func joinSep(parts []string, sep string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += sep
		}
		out += p
	}
	return out
}
