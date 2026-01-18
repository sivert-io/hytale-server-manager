package hytale

// Default constants for Hytale Server Manager (HSM)

const (
	// DefaultHytaleUser is the default system user for running servers
	DefaultHytaleUser = "hytaleservermanager"

	// DefaultNumServers is the default number of server instances to create
	DefaultNumServers = 1

	// DefaultBasePort is the default game port (Hytale uses UDP)
	DefaultBasePort = 5520

	// DefaultQueryPort is not typically used by Hytale, but kept for consistency
	DefaultQueryPort = 5521

	// DefaultHostnamePrefix for server instances
	DefaultHostnamePrefix = "hytale"

	// DefaultMaxPlayers per server
	DefaultMaxPlayers = 100

	// DefaultMaxViewRadius in chunks (Host Havoc recommendation: 12 chunks = 384 blocks)
	// Default Hytale is 32 chunks, but 12 is recommended for performance
	DefaultMaxViewRadius = 12

	// DefaultGameMode for servers
	DefaultGameMode = "Adventure"

	// DefaultBackupEnabled enables automatic backups
	DefaultBackupEnabled = true

	// DefaultBackupFrequency in minutes (Host Havoc recommendation: 60)
	DefaultBackupFrequency = 60

	// DefaultJVMArgs for server launch
	// Based on Host Havoc optimization guide: https://hosthavoc.com/blog/hytale-server-optimization-guide
	// -Xms and -Xmx should match to avoid memory resizing
	// -XX:+UseG1GC for better garbage collection
	// -XX:AOTCache for faster startup (HytaleServer.aot)
	DefaultJVMArgs = "-Xms6G -Xmx6G -XX:+UseG1GC -XX:AOTCache=HytaleServer.aot"

	// TmuxSessionPrefix for tmux session names
	TmuxSessionPrefix = "hytale-server"

	// DataDirBase is the base directory for server data
	DataDirBase = "/var/lib/hytale"

	// ConfigDir is the shared configuration directory
	ConfigDir = "/etc/hytale"

	// MaxServersPerLicense is the maximum number of servers allowed per game license
	// Per Hytale Server Manual: https://support.hytale.com/hc/en-us/articles/45326769420827-Hytale-Server-Manual
	// Default limit: 100 servers per game license; additional licenses or "Server Provider" account required for more
	MaxServersPerLicense = 100

	// HytaleDownloaderURL is the official URL for downloading hytale-downloader
	HytaleDownloaderURL = "https://downloader.hytale.com/hytale-downloader.zip"
	
	// HytaleDownloaderBinPath is where we install the hytale-downloader binary
	// Installed to /usr/local/bin/hytale-downloader for system-wide access
	HytaleDownloaderBinPath = "/usr/local/bin/hytale-downloader"
)

// BootstrapConfig holds configuration for initial server installation
type BootstrapConfig struct {
	HytaleUser       string
	NumServers       int
	BasePort         int
	QueryPort        int
	HostnamePrefix   string
	ServerPassword   string // Server join password (empty = public)
	MaxPlayers       int
	MaxViewRadius    int    // View distance in chunks
	GameMode         string // Adventure, Survival, or Creative
	JVMArgs          string
	BackupEnabled    bool   // Enable automatic backups
	BackupFrequency  int    // Backup frequency in minutes
	// OAuth credentials for hytale-downloader authentication
	OAuthClientID     string // OAuth client ID (optional)
	OAuthClientSecret string // OAuth client secret (optional)
	OAuthAccessToken  string // OAuth access token (optional, alternative to Client ID/Secret)
}
