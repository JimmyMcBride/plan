package skills

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
)

type skillBundle struct {
	Name string
	FS   fs.FS
	Hash string
}

var (
	bundleMu          sync.RWMutex
	registeredBundles map[string]fs.FS
)

var defaultSkillNames = []string{"plan", "plan-execute"}

func RegisterBundle(bundle fs.FS) {
	if bundle == nil {
		RegisterBundles(nil)
		return
	}
	RegisterBundles(map[string]fs.FS{"plan": bundle})
}

func RegisterBundles(bundles map[string]fs.FS) {
	bundleMu.Lock()
	if bundles == nil {
		registeredBundles = nil
	} else {
		registeredBundles = map[string]fs.FS{}
		for name, bundle := range bundles {
			if name == "" || bundle == nil {
				continue
			}
			registeredBundles[name] = bundle
		}
	}
	bundleMu.Unlock()
}

func loadBundles() ([]skillBundle, error) {
	bundleFSes, err := currentBundleFSes()
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(bundleFSes))
	for name := range bundleFSes {
		names = append(names, name)
	}
	sort.Strings(names)

	bundles := make([]skillBundle, 0, len(names))
	for _, name := range names {
		bundleFS := bundleFSes[name]
		if err := validateSkillFS(name, bundleFS); err != nil {
			return nil, err
		}
		hash, err := bundleHash(bundleFS)
		if err != nil {
			return nil, err
		}
		bundles = append(bundles, skillBundle{Name: name, FS: bundleFS, Hash: hash})
	}
	return bundles, nil
}

func currentBundleFSes() (map[string]fs.FS, error) {
	bundleMu.RLock()
	registered := registeredBundles
	var copy map[string]fs.FS
	if len(registered) > 0 {
		copy = map[string]fs.FS{}
		for name, bundle := range registered {
			copy[name] = bundle
		}
	}
	bundleMu.RUnlock()
	if len(copy) > 0 {
		return copy, nil
	}
	return sourceTreeBundles()
}

func sourceTreeBundles() (map[string]fs.FS, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("resolve bundled skill source")
	}
	root := filepath.Dir(filepath.Dir(filepath.Dir(file)))
	bundles := map[string]fs.FS{}
	for _, name := range defaultSkillNames {
		source := filepath.Join(root, "skills", name)
		if err := validateSkillSource(source); err != nil {
			return nil, err
		}
		bundles[name] = os.DirFS(source)
	}
	return bundles, nil
}

func validateSkillSource(source string) error {
	info, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("skill source %s: %w", source, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("skill source is not a directory: %s", source)
	}
	if _, err := os.Stat(filepath.Join(source, "SKILL.md")); err != nil {
		return fmt.Errorf("skill source missing SKILL.md: %w", err)
	}
	return nil
}

func validateSkillFS(name string, bundle fs.FS) error {
	info, err := fs.Stat(bundle, ".")
	if err != nil {
		return fmt.Errorf("skill bundle %s: %w", name, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("skill bundle %s is not a directory", name)
	}
	if _, err := fs.Stat(bundle, "SKILL.md"); err != nil {
		return fmt.Errorf("skill bundle %s missing SKILL.md: %w", name, err)
	}
	return nil
}

func bundleHash(bundle fs.FS) (string, error) {
	files, err := bundleFiles(bundle)
	if err != nil {
		return "", err
	}
	sum := sha256.New()
	for _, path := range files {
		data, err := fs.ReadFile(bundle, path)
		if err != nil {
			return "", fmt.Errorf("read bundled skill file %s: %w", path, err)
		}
		sum.Write([]byte(path))
		sum.Write([]byte{0})
		sum.Write(data)
		sum.Write([]byte{0})
	}
	return hex.EncodeToString(sum.Sum(nil)), nil
}

func bundleFiles(bundle fs.FS) ([]string, error) {
	var files []string
	if err := fs.WalkDir(bundle, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("walk bundled skill: %w", err)
	}
	sort.Strings(files)
	return files, nil
}

func copyBundle(bundle skillBundle, target string) error {
	return fs.WalkDir(bundle.FS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		dest := filepath.Join(target, path)
		if d.IsDir() {
			return os.MkdirAll(dest, 0o755)
		}
		data, err := fs.ReadFile(bundle.FS, path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dest, data, 0o644)
	})
}
