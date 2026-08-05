package domain

import "testing"

func TestCreateResumeVersionInputRequiresUniqueSelectedFacts(t *testing.T) {
	_, err := (CreateResumeVersionInput{JobID: "job", TemplateID: "template", SelectedFactIDs: []string{"fact", "fact"}}).Validate()
	if err == nil {
		t.Fatal("Validate() expected duplicate selected evidence error")
	}
	validated, err := (CreateResumeVersionInput{JobID: " job ", TemplateID: " template ", SelectedFactIDs: []string{" fact "}}).Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if validated.JobID != "job" || validated.TemplateID != "template" || validated.SelectedFactIDs[0] != "fact" {
		t.Fatalf("validated input = %#v", validated)
	}
}

func TestSaveResumeVersionEditInputAndContentHash(t *testing.T) {
	validated, err := (SaveResumeVersionEditInput{ApplicationID: " app ", BaseVersionID: " version ", LatexSource: "source"}).Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if validated.ApplicationID != "app" || validated.BaseVersionID != "version" {
		t.Fatalf("validated input = %#v", validated)
	}
	if ResumeContentHash("source") != "41cf6794ba4200b839c53531555f0f3998df4cbb01a4d5cb0b94e3ca5e23947d" {
		t.Fatalf("ResumeContentHash() = %q", ResumeContentHash("source"))
	}
	if _, err := (SaveResumeVersionEditInput{ApplicationID: "app", BaseVersionID: "version"}).Validate(); err == nil {
		t.Fatal("Validate() expected empty source error")
	}
}
