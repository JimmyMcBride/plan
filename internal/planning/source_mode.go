package planning

import (
	"fmt"
	"time"

	"plan/internal/workspace"
)

func (m *Manager) SourceMode() (workspace.SourceOfTruthMode, error) {
	meta, err := m.workspace.ReadWorkspaceMeta()
	if err != nil {
		return "", err
	}
	if meta.SourceMode == "" {
		return workspace.SourceOfTruthLocal, nil
	}
	return meta.SourceMode, nil
}

func (m *Manager) SetSourceMode(mode workspace.SourceOfTruthMode) (*workspace.WorkspaceMeta, error) {
	if mode != workspace.SourceOfTruthLocal && mode != workspace.SourceOfTruthGitHub && mode != workspace.SourceOfTruthHybrid && mode != workspace.SourceOfTruthLinear {
		return nil, fmt.Errorf("unsupported source mode %q", mode)
	}
	meta, err := m.workspace.ReadWorkspaceMeta()
	if err != nil {
		return nil, err
	}
	meta.SourceMode = mode
	meta.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := m.workspace.WriteWorkspaceMeta(*meta); err != nil {
		return nil, err
	}
	return meta, nil
}

func (m *Manager) LinearConfig() (*workspace.LinearState, error) {
	return m.workspace.ReadLinearState()
}

func (m *Manager) RequireLinearTeamConfigured() (*workspace.LinearState, error) {
	state, err := m.workspace.ReadLinearState()
	if err != nil {
		return nil, err
	}
	if state.TeamID == "" && state.TeamKey == "" {
		return nil, fmt.Errorf("linear promotion requires a selected Linear team; configure .plan/.meta/linear.json with team_id or team_key before applying Linear-owned planning")
	}
	return state, nil
}
