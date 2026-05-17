package planning

import (
	"strings"
	"testing"

	"plan/internal/workspace"
)

func TestRequireLinearTeamConfiguredFailsWithoutTeam(t *testing.T) {
	root := t.TempDir()
	manager := New(workspace.New(root))
	if _, err := manager.workspace.Init(); err != nil {
		t.Fatal(err)
	}

	if _, err := manager.RequireLinearTeamConfigured(); err == nil || !strings.Contains(err.Error(), "team_id or team_key") {
		t.Fatalf("expected actionable missing team error, got %v", err)
	}
}

func TestRequireLinearTeamConfiguredAcceptsTeamKey(t *testing.T) {
	root := t.TempDir()
	manager := New(workspace.New(root))
	if _, err := manager.workspace.Init(); err != nil {
		t.Fatal(err)
	}

	state, err := manager.workspace.ReadLinearState()
	if err != nil {
		t.Fatal(err)
	}
	state.TeamKey = "PLAN"
	if err := manager.workspace.WriteLinearState(*state); err != nil {
		t.Fatal(err)
	}

	linearState, err := manager.RequireLinearTeamConfigured()
	if err != nil {
		t.Fatal(err)
	}
	if linearState.TeamKey != "PLAN" {
		t.Fatalf("expected configured team key: %+v", linearState)
	}
}

func TestSetSourceModeAcceptsLinear(t *testing.T) {
	root := t.TempDir()
	manager := New(workspace.New(root))
	if _, err := manager.workspace.Init(); err != nil {
		t.Fatal(err)
	}

	meta, err := manager.SetSourceMode(workspace.SourceOfTruthLinear)
	if err != nil {
		t.Fatal(err)
	}
	if meta.SourceMode != workspace.SourceOfTruthLinear {
		t.Fatalf("expected linear source mode: %+v", meta)
	}
}
