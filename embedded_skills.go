package main

import (
	"embed"
	"io/fs"

	"plan/internal/skills"
)

//go:embed skills/plan/**
var embeddedSkills embed.FS

func init() {
	bundle, err := fs.Sub(embeddedSkills, "skills/plan")
	if err == nil {
		skills.RegisterBundle(bundle)
	}
}
