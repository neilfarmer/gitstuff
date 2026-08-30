package cmd

import (
	"fmt"
	"strings"
	"time"

	"gitstuff/internal/config"
	"gitstuff/internal/git"
	"gitstuff/internal/paths"
	"gitstuff/internal/scm"
	"gitstuff/internal/verbosity"

	"github.com/spf13/cobra"
)

var cloneCmd = &cobra.Command{
	Use:   "clone [repository-path|group-path]",
	Short: "Clone repositories from configured SCM providers",
	Long: `Clone a specific repository, all repositories, or all repositories in a group/subgroup.

Examples:
  gitstuff clone owner/repo           # Clone specific repository (SSH)
  gitstuff clone --all                # Clone all repositories (SSH)
  gitstuff clone group --all          # Clone all repositories in a group (SSH)
  gitstuff clone group/subgroup --all # Clone all repositories in a subgroup (SSH)
  gitstuff clone owner/repo --https   # Clone specific repository using HTTPS

Repository/group path format: 'owner/repo' or 'group' or 'group/subgroup'`,
	RunE: runClone,
}

func init() {
	rootCmd.AddCommand(cloneCmd)
	cloneCmd.Flags().BoolP("all", "a", false, "Clone all repositories (or all in specified group)")
	cloneCmd.Flags().BoolP("ssh", "s", true, "Use SSH for cloning (default: SSH)")
	cloneCmd.Flags().Bool("https", false, "Use HTTPS for cloning")
	cloneCmd.Flags().BoolP("update", "u", false, "Pull latest changes for already cloned repositories")
}

func runClone(cmd *cobra.Command, args []string) error {
	start := time.Now()
	verbosity.Debug("Starting clone operation with args: %v", args)

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w (run 'gitstuff config' first)", err)
	}
	verbosity.Debug("Loaded configuration with %d providers", len(cfg.Providers))

	if len(cfg.Providers) == 0 {
		return fmt.Errorf("no providers configured")
	}

	// Create clients for all providers
	clients := make([]scm.Client, 0, len(cfg.Providers))
	for _, providerConfig := range cfg.Providers {
		verbosity.Debug("Creating client for provider: %s (%s)", providerConfig.Name, providerConfig.Type)
		client, err := createClient(providerConfig)
		if err != nil {
			return fmt.Errorf("failed to create client for provider %s: %w", providerConfig.Name, err)
		}
		clients = append(clients, client)
	}

	cloneAll, _ := cmd.Flags().GetBool("all")
	useSSH, _ := cmd.Flags().GetBool("ssh")
	useHTTPS, _ := cmd.Flags().GetBool("https")
	update, _ := cmd.Flags().GetBool("update")

	verbosity.Debug("Clone flags: all=%t, ssh=%t, https=%t, update=%t", cloneAll, useSSH, useHTTPS, update)

	// If --https is explicitly set, override SSH default
	if useHTTPS {
		useSSH = false
		verbosity.Debug("Using HTTPS for cloning (SSH disabled)")
	} else {
		verbosity.Debug("Using SSH for cloning")
	}

	if cloneAll && len(args) == 0 {
		verbosity.Info("Cloning all repositories from all providers")
		result := cloneAllRepositories(clients, cfg, useSSH, update)
		verbosity.DebugTiming(start, "Clone all operation completed")
		return result
	}

	if cloneAll && len(args) == 1 {
		verbosity.Info("Cloning all repositories in group: %s", args[0])
		result := cloneGroupRepositories(clients, cfg, args[0], useSSH, update)
		verbosity.DebugTiming(start, "Clone group operation completed")
		return result
	}

	if len(args) == 0 {
		verbosity.Info("No specific repository specified, cloning all repositories")
		result := cloneAllRepositories(clients, cfg, useSSH, update)
		verbosity.DebugTiming(start, "Clone all operation completed")
		return result
	}

	verbosity.Info("Cloning single repository: %s", args[0])
	result := cloneSingleRepository(clients, cfg, args[0], useSSH, update)
	verbosity.DebugTiming(start, "Clone single operation completed")
	return result
}

