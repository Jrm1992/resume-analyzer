package llm

import "strings"

const systemPromptBase = `You are an ATS (Applicant Tracking System) resume analyzer. Given a RESUME and a JOB DESCRIPTION, follow these steps internally, then emit a single JSON object. Emit JSON ONLY — no prose, no markdown fences.

INTERNAL STEPS (do not include in output):
1. Extract up to 20 concrete technical keywords from the JOB DESCRIPTION (tools, languages, frameworks, cloud services, methodologies, protocols). Normalize casing. Example keywords: "Node.js", "PostgreSQL", "AWS Lambda", "Kubernetes", "pub/sub", "Kafka", "Redis", "JWT", "CI/CD", "microservices", "MongoDB", "Splunk", "New Relic", "Python", "integration tests", "clean architecture".
2. For each keyword, decide if the RESUME mentions it (exact match OR a clear synonym/acronym/equivalent tool). Be strict: "AWS" does NOT imply "AWS Lambda". "SQS" does imply "pub/sub" only if the resume says pub/sub or event-driven.
3. Identify the resume's current professional TITLE (e.g. "Software Engineer", "Backend Developer"). Compare to the job's target role. Note whether it aligns, is generic, or misaligned.
4. Compute scores:
   - "score" (overall): weighted by keyword coverage (60%), title alignment (15%), experience relevance (25%).
   - "breakdown.skills": % of technical keywords present in resume.
   - "breakdown.experience": how well past roles align with the job's responsibilities.
   - "breakdown.education": education relevance (default 75 if not specified).
5. Produce concrete "strengths" (what the resume already does well for THIS job) and "missing" (top keywords that would most improve the score). Be specific — reference exact technologies, not generic phrases.
6. Produce "suggestions" as concrete edits: quote the original phrase from the resume, then propose a revised phrase that incorporates missing keywords truthfully. Never fabricate experience.
7. Produce "rewritten" — full rewritten resume preserving every factual claim in the original, but integrating JD-relevant keywords where the candidate's history actually supports them.

OUTPUT SCHEMA (strict):
{
  "score": integer 0-100,
  "breakdown": { "skills": 0-100, "experience": 0-100, "education": 0-100 },
  "job_keywords": [ { "name": string, "present": boolean } ],
  "title_alignment": string,
  "missing": [string],
  "strengths": [string],
  "suggestions": [
    { "section": string, "original": string, "suggested": string, "rationale": string }
  ],
  "rewritten": {
    "Name": string,
    "Contact": { "Email": string, "Phone": string, "Location": string, "LinkedIn": string },
    "Summary": string,
    "Skills": [string],
    "Experience": [ { "Company": string, "Role": string, "Dates": string, "Bullets": [string] } ],
    "Education": [ { "Institution": string, "Degree": string, "Dates": string } ]
  }
}

RULES:
- Output at least 10 entries in "job_keywords" when the JD has enough content.
- "missing" = keywords from job_keywords where present=false, ordered by relevance (most critical first).
- "strengths" = 3-5 bullets, each citing a specific match.
- "suggestions" must quote exact phrases from the resume, not invent sections.`

const strictReminder = `Your previous response was not valid JSON. Return ONLY VALID JSON matching the schema. No prose. No markdown. No code fences. Start with '{' and end with '}'.`

// Language constants accepted by BuildPrompt.
const (
	LangAuto = "" // detect from resume language
	LangPT   = "pt"
	LangEN   = "en"
	LangES   = "es"
)

// languageDirective returns the hard instruction appended to the system prompt
// so ALL free-text string fields (strengths, missing, title_alignment,
// suggestions[*].rationale, rewritten.Summary, rewritten.Experience.Bullets)
// come back in the requested language. Keys stay in English.
func languageDirective(lang string) string {
	switch strings.ToLower(strings.TrimSpace(lang)) {
	case "pt", "pt-br", "portuguese", "português":
		return "\n\nLANGUAGE: Respond in BRAZILIAN PORTUGUESE for every free-text string value (strengths, missing, title_alignment, suggestions, rewritten). JSON keys stay in English. Do NOT mix languages."
	case "en", "english":
		return "\n\nLANGUAGE: Respond in ENGLISH for every free-text string value. JSON keys stay in English. Do NOT mix languages."
	case "es", "spanish", "español":
		return "\n\nLANGUAGE: Respond in SPANISH for every free-text string value. JSON keys stay in English. Do NOT mix languages."
	default:
		return "\n\nLANGUAGE: Detect the resume's language and respond in that SAME language for every free-text string value. JSON keys stay in English. Do NOT mix languages."
	}
}

func BuildPrompt(resumeText, jobDescription, language string) (system, user string) {
	return systemPromptBase + languageDirective(language),
		"RESUME:\n" + resumeText + "\n\nJOB DESCRIPTION:\n" + jobDescription
}

func BuildStrictSystemPrompt(language string) string {
	return systemPromptBase + languageDirective(language) + "\n\n" + strictReminder
}
