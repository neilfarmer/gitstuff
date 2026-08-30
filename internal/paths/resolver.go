package paths

import (
	"os"
	"path/filepath"

	"gitstuff/internal/config"
	"gitstuff/internal/scm"
	"gitstuff/internal/verbosity"
)

// ResolveRepositoryPath determines the correct local path for a repository.
// It first tries the new provider-based structure: {BaseDir}/{Provider}/{FullPath}
// If that doesn't exist, it falls back to legacy structure: {BaseDir}/{FullPath}
func ResolveRepositoryPath(cfg *config.Config, repo *scm.Repository) string {
	// New provider-based structure (current default)
	providerPath := filepath.Join(cfg.Local.BaseDir, repo.Provider, repo.FullPath)

	verbosity.Trace("Checking provider-based path: %s", providerPath)
	if _, err := os.Stat(providerPath); err == nil {
		verbosity.Debug("Found repository at provider-based path: %s", providerPath)
		return providerPath
	}

	// Legacy structure fallback
	legacyPath := filepath.Join(cfg.Local.BaseDir, repo.FullPath)
	verbosity.Trace("Checking legacy path: %s", legacyPath)
	if _, err := os.Stat(legacyPath); err == nil {
		verbosity.Debug("Found repository at legacy path: %s", legacyPath)
		return legacyPath
	}

	// If neither exists, return the provider-based path (for new clones)
	verbosity.Debug("Repository not found at either path, returning provider-based path for potential clone: %s", providerPath)
	return providerPath
}

// GetClonePath returns the path where a new repository should be cloned.
// This always uses the provider-based structure for new clones to maintain consistency.
func GetClonePath(cfg *config.Config, repo *scm.Repository) string {
	path := filepath.Join(cfg.Local.BaseDir, repo.Provider, repo.FullPath)
	verbosity.Debug("Clone path for %s: %s", repo.FullPath, path)
	return path
}
