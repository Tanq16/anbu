package vault

import (
	"fmt"
	"time"
	"uuid"
)

const (
	ExportFormat  = "anbu-vault-export"
	ExportVersion = 1
)

type Export struct {
	Format     string    `json:"format"`
	Version    int       `json:"version"`
	ExportedAt time.Time `json:"exported_at"`
	Secrets    []Secret  `json:"secrets"`
}

func (s *Store) Export() Export {
	return Export{
		Format:     ExportFormat,
		Version:    ExportVersion,
		ExportedAt: time.Now().UTC(),
		Secrets:    s.List(),
	}
}

func (s *Store) Import(doc Export) (added, skipped []string, err error) {
	if doc.Format != ExportFormat || doc.Version != ExportVersion {
		return nil, nil, invalid("unsupported export format %q version %d", doc.Format, doc.Version)
	}
	incoming := make([]Secret, len(doc.Secrets))
	for i, sec := range doc.Secrets {
		if err := sec.normalize(); err != nil {
			return nil, nil, fmt.Errorf("%w (secret %q)", err, sec.Name)
		}
		if sec.ID == "" {
			sec.ID = uuid.NewV7().String()
		} else if _, err := uuid.Parse(sec.ID); err != nil {
			return nil, nil, invalid("secret %s has an invalid id %q", sec.Name, sec.ID)
		}
		incoming[i] = sec
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	added, skipped = []string{}, []string{}
	var inserted []string
	for _, sec := range incoming {
		if _, exists := s.secrets[sec.ID]; exists || s.nameTaken(sec.Name, "") {
			skipped = append(skipped, sec.Name)
			continue
		}
		if sec.CreatedAt.IsZero() {
			sec.CreatedAt = now
		}
		if sec.UpdatedAt.IsZero() {
			sec.UpdatedAt = now
		}
		s.secrets[sec.ID] = sec
		inserted = append(inserted, sec.ID)
		added = append(added, sec.Name)
	}
	if len(inserted) == 0 {
		return added, skipped, nil
	}
	if err := s.persist(); err != nil {
		for _, id := range inserted {
			delete(s.secrets, id)
		}
		return nil, nil, err
	}
	return added, skipped, nil
}
