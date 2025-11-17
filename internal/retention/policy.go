package retention

import (
	"fmt"
	"sort"
	"time"

	"nexus-retention-policy/internal/config"
	"nexus-retention-policy/internal/logger"
	"nexus-retention-policy/internal/nexus"
)

type PolicyEngine struct {
	client  *nexus.Client
	config  *config.Config
	logger  *logger.Logger
	dryRun  bool
	verbose bool
}

type ImageGroup struct {
	Name       string
	Components []nexus.Component
}

func NewPolicyEngine(client *nexus.Client, cfg *config.Config, log *logger.Logger, dryRun bool, verbose bool) *PolicyEngine {
	return &PolicyEngine{
		client:  client,
		config:  cfg,
		logger:  log,
		dryRun:  dryRun,
		verbose: verbose,
	}
}

func (p *PolicyEngine) Execute() error {
	fmt.Println("Starting retention policy execution...")
	if p.dryRun {
		fmt.Println("🔍 DRY RUN MODE - No actual deletions will be performed")
	} else {
		fmt.Println("⚠️  EXECUTION MODE - Deletions will be performed")
	}

	repos, err := p.client.GetDockerRepositories()
	if err != nil {
		return fmt.Errorf("failed to get repositories: %w", err)
	}

	fmt.Printf("Found %d Docker hosted repositories\n", len(repos))

	totalDeleted := 0
	totalKept := 0

	for _, repo := range repos {
		fmt.Printf("\n📦 Processing repository: %s\n", repo.Name)

		components, err := p.client.GetComponents(repo.Name)
		if err != nil {
			fmt.Printf("  ⚠️  Error getting components: %v\n", err)
			continue
		}

		fmt.Printf("  Found %d components\n", len(components))

		// Group components by image name
		imageGroups := p.groupByImageName(components)

		for imageName, group := range imageGroups {
			deleted, kept := p.processImageGroup(repo.Name, imageName, group)
			totalDeleted += deleted
			totalKept += kept
		}
	}

	fmt.Printf("\n✅ Execution completed\n")
	fmt.Printf("   Deleted: %d components\n", totalDeleted)
	fmt.Printf("   Kept: %d components\n", totalKept)

	return nil
}

func (p *PolicyEngine) groupByImageName(components []nexus.Component) map[string][]nexus.Component {
	groups := make(map[string][]nexus.Component)

	for _, comp := range components {
		imageName := comp.Name
		groups[imageName] = append(groups[imageName], comp)
	}

	return groups
}

func (p *PolicyEngine) processImageGroup(repoName, imageName string, components []nexus.Component) (deleted, kept int) {
	if len(components) == 0 {
		return 0, 0
	}

	keepCount, ruleName, matched := p.config.GetKeepCount(imageName)

	if !matched {
		if p.verbose {
			fmt.Printf("  ⏭️  Image: %s (no matching rule, skipping)\n", imageName)
		}
		return 0, 0
	}

	fmt.Printf("  🏷️  Image: %s (rule: %s, keep: %d)\n", imageName, ruleName, keepCount)

	// Sort by last modified date (most recent first)
	sort.Slice(components, func(i, j int) bool {
		return p.getLastModified(components[i]).After(p.getLastModified(components[j]))
	})

	// Separate protected and non-protected components
	var protectedComps []nexus.Component
	var regularComps []nexus.Component

	for _, comp := range components {
		if p.config.IsProtected(comp.Version) {
			protectedComps = append(protectedComps, comp)
		} else {
			regularComps = append(regularComps, comp)
		}
	}

	// Keep the most recent N components (excluding protected ones)
	toKeep := regularComps
	toDelete := []nexus.Component{}

	if len(regularComps) > keepCount {
		toKeep = regularComps[:keepCount]
		toDelete = regularComps[keepCount:]
	}

	// Log kept components (in both modes)
	for _, comp := range protectedComps {
		fmt.Printf("     ✓ Keeping %s (protected)\n", comp.Version)
		kept++
	}

	for _, comp := range toKeep {
		fmt.Printf("     ✓ Keeping %s\n", comp.Version)
		kept++
	}

	// Delete old components
	for _, comp := range toDelete {
		if p.dryRun {
			fmt.Printf("     🗑️  Would delete %s\n", comp.Version)
		} else {
			fmt.Printf("     🗑️  Deleting %s\n", comp.Version)
			if err := p.client.DeleteComponent(comp.ID); err != nil {
				fmt.Printf("     ⚠️  Failed to delete: %v\n", err)
				continue
			}
		}

		// Log deletion
		p.logger.LogDeletion(logger.DeletionRecord{
			Timestamp:   time.Now(),
			Repository:  repoName,
			ImageName:   imageName,
			Tag:         comp.Version,
			ComponentID: comp.ID,
			Rule:        ruleName,
			DryRun:      p.dryRun,
		})

		deleted++
	}

	return deleted, kept
}

func (p *PolicyEngine) getLastModified(comp nexus.Component) time.Time {
	if len(comp.Assets) == 0 {
		return time.Time{}
	}

	latest := comp.Assets[0].LastModified
	for _, asset := range comp.Assets {
		if asset.LastModified.After(latest) {
			latest = asset.LastModified
		}
	}

	return latest
}
