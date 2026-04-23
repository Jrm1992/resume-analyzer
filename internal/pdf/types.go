package pdf

type ContactInfo struct {
	Email    string
	Phone    string
	Location string
	LinkedIn string
}

type ExperienceEntry struct {
	Company string
	Role    string
	Dates   string
	Bullets []string
}

type EducationEntry struct {
	Institution string
	Degree      string
	Dates       string
}

type RewrittenResume struct {
	Name       string
	Contact    ContactInfo
	Summary    string
	Skills     []string
	Experience []ExperienceEntry
	Education  []EducationEntry
}
