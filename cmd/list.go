package cmd

import (
	"fmt"
	"strings"
	"time"

	"gitstuff/internal/config"
	"gitstuff/internal/git"
	"gitstuff/internal/github"
	"gitstuff/internal/gitlab"
	"gitstuff/internal/paths"
	"gitstuff/internal/scm"
	"gitstuff/internal/verbosity"

	"github.com/spf13/cobra"
)

// createClient creates an SCM client based on the provider config
func createClient(providerConfig config.ProviderConfig) (scm.Client, error) {
	switch providerConfig.Type {
	case "gitlab":
		return gitlab.NewClient(providerConfig.URL, providerConfig.Token, providerConfig.Insecure)
	case "github":
		return github.NewClient(providerConfig.URL, providerConfig.Token, providerConfig.Insecure)
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerConfig.Type)
	}
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all repositories from configured SCM providers",
	Long:  `List all repositories from GitLab and GitHub with their status including clone status and current branch.`,
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolP("tree", "t", false, "Display repositories in tree structure with groups")
	listCmd.Flags().BoolP("status", "s", true, "Show local repository status")
	listCmd.Flags().StringP("group", "g", "", "Filter repositories to only those in the specified group")
}

func runList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w (run 'gitstuff config' first)", err)
	}

	// Create clients for all configured providers
	var clients []scm.Client
	for _, providerConfig := range cfg.Providers {
		client, err := createClient(providerConfig)
		if err != nil {
			return fmt.Errorf("failed to create client for provider %s: %w", providerConfig.Name, err)
		}
		clients = append(clients, client)
	}

	showTree, _ := cmd.Flags().GetBool("tree")
	showStatus, _ := cmd.Flags().GetBool("status")
	groupFilter, _ := cmd.Flags().GetString("group")

	// Use group from flag first, then from any provider config, then empty string
	targetGroup := groupFilter
	if targetGroup == "" {
		for _, providerConfig := range cfg.Providers {
			if providerConfig.Group != "" {
				targetGroup = providerConfig.Group
				break
			}
		}
	}

	if showTree {
		return displayRepositoryTree(clients, cfg, showStatus, targetGroup)
	} else {
		return displayRepositoryList(clients, cfg, showStatus, targetGroup)
	}
}

func displayRepositoryList(clients []scm.Client, cfg *config.Config, showStatus bool, groupFilter string) error {
	start := time.Now()
	verbosity.Debug("Starting repository list from %d providers", len(clients))

	var allRepos []*scm.Repository

	for _, client := range clients {
		var repos []*scm.Repository
		var err error

		clientStart := time.Now()
		if groupFilter != "" {
			verbosity.Debug("Fetching repositories from %s provider in group: %s", client.GetProviderType(), groupFilter)
			repos, err = client.ListRepositoriesInGroup(groupFilter)
		} else {
			verbosity.Debug("Fetching all repositories from %s provider", client.GetProviderType())
			repos, err = client.ListAllRepositories()
		}
		if err != nil {
			return fmt.Errorf("error from %s provider: %w", client.GetProviderType(), err)
		}
		verbosity.DebugTiming(clientStart, "Fetched %d repositories from %s provider", len(repos), client.GetProviderType())
		allRepos = append(allRepos, repos...)
	}

	verbosity.DebugTiming(start, "Repository discovery completed")
	fmt.Printf("Found %d repositories:\n\n", len(allRepos))

	for _, repo := range allRepos {
		fmt.Printf("📁 [%s] %s\n", repo.Provider, repo.FullPath)

		if verbosity.IsEnabled(verbosity.InfoLevel) {
			fmt.Printf("   Web URL: %s\n", repo.WebURL)
			fmt.Printf("   SSH URL: %s\n", repo.SSHCloneURL)
		}

		if verbosity.IsEnabled(verbosity.DebugLevel) {
			fmt.Printf("   Clone URL: %s\n", repo.CloneURL)
			fmt.Printf("   Default Branch: %s\n", repo.DefaultBranch)
			fmt.Printf("   Provider: %s\n", repo.Provider)
		}

		if showStatus {
			localPath := paths.ResolveRepositoryPath(cfg, repo)
			status, err := git.GetRepositoryStatus(localPath)
			if err != nil {
				fmt.Printf("   Status: ❌ Error checking status: %v\n", err)
			} else {
				displayStatus(status)
			}
		}

		fmt.Print("\n")
	}

	return nil
}

