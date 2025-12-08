package workspacer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/JamesTiberiusKirk/workspacer/config"
	"github.com/JamesTiberiusKirk/workspacer/log"
	"github.com/JamesTiberiusKirk/workspacer/util"
)

const (
	cacheFileName = ".workspacer-cache.json"
)

// ProjectCache holds cached information for a single project
type ProjectCache struct {
	GitBranch         string `json:"git_branch,omitempty"`
	GitChanges        int    `json:"git_changes,omitempty"`
	TenantBranch      string `json:"tenant_branch,omitempty"`
	TenantChanges     int    `json:"tenant_changes,omitempty"`
	AccessCountTotal  int    `json:"access_count_total"`
	AccessCountRecent int    `json:"access_count_recent"`
}

// AccessRecord tracks a single project access
type AccessRecord struct {
	Project   string    `json:"project"`
	Timestamp time.Time `json:"timestamp"`
}

// WorkspaceCache holds all cached data for a workspace
type WorkspaceCache struct {
	LastUpdated        time.Time                `json:"last_updated"`
	Projects           map[string]ProjectCache  `json:"projects"`
	GithubRepos        []string                 `json:"github_repos,omitempty"`
	GithubReposUpdated time.Time                `json:"github_repos_updated,omitempty"`
	RecentAccesses     []AccessRecord           `json:"recent_accesses"`
}

// GetCachePath returns the full path to the cache file for a workspace
func GetCachePath(wc config.WorkspaceConfig) string {
	return filepath.Join(util.GetWorkspacePath(wc), cacheFileName)
}

// LoadCache loads the cache from disk, returns empty cache if not found or invalid
func LoadCache(wc config.WorkspaceConfig) *WorkspaceCache {
	if !wc.EnableCache {
		return &WorkspaceCache{
			Projects:       make(map[string]ProjectCache),
			RecentAccesses: []AccessRecord{},
		}
	}

	cachePath := GetCachePath(wc)
	data, err := os.ReadFile(cachePath)
	if err != nil {
		// Cache doesn't exist yet, return empty cache
		return &WorkspaceCache{
			Projects:       make(map[string]ProjectCache),
			RecentAccesses: []AccessRecord{},
		}
	}

	var cache WorkspaceCache
	if err := json.Unmarshal(data, &cache); err != nil {
		log.Error("Failed to parse cache file, ignoring: %s", err.Error())
		// Return empty cache on parse error
		return &WorkspaceCache{
			Projects:       make(map[string]ProjectCache),
			RecentAccesses: []AccessRecord{},
		}
	}

	// Initialize maps if nil
	if cache.Projects == nil {
		cache.Projects = make(map[string]ProjectCache)
	}
	if cache.RecentAccesses == nil {
		cache.RecentAccesses = []AccessRecord{}
	}

	return &cache
}

// SaveCache writes the cache to disk
func SaveCache(wc config.WorkspaceConfig, cache *WorkspaceCache) error {
	if !wc.EnableCache {
		return nil
	}

	cache.LastUpdated = time.Now()

	cachePath := GetCachePath(wc)
	data, err := json.MarshalIndent(cache, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// UpdateGitInfo updates git information for a project in the cache
func (c *WorkspaceCache) UpdateGitInfo(projectName string, info repoGitInfo) {
	project, exists := c.Projects[projectName]
	if !exists {
		project = ProjectCache{}
	}

	project.GitBranch = info.branch
	project.GitChanges = info.changesCount
	if info.hasTenant {
		project.TenantBranch = info.tenantBranch
		project.TenantChanges = info.tenantChanges
	}

	c.Projects[projectName] = project
}

// UpdateGithubRepos updates the GitHub repos list in the cache
func (c *WorkspaceCache) UpdateGithubRepos(repos []string) {
	c.GithubRepos = repos
	c.GithubReposUpdated = time.Now()
}

// RecordAccess records a project access and updates usage statistics
func (c *WorkspaceCache) RecordAccess(projectName string, windowSize int) {
	// Add new access record
	record := AccessRecord{
		Project:   projectName,
		Timestamp: time.Now(),
	}
	c.RecentAccesses = append(c.RecentAccesses, record)

	// Keep only last N accesses (sliding window)
	if len(c.RecentAccesses) > windowSize {
		c.RecentAccesses = c.RecentAccesses[len(c.RecentAccesses)-windowSize:]
	}

	// Update project stats
	project, exists := c.Projects[projectName]
	if !exists {
		project = ProjectCache{}
	}

	// Increment total count
	project.AccessCountTotal++

	// Count recent accesses for this project
	recentCount := 0
	for _, acc := range c.RecentAccesses {
		if acc.Project == projectName {
			recentCount++
		}
	}
	project.AccessCountRecent = recentCount

	c.Projects[projectName] = project
}

// GetProjectCache retrieves cache for a specific project
func (c *WorkspaceCache) GetProjectCache(projectName string) (ProjectCache, bool) {
	project, exists := c.Projects[projectName]
	return project, exists
}

// ClearCache deletes the cache file from disk
func ClearCache(wc config.WorkspaceConfig) error {
	cachePath := GetCachePath(wc)
	if err := os.Remove(cachePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete cache file: %w", err)
	}
	return nil
}

// GetCacheStats returns statistics about the cache
func GetCacheStats(wc config.WorkspaceConfig) (map[string]interface{}, error) {
	cachePath := GetCachePath(wc)

	stats := make(map[string]interface{})
	stats["cache_path"] = cachePath

	fileInfo, err := os.Stat(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			stats["exists"] = false
			return stats, nil
		}
		return nil, err
	}

	stats["exists"] = true
	stats["size_bytes"] = fileInfo.Size()
	stats["modified"] = fileInfo.ModTime()

	cache := LoadCache(wc)
	stats["num_projects"] = len(cache.Projects)
	stats["num_github_repos"] = len(cache.GithubRepos)
	stats["num_recent_accesses"] = len(cache.RecentAccesses)
	stats["last_updated"] = cache.LastUpdated

	return stats, nil
}