func cloneAllRepositories(clients []scm.Client, cfg *config.Config, useSSH, update bool) error {
	start := time.Now()
	verbosity.Debug("Collecting repositories from %d providers", len(clients))
	var allRepos []*scm.Repository

	// Collect all repositories from all providers
	for _, client := range clients {
		clientStart := time.Now()
		verbosity.Debug("Fetching repositories from %s provider", client.GetProviderType())
		repos, err := client.ListAllRepositories()
		if err != nil {
			fmt.Printf("❌ Error getting repositories from %s provider: %v\n", client.GetProviderType(), err)
			continue
		}
		verbosity.DebugTiming(clientStart, "Fetched %d repositories from %s provider", len(repos), client.GetProviderType())
		allRepos = append(allRepos, repos...)
	}

	verbosity.DebugTiming(start, "Repository collection completed")
	fmt.Printf("Found %d repositories to clone/update\n\n", len(allRepos))

	successful := 0
	failed := 0

	for i, repo := range allRepos {
		repoStart := time.Now()
		fmt.Printf("[%d/%d] Processing %s [%s]...\n", i+1, len(allRepos), repo.FullPath, repo.Provider)

		// Check if repo exists in either location (new or legacy structure)
		checkPath := paths.ResolveRepositoryPath(cfg, repo)
		verbosity.Debug("Checking repository status at: %s", checkPath)
		status, err := git.GetRepositoryStatus(checkPath)
		if err != nil {
			fmt.Printf("❌ Error checking status: %v\n\n", err)
			failed++
			continue
		}

		if status.Exists && status.IsGitRepo {
			if update {
				verbosity.Debug("Repository exists, pulling latest changes")
				fmt.Printf("🔄 Pulling latest changes...\n")
				pullStart := time.Now()
				if err := git.PullRepository(checkPath); err != nil {
					fmt.Printf("❌ Failed to pull: %v\n\n", err)
					failed++
				} else {
					verbosity.DebugTiming(pullStart, "Pull completed for %s", repo.FullPath)
					fmt.Printf("✅ Updated successfully\n\n")
					successful++
				}
			} else {
				verbosity.Debug("Repository already exists, skipping (no update flag)")
				fmt.Printf("⏭️  Already cloned (use --update to pull latest changes)\n\n")
				successful++
			}
			verbosity.DebugTiming(repoStart, "Processed existing repository: %s", repo.FullPath)
			continue
		}

		cloneURL := repo.CloneURL
		if useSSH {
			cloneURL = repo.SSHCloneURL
		}

		verbosity.Debug("Cloning repository using %s protocol: %s", map[bool]string{true: "SSH", false: "HTTPS"}[useSSH], cloneURL)
		fmt.Printf("📥 Cloning from %s...\n", cloneURL)
		cloneStart := time.Now()
		if err := git.CloneRepository(cloneURL, paths.GetClonePath(cfg, repo), useSSH); err != nil {
			fmt.Printf("❌ Failed to clone: %v\n\n", err)
			failed++
		} else {
			verbosity.DebugTiming(cloneStart, "Clone completed for %s", repo.FullPath)
			fmt.Printf("✅ Cloned successfully\n\n")
			successful++
		}
		verbosity.DebugTiming(repoStart, "Processed new repository: %s", repo.FullPath)
	}

	fmt.Printf("Summary: %d successful, %d failed\n", successful, failed)
	return nil
}