func displayRepositoryTree(clients []scm.Client, cfg *config.Config, showStatus bool, groupFilter string) error {
	fmt.Println("Repository tree structure:")

	for _, client := range clients {
		fmt.Printf("\n=== %s Provider ===\n", strings.ToUpper(client.GetProviderType()))

		tree, err := client.BuildRepositoryTree()
		if err != nil {
			fmt.Printf("Error building tree for %s: %v\n", client.GetProviderType(), err)
			continue
		}

		if groupFilter != "" {
			fmt.Printf("(filtered by group: %s)\n", groupFilter)
			displayFilteredTree(tree, groupFilter, cfg, showStatus, client.GetProviderType())
		} else {
			if len(tree.Repositories) > 0 {
				fmt.Println("Root repositories:")
				for _, repo := range tree.Repositories {
					repoLine := fmt.Sprintf("📁 %s", repo.Name)

					if showStatus {
						localPath := paths.ResolveRepositoryPath(cfg, repo)
						status, err := git.GetRepositoryStatus(localPath)
						if err != nil {
							repoLine += fmt.Sprintf(" - ❌ Error: %v", err)
						} else {
							repoLine += " - " + getCompactStatus(status, repo.DefaultBranch)
						}
					}

					fmt.Println(repoLine)

					if verbosity.IsEnabled(verbosity.InfoLevel) {
						fmt.Printf("   Web URL: %s\n", repo.WebURL)
						fmt.Printf("   SSH URL: %s\n", repo.SSHCloneURL)
					}
				}
			}

			for groupName, groupNode := range tree.Groups {
				displayGroup(groupNode, 0, cfg, showStatus)
				_ = groupName
			}
		}
	}

	return nil
}

func displayFilteredTree(tree *scm.RepositoryTree, groupFilter string, cfg *config.Config, showStatus bool, providerType string) {
	targetGroup := findGroupInTree(tree, groupFilter)
	if targetGroup != nil {
		displayGroup(targetGroup, 0, cfg, showStatus)
	} else {
		fmt.Printf("Group '%s' not found in %s\n", groupFilter, providerType)
	}
}

func findGroupInTree(tree *scm.RepositoryTree, groupPath string) *scm.GroupNode {
	parts := strings.Split(groupPath, "/")

	current := tree.Groups
	var currentNode *scm.GroupNode

	for _, part := range parts {
		if node, exists := current[part]; exists {
			currentNode = node
			current = node.SubGroups
		} else {
			return nil
		}
	}

	return currentNode
}

func displayGroup(group *scm.GroupNode, indent int, cfg *config.Config, showStatus bool) {
	prefix := strings.Repeat("  ", indent)
	fmt.Printf("%s📂 %s/\n", prefix, group.Group.Name)

	for _, repo := range group.Repositories {
		repoLine := fmt.Sprintf("%s  📁 %s", prefix, repo.Name)

		if showStatus {
			localPath := paths.ResolveRepositoryPath(cfg, repo)
			status, err := git.GetRepositoryStatus(localPath)
			if err != nil {
				repoLine += fmt.Sprintf(" - ❌ Error: %v", err)
			} else {
				repoLine += " - " + getCompactStatus(status, repo.DefaultBranch)
			}
		}

		fmt.Println(repoLine)

		if verbosity.IsEnabled(verbosity.InfoLevel) {
			fmt.Printf("%s     Web URL: %s\n", prefix, repo.WebURL)
			fmt.Printf("%s     SSH URL: %s\n", prefix, repo.SSHCloneURL)
		}
	}

	for _, subGroup := range group.SubGroups {
		displayGroup(subGroup, indent+1, cfg, showStatus)
	}
}

func getCompactStatus(status *git.Status, defaultBranch string) string {
	if !status.Exists {
		return "❌ Not cloned"
	}

	if !status.IsGitRepo {
		return "⚠️ Not a git repo"
	}

	result := "✅"
	if status.HasChanges {
		result += " 🔄"
	}
	if status.CurrentBranch != "" {
		// Only show branch name if it's not the default branch and not main/master
		if !isDefaultBranch(status.CurrentBranch, defaultBranch) {
			result += fmt.Sprintf(" (%s)", status.CurrentBranch)
		}
	}
	return result
}

func isDefaultBranch(currentBranch, defaultBranch string) bool {
	// Check against the repo's actual default branch
	if defaultBranch != "" && currentBranch == defaultBranch {
		return true
	}
	// Also check against common default branch names
	return currentBranch == "main" || currentBranch == "master"
}

func displayStatus(status *git.Status) {
	if !status.Exists {
		fmt.Print("Status: ❌ Not cloned\n")
		return
	}

	if !status.IsGitRepo {
		fmt.Print("Status: ⚠️  Directory exists but not a git repository\n")
		return
	}

	fmt.Printf("Status: ✅ Cloned")
	if status.CurrentBranch != "" {
		fmt.Printf(" (branch: %s)", status.CurrentBranch)
	}
	if status.HasChanges {
		fmt.Print(" 🔄 Has uncommitted changes")
	}
	fmt.Print("\n")
}
