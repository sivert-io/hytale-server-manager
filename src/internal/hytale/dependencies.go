package hytale

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Dependency represents a required system dependency
type Dependency struct {
	Name        string
	Description string
	CheckCmd    []string // Command to check if installed (e.g., ["java", "-version"])
	InstallCmd  []string // Command to install (varies by package manager)
	Required    bool     // Whether this is required or optional
}

// PackageManager represents a system package manager
type PackageManager int

const (
	PackageManagerUnknown PackageManager = iota
	PackageManagerAPT     // Debian/Ubuntu
	PackageManagerYUM     // RHEL/CentOS (older)
	PackageManagerDNF     // RHEL/CentOS/Fedora (newer)
	PackageManagerPACMAN // Arch Linux
	PackageManagerZYPPER  // openSUSE
)

// DetectPackageManager detects the system's package manager
func DetectPackageManager() PackageManager {
	// Check for package managers in order of preference
	if _, err := exec.LookPath("apt"); err == nil {
		return PackageManagerAPT
	}
	if _, err := exec.LookPath("dnf"); err == nil {
		return PackageManagerDNF
	}
	if _, err := exec.LookPath("yum"); err == nil {
		return PackageManagerYUM
	}
	if _, err := exec.LookPath("pacman"); err == nil {
		return PackageManagerPACMAN
	}
	if _, err := exec.LookPath("zypper"); err == nil {
		return PackageManagerZYPPER
	}
	return PackageManagerUnknown
}

// GetRequiredDependencies returns the list of required dependencies
func GetRequiredDependencies() []Dependency {
	return []Dependency{
		{
			Name:        "java",
			Description: "Java Runtime Environment (JRE) 17 or later",
			CheckCmd:    []string{"java", "-version"},
			Required:    true,
		},
		{
			Name:        "tmux",
			Description: "Terminal multiplexer for server process management",
			CheckCmd:    []string{"tmux", "-V"},
			Required:    true,
		},
		{
			Name:        "wget",
			Description: "File downloader (for plugin downloads)",
			CheckCmd:    []string{"wget", "--version"},
			Required:    false, // Optional, curl can be used instead
		},
		{
			Name:        "curl",
			Description: "File downloader (alternative to wget)",
			CheckCmd:    []string{"curl", "--version"},
			Required:    false, // Optional, wget can be used instead
		},
	}
}

