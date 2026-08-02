package domain

import "testing"

func TestEducationInputValidateNormalizesFields(t *testing.T) {
	education, err := (EducationInput{
		Institution:  "  Example   Institute ",
		Degree:       " Bachelor of Science ",
		FieldOfStudy: " Computer   Science ",
		StartDate:    "2021-08",
		EndDate:      "2025-05",
	}).Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if education.Institution != "Example Institute" || education.FieldOfStudy != "Computer Science" {
		t.Fatalf("Education = %#v", education)
	}
}

func TestEducationInputValidateCurrentStudyClearsEndDate(t *testing.T) {
	education, err := (EducationInput{
		Institution: "Example Institute",
		Degree:      "Master of Science",
		StartDate:   "2024-08",
		EndDate:     "2026-05",
		Current:     true,
	}).Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if education.EndDate != "" {
		t.Fatalf("EndDate = %q, want empty", education.EndDate)
	}
}

func TestEducationInputValidateRejectsInvalidValues(t *testing.T) {
	if _, err := (EducationInput{Degree: "Bachelor of Science"}).Validate(); err == nil {
		t.Fatal("Validate() expected institution error")
	}
	if _, err := (EducationInput{Institution: "Example", Degree: "Bachelor of Science", StartDate: "2025-08", EndDate: "2024-05"}).Validate(); err == nil {
		t.Fatal("Validate() expected date order error")
	}
}
