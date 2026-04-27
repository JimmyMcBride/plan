package main

import (
	"embed"
	"io/fs"

	"plan/internal/skills"
)

//go:embed skills/plan/**
//go:embed skills/plan-execute/**
var embeddedSkills embed.FS

func init() {
	bundles := map[string]fs.FS{}
	for _, name := range []string{"plan", "plan-execute"} {
		bundle, err := fs.Sub(embeddedSkills, "skills/"+name)
		if err == nil {
			bundles[name] = bundle
		}
	}
	skills.RegisterBundles(bundles)
}
