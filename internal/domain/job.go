package domain

import (
	"fmt"
	"strings"
)

const maxJobDescriptionLength = 100_000

// Job is a locally saved opportunity. Analysis is intentionally computed from
// the latest career profile rather than stored as an opaque score.
type Job struct {
	ID          string `json:"id"`
	Company     string `json:"company"`
	Role        string `json:"role"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type JobInput struct {
	ID          string `json:"id"`
	Company     string `json:"company"`
	Role        string `json:"role"`
	Description string `json:"description"`
}

func (input JobInput) Validate() (Job, error) {
	job := Job{
		ID:          strings.TrimSpace(input.ID),
		Company:     strings.Join(strings.Fields(input.Company), " "),
		Role:        strings.Join(strings.Fields(input.Role), " "),
		Description: strings.TrimSpace(input.Description),
	}
	if len(job.Company) > 180 {
		return Job{}, fmt.Errorf("company must be 180 characters or fewer")
	}
	if len(job.Role) > 180 {
		return Job{}, fmt.Errorf("role must be 180 characters or fewer")
	}
	if len(job.Description) < 40 {
		return Job{}, fmt.Errorf("job description must contain at least 40 characters")
	}
	if len(job.Description) > maxJobDescriptionLength {
		return Job{}, fmt.Errorf("job description must be %d characters or fewer", maxJobDescriptionLength)
	}
	return job, nil
}