// CheckDependency checks if a dependency is installed
func CheckDependency(dep Dependency) (bool, string, error) {
	if len(dep.CheckCmd) == 0 {
		return false, "", fmt.Errorf("no check command defined for %s", dep.Name)
	}

	cmd := exec.Command(dep.CheckCmd[0], dep.CheckCmd[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, "", nil // Not installed, but not an error
	}

	version := strings.TrimSpace(string(output))
	return true, version, nil
}

// CheckAllDependencies checks all required dependencies
func CheckAllDependencies() ([]DependencyStatus, error) {
	deps := GetRequiredDependencies()
	statuses := make([]DependencyStatus, 0, len(deps))

	for _, dep := range deps {
		installed, version, err := CheckDependency(dep)
		if err != nil {
			return nil, fmt.Errorf("failed to check %s: %w", dep.Name, err)
		}

		statuses = append(statuses, DependencyStatus{
			Dependency: dep,
			Installed:  installed,
			Version:    version,
		})
	}

	return statuses, nil
}

// DependencyStatus represents the status of a dependency check
type DependencyStatus struct {
	Dependency Dependency
	Installed  bool
	Version    string
	Error      error
}

// InstallDependency installs a dependency using the system package manager
func InstallDependency(ctx context.Context, dep Dependency, progressCallback ProgressCallback) error {
	pm := DetectPackageManager()
	if pm == PackageManagerUnknown {
		return fmt.Errorf("unknown package manager - cannot install dependencies automatically")
	}

	var installCmd []string
	var installArgs []string

	switch pm {
	case PackageManagerAPT:
		installCmd = []string{"apt-get", "install", "-y"}
		switch dep.Name {
		case "java":
			// Try to install OpenJDK 17 or later
			installArgs = []string{"openjdk-17-jre-headless"}
		case "tmux":
			installArgs = []string{"tmux"}
		case "wget":
			installArgs = []string{"wget"}
		case "curl":
			installArgs = []string{"curl"}
		default:
			return fmt.Errorf("unknown dependency: %s", dep.Name)
		}

	case PackageManagerDNF, PackageManagerYUM:
		cmdName := "dnf"
		if pm == PackageManagerYUM {
			cmdName = "yum"
		}
		installCmd = []string{cmdName, "install", "-y"}
		switch dep.Name {
		case "java":
			// Try to install OpenJDK 17 or later
			installArgs = []string{"java-17-openjdk-headless"}
		case "tmux":
			installArgs = []string{"tmux"}
		case "wget":
			installArgs = []string{"wget"}
		case "curl":
			installArgs = []string{"curl"}
		default:
			return fmt.Errorf("unknown dependency: %s", dep.Name)
		}

	case PackageManagerPACMAN:
		installCmd = []string{"pacman", "-S", "--noconfirm"}
		switch dep.Name {
		case "java":
			installArgs = []string{"jre-openjdk"}
		case "tmux":
			installArgs = []string{"tmux"}
		case "wget":
			installArgs = []string{"wget"}
		case "curl":
			installArgs = []string{"curl"}
		default:
			return fmt.Errorf("unknown dependency: %s", dep.Name)
		}

	case PackageManagerZYPPER:
		installCmd = []string{"zypper", "install", "-y"}
		switch dep.Name {
		case "java":
			installArgs = []string{"java-17-openjdk-headless"}
		case "tmux":
			installArgs = []string{"tmux"}
		case "wget":
			installArgs = []string{"wget"}
		case "curl":
			installArgs = []string{"curl"}
		default:
			return fmt.Errorf("unknown dependency: %s", dep.Name)
		}
	}

	// Check if running as root
	if os.Geteuid() != 0 {
		// Prepend sudo
		installCmd = append([]string{"sudo"}, installCmd...)
	}

	// Combine command and args
	fullCmd := append(installCmd, installArgs...)
	cmd := exec.CommandContext(ctx, fullCmd[0], fullCmd[1:]...)

	if progressCallback != nil {
		progressCallback(0.0, fmt.Sprintf("Installing %s...", dep.Name))
	}

	// Run installation
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install %s: %w\nOutput: %s", dep.Name, err, string(output))
	}

	if progressCallback != nil {
		progressCallback(1.0, fmt.Sprintf("%s installed successfully", dep.Name))
	}

	// Verify installation
	installed, _, checkErr := CheckDependency(dep)
	if checkErr != nil {
		return fmt.Errorf("failed to verify installation of %s: %w", dep.Name, checkErr)
	}
	if !installed {
		return fmt.Errorf("%s installation completed but dependency not found", dep.Name)
	}

	return nil
}

// InstallAllDependencies installs all missing required dependencies
func InstallAllDependencies(ctx context.Context, progressCallback ProgressCallback) error {
	statuses, err := CheckAllDependencies()
	if err != nil {
		return fmt.Errorf("failed to check dependencies: %w", err)
	}

	// Filter to only missing required dependencies
	missing := make([]Dependency, 0)
	for _, status := range statuses {
		if !status.Installed && status.Dependency.Required {
			missing = append(missing, status.Dependency)
		}
	}

	if len(missing) == 0 {
		if progressCallback != nil {
			progressCallback(1.0, "All dependencies are installed")
		}
		return nil
	}

	// Install each missing dependency
	total := len(missing)
	for i, dep := range missing {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if progressCallback != nil {
			progress := float64(i) / float64(total)
			progressCallback(progress, fmt.Sprintf("Installing %s (%d/%d)...", dep.Name, i+1, total))
		}

		if err := InstallDependency(ctx, dep, nil); err != nil {
			return fmt.Errorf("failed to install %s: %w", dep.Name, err)
		}
	}

	if progressCallback != nil {
		progressCallback(1.0, "All dependencies installed successfully")
	}

	return nil
}

// ValidateDependencies checks if all required dependencies are installed
func ValidateDependencies() error {
	statuses, err := CheckAllDependencies()
	if err != nil {
		return fmt.Errorf("failed to check dependencies: %w", err)
	}

	var missing []string
	for _, status := range statuses {
		if !status.Installed && status.Dependency.Required {
			missing = append(missing, status.Dependency.Name)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required dependencies: %s. Run 'Install Dependencies' from the Tools tab", strings.Join(missing, ", "))
	}

	return nil
}
