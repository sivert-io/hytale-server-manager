package hytale

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// HytaleConfig represents the Hytale server configuration
// Based on actual Hytale config.json structure and Host Havoc optimization guide
type HytaleConfig struct {
	Version            int    `json:"Version,omitempty"`
	ServerName         string `json:"ServerName,omitempty"`
	MOTD               string `json:"MOTD,omitempty"` // Message of the Day
	Password           string `json:"Password,omitempty"` // Server password (empty = public)
	MaxPlayers         int    `json:"MaxPlayers,omitempty"`
	MaxViewRadius      int    `json:"MaxViewRadius,omitempty"` // View distance in chunks (Host Havoc: 12 recommended)
	
	// Performance optimization settings (Host Havoc recommendations)
	MaxEntitiesPerChunk int `json:"MaxEntitiesPerChunk,omitempty"` // Default: 50
	MobSpawnLimit       int `json:"MobSpawnLimit,omitempty"`       // Default: 100
	ItemDespawnTime     int `json:"ItemDespawnTime,omitempty"`     // Default: 300 seconds (5 minutes)
	
	Defaults struct {
		World    string `json:"World,omitempty"`
		GameMode string `json:"GameMode,omitempty"` // Adventure, Survival, Creative
	} `json:"Defaults,omitempty"`
	
	ConnectionTimeouts struct {
		JoinTimeouts map[string]interface{} `json:"JoinTimeouts,omitempty"`
	} `json:"ConnectionTimeouts,omitempty"`
	
	RateLimit        map[string]interface{} `json:"RateLimit,omitempty"`
	Modules          map[string]interface{} `json:"Modules,omitempty"`
	LogLevels        map[string]interface{} `json:"LogLevels,omitempty"`
	Mods             map[string]interface{} `json:"Mods,omitempty"`
	DisplayTmpTagsInStrings bool `json:"DisplayTmpTagsInStrings,omitempty"`
	
	PlayerStorage struct {
		Type string `json:"Type,omitempty"`
	} `json:"PlayerStorage,omitempty"`
	
	AuthCredentialStore struct {
		Type string `json:"Type,omitempty"`
		Path string `json:"Path,omitempty"`
	} `json:"AuthCredentialStore,omitempty"`
}

// GetSharedConfigDir returns the path to the shared config directory
func GetSharedConfigDir() string {
	return filepath.Join(DataDirBase, "shared")
}

// GetServerConfigPath returns the path to a server's config.json
func GetServerConfigPath(serverNum int) string {
	return filepath.Join(DataDirBase, fmt.Sprintf("server-%d", serverNum), "config.json")
}

// GetServerDir returns the path to a server's directory
func GetServerDir(serverNum int) string {
	return filepath.Join(DataDirBase, fmt.Sprintf("server-%d", serverNum))
}

// ReadConfig reads a config.json file
func ReadConfig(configPath string) (*HytaleConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config HytaleConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// WriteConfig writes a config.json file
func WriteConfig(configPath string, config *HytaleConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// CreateDefaultConfig creates a default config with the given settings
// Uses Host Havoc optimization recommendations: https://hosthavoc.com/blog/hytale-server-optimization-guide
func CreateDefaultConfig(port int, hostname string, maxPlayers int, maxViewRadius int, gameMode string, serverPassword string, adminPassword string) *HytaleConfig {
	config := &HytaleConfig{}
	config.Version = 3
	config.ServerName = hostname
	config.MOTD = fmt.Sprintf("Welcome to %s", hostname)
	config.Password = serverPassword // Server join password (empty = public)
	config.MaxPlayers = maxPlayers
	config.MaxViewRadius = maxViewRadius
	
	// Performance optimization settings (Host Havoc recommendations)
	config.MaxEntitiesPerChunk = 50   // Limit entities per chunk
	config.MobSpawnLimit = 100       // Limit mob spawning
	config.ItemDespawnTime = 300     // 5 minutes despawn time
	
	// Defaults
	config.Defaults.World = "default"
	if gameMode == "" {
		gameMode = "Adventure"
	}
	config.Defaults.GameMode = gameMode
	
	// Initialize maps
	config.ConnectionTimeouts.JoinTimeouts = make(map[string]interface{})
	config.RateLimit = make(map[string]interface{})
	config.Modules = make(map[string]interface{})
	config.LogLevels = make(map[string]interface{})
	config.Mods = make(map[string]interface{})
	config.DisplayTmpTagsInStrings = false
	
	// Player storage
	config.PlayerStorage.Type = "Hytale"
	
	// Auth credential store
	config.AuthCredentialStore.Type = "Encrypted"
	config.AuthCredentialStore.Path = "auth.enc"
	
	return config
}

// UpdateServerConfig updates a server's config.json with server-specific settings
// Preserves optimization settings from Host Havoc guide
func UpdateServerConfig(serverNum int, port int, hostname string, maxPlayers int, maxViewRadius int, gameMode string, serverPassword string) error {
	configPath := GetServerConfigPath(serverNum)
	
	// Read existing config or create new one
	config, err := ReadConfig(configPath)
	if err != nil {
		// Config doesn't exist, create default with optimizations
		config = CreateDefaultConfig(port, hostname, maxPlayers, maxViewRadius, gameMode, serverPassword, "")
	} else {
		// Update existing config (preserve optimization settings if they exist)
		config.ServerName = hostname
		config.MaxPlayers = maxPlayers
		config.MaxViewRadius = maxViewRadius
		if serverPassword != "" {
			config.Password = serverPassword
		}
		if gameMode != "" {
			config.Defaults.GameMode = gameMode
		}
		
		// Ensure optimization settings are set (use defaults if not present)
		if config.MaxViewRadius == 0 {
			config.MaxViewRadius = 12
		}
		if config.MaxEntitiesPerChunk == 0 {
			config.MaxEntitiesPerChunk = 50
		}
		if config.MobSpawnLimit == 0 {
			config.MobSpawnLimit = 100
		}
		if config.ItemDespawnTime == 0 {
			config.ItemDespawnTime = 300
		}
	}

	return WriteConfig(configPath, config)
}

// GetServerPortFromConfig reads the port from config.json
// Note: Hytale doesn't store port in config.json, it's passed via --bind
// This function is kept for consistency but port comes from command line
func GetServerPortFromConfig(serverNum int) (int, error) {
	// Port is not stored in config.json for Hytale
	// It's passed via --bind argument
	// This is a placeholder for future use
	return 0, fmt.Errorf("port is not stored in config.json, use --bind argument")
}