func cloneGroupRepositories(clients []scm.Client, cfg *config.Config, groupPath string, useSSH, update bool) error {
	var allRepos []*scm.Repository

	// Collect repositories from the specified group across all providers
	for _, client := range clients {
		repos, err := client.ListRepositoriesInGroup(groupPath)
		if err != nil {
			continue
		}
		if len(repos) > 0 {
			fmt.Printf("✅ Found %d repositories in %s provider\n", len(repos), client.GetProviderType())
		}
		allRepos = append(allRepos, repos...)
	}

	if len(allRepos) == 0 {
		return fmt.Errorf("no repositories found in group '%s'", groupPath)
	}

	fmt.Printf("Found %d repositories in group '%s' to clone/update\n\n", len(allRepos), groupPath)

	successful := 0
	failed := 0

	for i, repo := range allRepos {
		repoStart := time.Now()
		fmt.Printf("[%d/%d] Processing %s [%s]...\n", i+1, len(allRepos), repo.FullPath, repo.Provider)

		checkPath := paths.ResolveRepositoryPath(cfg, repo)
		verbosity.Debug("Checking repository status at: %s", checkPath)
		status, err := git.GetRepositoryStatus(checkPath)
		if err != nil {
			fmt.Printf("❌ Error checking status: %v\n\n", err)
			failed++
			continue
		}

		if status.Exists && status.IsGitRepo {
			if update {
				verbosity.Debug("Repository exists, pulling latest changes")
				fmt.Printf("🔄 Pulling latest changes...\n")
				pullStart := time.Now()
				if err := git.PullRepository(checkPath); err != nil {
					fmt.Printf("❌ Failed to pull: %v\n\n", err)
					failed++
				} else {
					verbosity.DebugTiming(pullStart, "Pull completed for %s", repo.FullPath)
					fmt.Printf("✅ Updated successfully\n\n")
					successful++
				}
			} else {
				verbosity.Debug("Repository already exists, skipping (no update flag)")
				fmt.Printf("⏭️  Already cloned (use --update to pull latest changes)\n\n")
				successful++
			}
			verbosity.DebugTiming(repoStart, "Processed existing repository: %s", repo.FullPath)
			continue
		}

		cloneURL := repo.CloneURL
		if useSSH {
			cloneURL = repo.SSHCloneURL
		}

		verbosity.Debug("Cloning repository using %s protocol: %s", map[bool]string{true: "SSH", false: "HTTPS"}[useSSH], cloneURL)
		fmt.Printf("📥 Cloning from %s...\n", cloneURL)
		cloneStart := time.Now()
		if err := git.CloneRepository(cloneURL, paths.GetClonePath(cfg, repo), useSSH); err != nil {
			fmt.Printf("❌ Failed to clone: %v\n\n", err)
			failed++
		} else {
			verbosity.DebugTiming(cloneStart, "Clone completed for %s", repo.FullPath)
			fmt.Printf("✅ Cloned successfully\n\n")
			successful++
		}
		verbosity.DebugTiming(repoStart, "Processed new repository: %s", repo.FullPath)
	}

	fmt.Printf("Summary: %d successful, %d failed\n", successful, failed)
	return nil
}

func cloneSingleRepository(clients []scm.Client, cfg *config.Config, repoPath string, useSSH, update bool) error {
	// Search for the repository across all providers
	var foundRepo *scm.Repository

	for _, client := range clients {
		// Try to find the repository in this provider
		repo, err := findRepositoryByPath(client, repoPath)
		if err == nil && repo != nil {
			foundRepo = repo
			break
		}
	}

	if foundRepo == nil {
		return fmt.Errorf("repository '%s' not found in any configured provider", repoPath)
	}

	fmt.Printf("Found repository: %s [%s]\n", foundRepo.FullPath, foundRepo.Provider)

	checkPath := paths.ResolveRepositoryPath(cfg, foundRepo)
	status, err := git.GetRepositoryStatus(checkPath)
	if err != nil {
		return fmt.Errorf("error checking repository status: %w", err)
	}

	if status.Exists && status.IsGitRepo {
		if update {
			fmt.Printf("🔄 Pulling latest changes...\n")
			if err := git.PullRepository(checkPath); err != nil {
				return fmt.Errorf("failed to pull repository: %w", err)
			}
			fmt.Printf("✅ Repository updated successfully\n")
		} else {
			fmt.Printf("⏭️  Repository already cloned at: %s\n", checkPath)
			fmt.Printf("   Use --update flag to pull latest changes\n")
		}
		return nil
	}

	if status.Exists && !status.IsGitRepo {
		return fmt.Errorf("directory %s exists but is not a git repository", checkPath)
	}

	cloneURL := foundRepo.CloneURL
	if useSSH {
		cloneURL = foundRepo.SSHCloneURL
	}

	clonePath := paths.GetClonePath(cfg, foundRepo)
	fmt.Printf("📥 Cloning from %s to %s...\n", cloneURL, clonePath)
	if err := git.CloneRepository(cloneURL, clonePath, useSSH); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	fmt.Printf("✅ Repository cloned successfully\n")
	return nil
}

// findRepositoryByPath searches for a repository by its path (owner/repo format)
func findRepositoryByPath(client scm.Client, repoPath string) (*scm.Repository, error) {
	// Get all repositories from this provider
	repos, err := client.ListAllRepositories()
	if err != nil {
		return nil, err
	}

	// Search for exact match or partial match
	for _, repo := range repos {
		if repo.FullPath == repoPath || strings.HasSuffix(repo.FullPath, repoPath) {
			return repo, nil
		}
	}

	return nil, fmt.Errorf("repository not found")
}
