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

	// DefaultJVMArgs for server launch
	DefaultJVMArgs = "-Xms4G -Xmx8G -XX:+UseG1GC"

	// TmuxSessionPrefix for tmux session names
	TmuxSessionPrefix = "hytale-server"

	// DataDirBase is the base directory for server data
	DataDirBase = "/var/lib/hytale"

	// ConfigDir is the shared configuration directory
	ConfigDir = "/etc/hytale"
)

// BootstrapConfig holds configuration for initial server installation
type BootstrapConfig struct {
	HytaleUser     string
	NumServers     int
	BasePort       int
	QueryPort      int
	HostnamePrefix string
	AdminPassword  string
	MaxPlayers     int
	JVMArgs        string
}
