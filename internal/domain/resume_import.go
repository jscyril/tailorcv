package domain

// ResumeImport is what TailorCV read from an existing resume file, ready for
// the user to review before it is saved. Every entry already passed the same
// validation as the profile forms.
type ResumeImport struct {
	FileName    string            `json:"fileName"`
	Profile     ProfileInput      `json:"profile"`
	Experiences []ExperienceInput `json:"experiences"`
	Projects    []ProjectInput    `json:"projects"`
	Educations  []EducationInput  `json:"educations"`
	Warnings    []string          `json:"warnings"`
	// PDFBase64 is set only when previewing a PDF, for the UI to extract text.
	PDFBase64 string `json:"pdfBase64,omitempty"`
}

// ResumeImportResult counts what an import added and skipped.
type ResumeImportResult struct {
	Experiences int `json:"experiences"`
	Projects    int `json:"projects"`
	Educations  int `json:"educations"`
	Skills      int `json:"skills"`
	Duplicates  int `json:"duplicates"`
}
