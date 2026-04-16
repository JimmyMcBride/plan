package planning

import (
	"os"
	"path/filepath"
	"testing"

	"plan/internal/workspace"
)

func TestParseRoadmapReadsVersionStructureAndOrder(t *testing.T) {
	roadmap := ParseRoadmap(".plan/ROADMAP.md", `
# Roadmap: plan

## Overview

Alpha overview.

## v1: Local-First Core

Goal: Ship the trustworthy foundation.

- [ ] Core Workspace and Artifact System
- [ ] Spec-Driven Planning Workflow

Summary:
- establish .plan as the workspace
- make the core loop work

## v2: Planning Rigor

Goal: Make plans sharper.

- [x] Roadmap and Portfolio Planning

Summary:
- add roadmap helpers

## Ordering Notes

- v1 first
- v2 second

## Parking Lot

- hosted dashboards
`)

	if roadmap.Path != ".plan/ROADMAP.md" {
		t.Fatalf("unexpected roadmap path: %+v", roadmap)
	}
	if roadmap.Overview != "Alpha overview." {
		t.Fatalf("unexpected overview: %q", roadmap.Overview)
	}
	if len(roadmap.Versions) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(roadmap.Versions))
	}
	if roadmap.Versions[0].Key != "v1" || roadmap.Versions[0].Title != "Local-First Core" {
		t.Fatalf("unexpected first version: %+v", roadmap.Versions[0])
	}
	if roadmap.Versions[0].Goal != "Ship the trustworthy foundation." {
		t.Fatalf("unexpected first goal: %+v", roadmap.Versions[0])
	}
	if len(roadmap.Versions[0].Epics) != 2 || roadmap.Versions[0].Epics[1].Title != "Spec-Driven Planning Workflow" {
		t.Fatalf("unexpected first version epics: %+v", roadmap.Versions[0].Epics)
	}
	if len(roadmap.Versions[1].Epics) != 1 || !roadmap.Versions[1].Epics[0].Done {
		t.Fatalf("expected checked roadmap epic to parse as done: %+v", roadmap.Versions[1].Epics)
	}
	if len(roadmap.OrderingNotes) != 2 || roadmap.OrderingNotes[0] != "v1 first" {
		t.Fatalf("unexpected ordering notes: %+v", roadmap.OrderingNotes)
	}
	if len(roadmap.ParkingLot) != 1 || roadmap.ParkingLot[0] != "hosted dashboards" {
		t.Fatalf("unexpected parking lot: %+v", roadmap.ParkingLot)
	}
}

func TestParseRoadmapToleratesEmptySections(t *testing.T) {
	roadmap := ParseRoadmap(".plan/ROADMAP.md", `
# Roadmap

## Overview

## v1: Core

Goal: Ship core.

## v2: Later

## Ordering Notes

## Parking Lot
`)

	if len(roadmap.Versions) != 2 {
		t.Fatalf("expected empty version sections to still parse, got %+v", roadmap.Versions)
	}
	if roadmap.Versions[0].Key != "v1" || roadmap.Versions[1].Key != "v2" {
		t.Fatalf("expected version order to stay stable: %+v", roadmap.Versions)
	}
	if roadmap.Versions[0].Goal != "Ship core." {
		t.Fatalf("unexpected goal from sparse roadmap: %+v", roadmap.Versions[0])
	}
	if len(roadmap.Versions[1].Epics) != 0 || len(roadmap.OrderingNotes) != 0 || len(roadmap.ParkingLot) != 0 {
		t.Fatalf("expected empty sections to stay empty: %+v", roadmap)
	}
}

func TestReadRoadmapParsesWorkspaceRoadmapFile(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)

	body := `# Roadmap: plan

## Overview

Overview text.

## v1: Local-First Core

Goal: Ship core.

- [ ] Core Workspace and Artifact System

Summary:
- establish .plan

## Ordering Notes

- core first

## Parking Lot

- integrations later
`
	if err := os.WriteFile(filepath.Join(root, ".plan", "ROADMAP.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	roadmap, err := manager.ReadRoadmap()
	if err != nil {
		t.Fatal(err)
	}
	if roadmap.Path != ".plan/ROADMAP.md" {
		t.Fatalf("unexpected roadmap path: %+v", roadmap)
	}
	if len(roadmap.Versions) != 1 || roadmap.Versions[0].Title != "Local-First Core" {
		t.Fatalf("unexpected parsed roadmap versions: %+v", roadmap.Versions)
	}
}
