package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"unicode"

	"github.com/jscyril/tailorcv/internal/domain"
)

type evidenceIndexExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func rebuildEvidenceSearch(ctx context.Context, executor evidenceIndexExecutor) error {
	for _, statement := range []string{
		`DELETE FROM career_evidence_fts`,
		`INSERT INTO career_evidence_fts(fact_id, source_id, source_type, source_label, text, skills)
		 SELECT b.id, e.id, 'experience', e.title || ' · ' || e.company, b.text, ''
		 FROM experience_bullets b JOIN experiences e ON e.id = b.experience_id`,
		`INSERT INTO career_evidence_fts(fact_id, source_id, source_type, source_label, text, skills)
		 SELECT b.id, p.id, 'project', p.name, b.text,
		        COALESCE((SELECT group_concat(name, ' ') FROM project_skills WHERE project_id = p.id), '') || ' ' ||
		        COALESCE((SELECT group_concat(name, ' ') FROM project_detected_languages WHERE project_id = p.id), '')
		 FROM project_bullets b JOIN projects p ON p.id = b.project_id`,
		`INSERT INTO career_evidence_fts(fact_id, source_id, source_type, source_label, text, skills)
		 SELECT p.id, p.id, 'project', p.name, CASE WHEN p.description = '' THEN p.name ELSE p.description END,
		        COALESCE((SELECT group_concat(name, ' ') FROM project_skills WHERE project_id = p.id), '') || ' ' ||
		        COALESCE((SELECT group_concat(name, ' ') FROM project_detected_languages WHERE project_id = p.id), '')
		 FROM projects p WHERE NOT EXISTS (SELECT 1 FROM project_bullets b WHERE b.project_id = p.id)`,
	} {
		if _, err := executor.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("rebuild career evidence search index: %w", err)
		}
	}
	return nil
}

func (s *Store) SearchEvidence(ctx context.Context, terms []string, limit int) ([]domain.EvidenceSearchHit, error) {
	query := evidenceSearchQuery(terms)
	if query == "" {
		return []domain.EvidenceSearchHit{}, nil
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT fact_id
		FROM career_evidence_fts
		WHERE career_evidence_fts MATCH ?
		ORDER BY bm25(career_evidence_fts)
		LIMIT ?
	`, query, limit)
	if err != nil {
		return nil, fmt.Errorf("search career evidence: %w", err)
	}
	defer rows.Close()
	hits := make([]domain.EvidenceSearchHit, 0)
	for rows.Next() {
		var factID string
		if err := rows.Scan(&factID); err != nil {
			return nil, fmt.Errorf("scan career evidence search result: %w", err)
		}
		score := 18 - len(hits)/2
		if score < 6 {
			score = 6
		}
		hits = append(hits, domain.EvidenceSearchHit{FactID: factID, Score: score})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate career evidence search results: %w", err)
	}
	return hits, nil
}

func evidenceSearchQuery(terms []string) string {
	phrases := make([]string, 0, len(terms))
	seen := make(map[string]struct{}, len(terms))
	for _, term := range terms {
		words := strings.FieldsFunc(strings.ToLower(term), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
		if len(words) == 0 {
			continue
		}
		phrase := strings.Join(words, " ")
		if _, exists := seen[phrase]; exists {
			continue
		}
		seen[phrase] = struct{}{}
		phrases = append(phrases, `"`+phrase+`"`)
		if len(phrases) == 20 {
			break
		}
	}
	return strings.Join(phrases, " OR ")
}
