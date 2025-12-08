# Async Workspace Loading Feature

## Overview
Optimize workspace project selection startup time by parallelizing data loading operations and adding user-configurable options to disable expensive operations.

## Problem Statement

### Current Behavior (Sequential)
The `ChoseProjectFromLocalWorkspace` function in `workspacer/lists.go` currently runs all operations sequentially:

1. Get open tmux sessions
2. Read workspace directory
3. **Fetch GitHub repos** (network call - SLOW)
4. For each local repo:
   - Get git branch (git command - SLOW)
   - Get uncommitted changes count (git command - SLOW)
   - If tenant repos enabled, repeat for tenant repo (2x SLOW)

### Performance Impact
- **10 repos** with git info: ~1-2 seconds
- **10 repos** with git info + tenant repos: ~2-4 seconds
- **Larger workspaces**: Can exceed 5+ seconds

### User Experience Issues
- No visual feedback during loading
- Active projects don't stay at top during search/filtering
- No way to disable expensive operations

---

## Solution Design

### 1. Parallel Data Loading

**Strategy**: Use goroutines to fetch all data concurrently instead of sequentially.

```
┌─────────────────────────────────────────┐
│ Show Loading Spinner                    │
└─────────────────────────────────────────┘
                  │
    ┌─────────────┴─────────────┐
    │                           │
    ▼                           ▼
┌─────────────┐         ┌──────────────────┐
│ GitHub API  │         │ Git Info         │
│ (goroutine) │         │ (N goroutines)   │
│             │         │                  │
│ Fetch       │         │ For each repo:   │
│ remote      │         │ - Branch         │
│ repos       │         │ - Changes count  │
│             │         │ - Tenant info    │
└─────────────┘         └──────────────────┘
    │                           │
    └─────────────┬─────────────┘
                  ▼
    ┌──────────────────────────┐
    │ sync.WaitGroup           │
    │ (wait for all)           │
    └──────────────────────────┘
                  │
                  ▼
    ┌──────────────────────────┐
    │ Build List Items         │
    └──────────────────────────┘
                  │
                  ▼
    ┌──────────────────────────┐
    │ Transition to List View  │
    └──────────────────────────┘
```

**Expected Performance**:
- Time reduced from `(N repos × git time)` to `max(single git time)`
- **10 repos**: 2 seconds → ~200-500ms (4-10x faster)

### 2. Configuration Options

Add two new fields to `WorkspaceConfig`:

```go
type WorkspaceConfig struct {
    // ... existing fields ...

    // ShowGitInfo enables/disables git branch and uncommitted changes display
    // Default: true (backward compatible)
    ShowGitInfo bool `json:"show_git_info"`

    // FetchRemoteRepos enables/disables GitHub API calls to fetch remote repos
    // Default: true (backward compatible)
    FetchRemoteRepos bool `json:"fetch_remote_repos"`
}
```

**Performance Matrix**:

| ShowGitInfo | FetchRemoteRepos | Estimated Startup Time |
|-------------|------------------|------------------------|
| true        | true             | ~200-500ms (parallel)  |
| true        | false            | ~50-100ms              |
| false       | true             | ~200-400ms             |
| false       | false            | ~10-50ms (instant!)    |

### 3. Loading Spinner

Use Bubble Tea spinner component to provide visual feedback.

**States**:
```
⠋ Loading workspace projects...
⠙ Loading workspace projects...
⠹ Loading workspace projects...
✓ Ready!
```

**Implementation**: Create a loading model that transitions to the list model when data is ready.

### 4. Search Priority for Active Projects

**Problem**: Active projects show first in initial list, but lose priority during search/filtering.

**Solution**: Add `IsActive` field to `list.Item` and use it in `FilterValue()`:

```go
type Item struct {
    Display, Subtitle, Value string
    IsActive                 bool  // NEW
}

func (i Item) FilterValue() string {
    prefix := "1" // inactive
    if i.IsActive {
        prefix = "0" // active items sort first
    }
    return prefix + i.Display + i.Subtitle + i.Value
}
```

This ensures active projects appear first even during filtering.

---

## Implementation Plan

### Files to Modify

#### 1. `config/config.go`
- Add `ShowGitInfo` field to `WorkspaceConfig`
- Add `FetchRemoteRepos` field to `WorkspaceConfig`

#### 2. `ui/list/list.go`
- Add `IsActive` field to `Item` struct
- Update `FilterValue()` to prioritize active items

#### 3. `workspacer/lists.go`
**Major refactoring of `ChoseProjectFromLocalWorkspace`**:

- Import `sync` package for `WaitGroup`
- Create loading spinner view
- Launch GitHub API call in goroutine (if `FetchRemoteRepos == true`)
- Launch git info fetching in goroutines (if `ShowGitInfo == true`)
- Use channels or WaitGroup to synchronize
- Transition from spinner to list when ready
- Mark items as active when building list items

#### 4. `ui/spinner/` (optional - new directory)
- Create reusable loading spinner component
- Handle message passing for data ready state

### Implementation Steps

1. ✅ Create documentation structure
2. Add config options to `WorkspaceConfig`
3. Add `IsActive` field to `list.Item`
4. Create loading spinner component
5. Refactor `ChoseProjectFromLocalWorkspace`:
   - Show spinner first
   - Add parallel data loading
   - Add conditional logic for config options
   - Handle spinner → list transition
6. Test with different configurations

---

## Open Questions

1. **Spinner style**: Simple "Loading..." vs with progress "Loading... (5/10)"?
   - **Recommendation**: Start simple, can enhance later

2. **Timeout behavior**: Should slow operations timeout and show partial results?
   - **Recommendation**: No timeout initially, but could add later with config option

3. **Error handling**: What if GitHub API fails? Show only local repos?
   - **Recommendation**: Continue with local repos only, log error

---

## Testing Strategy

### Test Cases

1. **All features enabled** (ShowGitInfo=true, FetchRemoteRepos=true)
   - Verify spinner shows
   - Verify all repos appear with git info
   - Measure startup time

2. **Only git info** (ShowGitInfo=true, FetchRemoteRepos=false)
   - Verify no GitHub API calls
   - Verify git info still shows
   - Measure startup time

3. **Only remote repos** (ShowGitInfo=false, FetchRemoteRepos=true)
   - Verify no git commands run
   - Verify remote repos appear
   - Measure startup time

4. **Minimal mode** (ShowGitInfo=false, FetchRemoteRepos=false)
   - Verify instant startup
   - Verify only local folders appear

5. **Active project search**
   - Open some tmux sessions
   - Search/filter in list
   - Verify active projects appear first in results

### Performance Benchmarks

Measure before/after with workspace of 10-20 repos:
- Sequential (current): ~2-4 seconds
- Parallel (target): ~200-500ms
- Minimal (target): <50ms

---

## Future Enhancements

1. **Caching**: Cache git info for faster subsequent loads
2. **Incremental loading**: Show list immediately, update items as info arrives
3. **Progress indicator**: Show "Loading... (5/10 repos)" in spinner
4. **Timeout handling**: Configurable timeout for slow operations
5. **Background refresh**: Periodically refresh git info while list is open

---

## References

- Current implementation: `workspacer/lists.go:61-187`
- Git utilities: `util/util.go:230-270`
- GitHub API: `workspacer/github.go:75+`
- List UI: `ui/list/list.go`
