package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type GuidedSessionState struct {
	SchemaVersion   int                            `json:"schema_version"`
	LastActiveChain string                         `json:"last_active_chain,omitempty"`
	LastUpdatedAt   string                         `json:"last_updated_at,omitempty"`
	Sessions        map[string]GuidedSessionRecord `json:"sessions"`
}

type GuidedSessionRecord struct {
	ChainID             string            `json:"chain_id"`
	Brainstorm          string            `json:"brainstorm,omitempty"`
	Epic                string            `json:"epic,omitempty"`
	Spec                string            `json:"spec,omitempty"`
	CurrentStage        string            `json:"current_stage,omitempty"`
	CurrentCluster      int               `json:"current_cluster,omitempty"`
	CurrentClusterLabel string            `json:"current_cluster_label,omitempty"`
	StageStatuses       map[string]string `json:"stage_statuses,omitempty"`
	Summary             string            `json:"summary,omitempty"`
	NextAction          string            `json:"next_action,omitempty"`
	CreatedAt           string            `json:"created_at,omitempty"`
	UpdatedAt           string            `json:"updated_at,omitempty"`
}

func (m *Manager) ReadGuidedSessionState() (*GuidedSessionState, error) {
	info, err := m.Resolve()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(info.SessionsFile); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		now := time.Now().UTC().Format(time.RFC3339)
		if _, err := ensureGuidedSessionState(info.SessionsFile, now); err != nil {
			return nil, err
		}
	}
	return readGuidedSessionStateFile(info.SessionsFile)
}

func (m *Manager) WriteGuidedSessionState(state GuidedSessionState) error {
	info, err := m.Resolve()
	if err != nil {
		return err
	}
	normalized, _, err := normalizeGuidedSessionState(state, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return err
	}
	return writeJSON(info.SessionsFile, normalized)
}

func readGuidedSessionStateFile(path string) (*GuidedSessionState, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read guided session state: %w", err)
	}
	var state GuidedSessionState
	if err := json.Unmarshal(raw, &state); err != nil {
		return nil, fmt.Errorf("parse guided session state: %w", err)
	}
	return &state, nil
}

func ensureGuidedSessionState(path, now string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, err
	}
	state := defaultGuidedSessionState(now)
	return true, writeJSON(path, state)
}

func reconcileGuidedSessionState(path, now string) (bool, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	var state GuidedSessionState
	if err := json.Unmarshal(raw, &state); err != nil {
		return true, writeJSON(path, defaultGuidedSessionState(now))
	}
	normalized, changed, err := normalizeGuidedSessionState(state, now)
	if err != nil {
		return false, err
	}
	if !changed {
		return false, nil
	}
	return true, writeJSON(path, normalized)
}

func defaultGuidedSessionState(now string) GuidedSessionState {
	return GuidedSessionState{
		SchemaVersion: CurrentSchemaVersion,
		LastUpdatedAt: now,
		Sessions:      map[string]GuidedSessionRecord{},
	}
}

func normalizeGuidedSessionState(state GuidedSessionState, now string) (GuidedSessionState, bool, error) {
	if state.SchemaVersion > CurrentSchemaVersion {
		return GuidedSessionState{}, false, fmt.Errorf("guided session schema version %d is newer than supported version %d", state.SchemaVersion, CurrentSchemaVersion)
	}

	normalized := state
	changed := false
	if normalized.SchemaVersion == 0 {
		normalized.SchemaVersion = CurrentSchemaVersion
		changed = true
	}
	if normalized.Sessions == nil {
		normalized.Sessions = map[string]GuidedSessionRecord{}
		changed = true
	}
	if normalized.LastUpdatedAt == "" {
		normalized.LastUpdatedAt = now
		changed = true
	}

	for key, record := range normalized.Sessions {
		recordChanged := false
		if record.ChainID == "" {
			record.ChainID = key
			recordChanged = true
		}
		if record.StageStatuses == nil {
			record.StageStatuses = map[string]string{}
			recordChanged = true
		}
		if record.CreatedAt == "" {
			record.CreatedAt = now
			recordChanged = true
		}
		if record.UpdatedAt == "" {
			record.UpdatedAt = now
			recordChanged = true
		}
		if recordChanged {
			normalized.Sessions[key] = record
			changed = true
		}
	}

	if changed {
		normalized.LastUpdatedAt = now
	}
	return normalized, changed, nil
}

func validateGuidedSessionState(state *GuidedSessionState) error {
	switch {
	case state == nil:
		return fmt.Errorf("guided session state is required")
	case state.SchemaVersion == 0:
		return fmt.Errorf("schema_version is required")
	case state.SchemaVersion > CurrentSchemaVersion:
		return fmt.Errorf("schema_version %d is newer than supported %d", state.SchemaVersion, CurrentSchemaVersion)
	case state.LastUpdatedAt == "":
		return fmt.Errorf("last_updated_at is required")
	case state.Sessions == nil:
		return fmt.Errorf("sessions are required")
	default:
		return nil
	}
}
