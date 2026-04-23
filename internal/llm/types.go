package llm

import "github.com/jose/resume-analyzer/internal/pdf"

type CategoryBreakdown struct {
	Skills     int `json:"skills"`
	Experience int `json:"experience"`
	Education  int `json:"education"`
}

type Suggestion struct {
	Section   string `json:"section"`
	Original  string `json:"original"`
	Suggested string `json:"suggested"`
	Rationale string `json:"rationale"`
}

type KeywordMatch struct {
	Name    string `json:"name"`
	Present bool   `json:"present"`
}

type AnalysisResult struct {
	Score           int                 `json:"score"`
	Breakdown       CategoryBreakdown   `json:"breakdown"`
	JobKeywords     []KeywordMatch      `json:"job_keywords"`
	TitleAlignment  string              `json:"title_alignment"`
	Missing         []string            `json:"missing"`
	Strengths       []string            `json:"strengths"`
	Suggestions     []Suggestion        `json:"suggestions"`
	Rewritten       pdf.RewrittenResume `json:"rewritten"`
}
