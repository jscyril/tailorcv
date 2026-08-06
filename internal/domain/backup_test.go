package domain

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestProfileBackupJSONRoundTripPreservesMetadata(t *testing.T) {
	backup := ProfileBackup{
		SchemaVersion: ProfileBackupSchemaVersion,
		ExportedAt:    time.Now().UTC().Format(time.RFC3339),
		Profile:       Profile{Name: "Ada Lovelace", Skills: []string{"Go"}, UpdatedAt: "2026-01-02T03:04:05Z"},
		Experiences: []Experience{{
			ID: "ccac4867-b2af-432a-a5c7-099eb9effd9f", Company: "Example", Title: "Engineer", StartDate: "2024-01", Position: 3, CreatedAt: "2024-01-01T00:00:00Z", UpdatedAt: "2025-01-01T00:00:00Z",
			Bullets: []EvidenceBullet{{ID: "a4e04b97-4646-44af-b4cd-ed925be390e5", Text: "Built a service", Provenance: ProvenanceManual, Verification: VerificationVerified, Position: 2, CreatedAt: "2024-01-01T00:00:00Z", UpdatedAt: "2025-01-01T00:00:00Z"}},
		}},
		Projects:   []Project{},
		Educations: []Education{},
		Jobs:       []Job{{ID: "f79dd67f-84c1-4452-9689-34ad2b282851", Company: "Example", Role: "Engineer", Description: "Build reliable services and deployment systems for a growing software platform.", CreatedAt: "2024-01-01T00:00:00Z", UpdatedAt: "2025-01-01T00:00:00Z"}},
	}
	data, err := json.Marshal(backup)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	decoded, err := DecodeProfileBackup(data)
	if err != nil {
		t.Fatalf("DecodeProfileBackup() error = %v", err)
	}
	if decoded.Profile.UpdatedAt != backup.Profile.UpdatedAt || decoded.Experiences[0].Position != 3 || decoded.Experiences[0].Bullets[0].Position != 2 || len(decoded.Jobs) != 1 {
		t.Fatalf("decoded backup lost metadata: %#v", decoded)
	}
}

func TestDecodeProfileBackupRejectsUnsupportedOrUnknownData(t *testing.T) {
	_, err := DecodeProfileBackup([]byte(`{"schemaVersion":99,"exportedAt":"2026-01-02T03:04:05Z","profile":{},"experiences":[],"projects":[],"educations":[]}`))
	if err == nil || !strings.Contains(err.Error(), "not supported") {
		t.Fatalf("unsupported schema error = %v", err)
	}
	_, err = DecodeProfileBackup([]byte(`{"schemaVersion":1,"exportedAt":"2026-01-02T03:04:05Z","profile":{},"experiences":[],"projects":[],"educations":[],"unexpected":true}`))
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("unknown field error = %v", err)
	}
}

func TestDecodeProfileBackupUpgradesVersionOne(t *testing.T) {
	decoded, err := DecodeProfileBackup([]byte(`{"schemaVersion":1,"exportedAt":"2026-01-02T03:04:05Z","profile":{},"experiences":[],"projects":[],"educations":[]}`))
	if err != nil {
		t.Fatalf("DecodeProfileBackup() error = %v", err)
	}
	if decoded.SchemaVersion != ProfileBackupSchemaVersion || decoded.Templates == nil || decoded.AIRuns == nil {
		t.Fatalf("upgraded backup = %#v", decoded)
	}
}
