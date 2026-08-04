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
