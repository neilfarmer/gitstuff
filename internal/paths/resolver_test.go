package paths

import (
	"os"
	"path/filepath"
	"testing"

	"gitstuff/internal/config"
	"gitstuff/internal/scm"
)

func TestResolveRepositoryPath(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gitstuff-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.Config{
		Local: config.LocalConfig{
			BaseDir: tempDir,
		},
	}

	repo := &scm.Repository{
		Provider: "gitlab",
		FullPath: "mygroup/myproject",
	}

	tests := []struct {
		name           string
		setupFunc      func() error
		expectedPath   string
		expectProvider bool // true if should use provider-based path
	}{
		{
			name: "Provider-based path exists",
			setupFunc: func() error {
				providerDir := filepath.Join(tempDir, "gitlab", "mygroup")
				if err := os.MkdirAll(providerDir, 0755); err != nil {
					return err
				}
				return os.MkdirAll(filepath.Join(providerDir, "myproject"), 0755)
			},
			expectedPath:   filepath.Join(tempDir, "gitlab", "mygroup", "myproject"),
			expectProvider: true,
		},
		{
			name: "Legacy path exists",
			setupFunc: func() error {
				legacyDir := filepath.Join(tempDir, "mygroup")
				if err := os.MkdirAll(legacyDir, 0755); err != nil {
					return err
				}
				return os.MkdirAll(filepath.Join(legacyDir, "myproject"), 0755)
			},
			expectedPath:   filepath.Join(tempDir, "mygroup", "myproject"),
			expectProvider: false,
		},
		{
			name: "Both paths exist - provider path takes precedence",
			setupFunc: func() error {
				// Create both paths
				providerDir := filepath.Join(tempDir, "gitlab", "mygroup")
				if err := os.MkdirAll(providerDir, 0755); err != nil {
					return err
				}
				if err := os.MkdirAll(filepath.Join(providerDir, "myproject"), 0755); err != nil {
					return err
				}

				legacyDir := filepath.Join(tempDir, "mygroup")
				if err := os.MkdirAll(legacyDir, 0755); err != nil {
					return err
				}
				return os.MkdirAll(filepath.Join(legacyDir, "myproject"), 0755)
			},
			expectedPath:   filepath.Join(tempDir, "gitlab", "mygroup", "myproject"),
			expectProvider: true,
		},
		{
			name: "Neither path exists - returns provider path for new clones",
			setupFunc: func() error {
				// Don't create any directories
				return nil
			},
			expectedPath:   filepath.Join(tempDir, "gitlab", "mygroup", "myproject"),
			expectProvider: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up from previous test
			os.RemoveAll(tempDir)
			_ = os.MkdirAll(tempDir, 0755)

			// Setup test scenario
			if err := tt.setupFunc(); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			// Test the function
			result := ResolveRepositoryPath(cfg, repo)

			if result != tt.expectedPath {
				t.Errorf("ResolveRepositoryPath() = %v, want %v", result, tt.expectedPath)
			}
		})
	}
}

func TestGetClonePath(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gitstuff-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.Config{
		Local: config.LocalConfig{
			BaseDir: tempDir,
		},
	}

	tests := []struct {
		name     string
		repo     *scm.Repository
		expected string
	}{
		{
			name: "GitLab repository",
			repo: &scm.Repository{
				Provider: "gitlab",
				FullPath: "mygroup/myproject",
			},
			expected: filepath.Join(tempDir, "gitlab", "mygroup", "myproject"),
		},
		{
			name: "GitHub repository",
			repo: &scm.Repository{
				Provider: "github",
				FullPath: "myuser/myrepo",
			},
			expected: filepath.Join(tempDir, "github", "myuser", "myrepo"),
		},
		{
			name: "Nested group structure",
			repo: &scm.Repository{
				Provider: "gitlab",
				FullPath: "parentgroup/subgroup/project",
			},
			expected: filepath.Join(tempDir, "gitlab", "parentgroup", "subgroup", "project"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetClonePath(cfg, tt.repo)
			if result != tt.expected {
				t.Errorf("GetClonePath() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPathResolutionWithRealDirectories(t *testing.T) {
	// This test verifies the path resolution works with actual directory structures
	tempDir, err := os.MkdirTemp("", "gitstuff-integration-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.Config{
		Local: config.LocalConfig{
			BaseDir: tempDir,
		},
	}

	// Create a legacy directory structure (like the user's current setup)
	legacyRepo := filepath.Join(tempDir, "cloudservices", "aws")
	if err := os.MkdirAll(legacyRepo, 0755); err != nil {
		t.Fatalf("Failed to create legacy repo dir: %v", err)
	}

	// Create a .git directory to make it look like a real repo
	if err := os.MkdirAll(filepath.Join(legacyRepo, ".git"), 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	repo := &scm.Repository{
		Provider: "gitlab",
		FullPath: "cloudservices/aws",
	}

	// Should find the legacy path
	result := ResolveRepositoryPath(cfg, repo)
	expected := filepath.Join(tempDir, "cloudservices", "aws")

	if result != expected {
		t.Errorf("Expected to find legacy repo at %s, but got %s", expected, result)
	}

	// Verify that GetClonePath still returns provider-based path for new clones
	clonePath := GetClonePath(cfg, repo)
	expectedClone := filepath.Join(tempDir, "gitlab", "cloudservices", "aws")

	if clonePath != expectedClone {
		t.Errorf("Expected clone path %s, but got %s", expectedClone, clonePath)
	}
}
