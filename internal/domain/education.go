package domain

import (
	"fmt"
	"strings"
)

const maxEducationDetailsLength = 1200

type Education struct {
	ID           string `json:"id"`
	Institution  string `json:"institution"`
	Degree       string `json:"degree"`
	FieldOfStudy string `json:"fieldOfStudy"`
	Location     string `json:"location"`
	StartDate    string `json:"startDate"`
	EndDate      string `json:"endDate"`
	Current      bool   `json:"current"`
	Details      string `json:"details"`
	Position     int    `json:"position"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

type EducationInput struct {
	ID           string `json:"id"`
	Institution  string `json:"institution"`
	Degree       string `json:"degree"`
	FieldOfStudy string `json:"fieldOfStudy"`
	Location     string `json:"location"`
	StartDate    string `json:"startDate"`
	EndDate      string `json:"endDate"`
	Current      bool   `json:"current"`
	Details      string `json:"details"`
}

func (input EducationInput) Validate() (Education, error) {
	education := Education{
		ID:           strings.TrimSpace(input.ID),
		Institution:  strings.Join(strings.Fields(input.Institution), " "),
		Degree:       strings.Join(strings.Fields(input.Degree), " "),
		FieldOfStudy: strings.Join(strings.Fields(input.FieldOfStudy), " "),
		Location:     strings.Join(strings.Fields(input.Location), " "),
		StartDate:    strings.TrimSpace(input.StartDate),
		EndDate:      strings.TrimSpace(input.EndDate),
		Current:      input.Current,
		Details:      strings.TrimSpace(input.Details),
	}

	if education.Institution == "" {
		return Education{}, fmt.Errorf("institution is required")
	}
	if len(education.Institution) > 180 {
		return Education{}, fmt.Errorf("institution must be 180 characters or fewer")
	}
	if education.Degree == "" {
		return Education{}, fmt.Errorf("degree is required")
	}
	if len(education.Degree) > 180 {
		return Education{}, fmt.Errorf("degree must be 180 characters or fewer")
	}
	if len(education.FieldOfStudy) > 180 {
		return Education{}, fmt.Errorf("field of study must be 180 characters or fewer")
	}
	if len(education.Location) > 180 {
		return Education{}, fmt.Errorf("location must be 180 characters or fewer")
	}
	if err := validateMonth("start date", education.StartDate, false); err != nil {
		return Education{}, err
	}
	if education.Current {
		education.EndDate = ""
	} else if err := validateMonth("end date", education.EndDate, false); err != nil {
		return Education{}, err
	}
	if education.StartDate != "" && education.EndDate != "" && education.EndDate < education.StartDate {
		return Education{}, fmt.Errorf("end date cannot be before start date")
	}
	if len(education.Details) > maxEducationDetailsLength {
		return Education{}, fmt.Errorf("education details must be %d characters or fewer", maxEducationDetailsLength)
	}
	return education, nil
}
