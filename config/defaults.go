package config

// BlankConfig is written to disk by `config new`
var BlankConfig = GlobalUserConfig{
	Workspaces:     map[string]WorkspaceConfig{},
	SessionPresets: map[string]SessionConfig{},
}
