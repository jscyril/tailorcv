package domain

import "testing"

func TestProfileInputValidateNormalizesSkills(t *testing.T) {
	input := ProfileInput{
		Name:   "  Ada Lovelace ",
		Email:  "ada@example.com",
		Skills: []string{" Go ", "go", "TypeScript", ""},
	}

	profile, err := input.Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if profile.Name != "Ada Lovelace" {
		t.Fatalf("Name = %q", profile.Name)
	}
	if len(profile.Skills) != 2 || profile.Skills[0] != "Go" || profile.Skills[1] != "TypeScript" {
		t.Fatalf("Skills = %#v", profile.Skills)
	}
}

func TestProfileInputValidateRejectsInvalidURLs(t *testing.T) {
	_, err := (ProfileInput{Website: "javascript:alert(1)"}).Validate()
	if err == nil {
		t.Fatal("Validate() expected URL error")
	}
}
