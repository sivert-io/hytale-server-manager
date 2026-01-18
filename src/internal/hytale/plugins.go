package hytale

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	// PerformanceSaverPluginRepo is the GitHub repository for the Nitrado Performance Saver plugin
	PerformanceSaverPluginRepo = "nitrado/hytale-plugin-performance-saver"
	// PerformanceSaverPluginName is the plugin directory name
	PerformanceSaverPluginName = "Nitrado_PerformanceSaver"
)

// PerformanceSaverConfig represents the plugin's configuration
type PerformanceSaverConfig struct {
	Tps struct {
		Enabled              bool     `json:"Enabled"`
		TpsLimit             int      `json:"TpsLimit"`
		TpsLimitEmpty        int      `json:"TpsLimitEmpty"`
		OnlyWorlds           []string `json:"OnlyWorlds"`
		InitialDelaySeconds  int      `json:"InitialDelaySeconds"`
		CheckIntervalSeconds int      `json:"CheckIntervalSeconds"`
		EmptyLimitDelaySeconds int    `json:"EmptyLimitDelaySeconds"`
	} `json:"Tps"`
	ViewRadius struct {
		Enabled                 bool   `json:"Enabled"`
		MinViewRadius           int    `json:"MinViewRadius"`
		DecreaseFactor          float64 `json:"DecreaseFactor"`
		IncreaseValue           int    `json:"IncreaseValue"`
		InitialDelaySeconds     int    `json:"InitialDelaySeconds"`
		CheckIntervalSeconds    int    `json:"CheckIntervalSeconds"`
		RecoveryWaitTimeSeconds int    `json:"RecoveryWaitTimeSeconds"`
		RequireNotifyPermission bool   `json:"RequireNotifyPermission"`
		GcMonitor               struct {
			Enabled             bool    `json:"Enabled"`
			HeapThresholdRatio float64 `json:"HeapThresholdRatio"`
			TriggerSequenceLength int  `json:"TriggerSequenceLength"`
			WindowSeconds       int    `json:"WindowSeconds"`
		} `json:"GcMonitor"`
		TpsMonitor struct {
			Enabled                bool     `json:"Enabled"`
			TpsWaterMarkHigh       float64  `json:"TpsWaterMarkHigh"`
			TpsWaterMarkLow        float64  `json:"TpsWaterMarkLow"`
			OnlyWorlds             []string `json:"OnlyWorlds"`
			AdjustmentDelaySeconds int      `json:"AdjustmentDelaySeconds"`
		} `json:"TpsMonitor"`
	} `json:"ViewRadius"`
	ChunkGarbageCollection struct {
		Enabled                       bool `json:"Enabled"`
		MinChunkCount                 int  `json:"MinChunkCount"`
		ChunkDropRatioThreshold       float64 `json:"ChunkDropRatioThreshold"`
		GarbageCollectionDelaySeconds int  `json:"GarbageCollectionDelaySeconds"`
		InitialDelaySeconds           int  `json:"InitialDelaySeconds"`
		CheckIntervalSeconds          int  `json:"CheckIntervalSeconds"`
	} `json:"ChunkGarbageCollection"`
}

// CreateDefaultPerformanceSaverConfig creates the default configuration for the Performance Saver plugin
func CreateDefaultPerformanceSaverConfig() *PerformanceSaverConfig {
	config := &PerformanceSaverConfig{}
	
	// TPS settings
	config.Tps.Enabled = true
	config.Tps.TpsLimit = 20
	config.Tps.TpsLimitEmpty = 5
	config.Tps.OnlyWorlds = []string{}
	config.Tps.InitialDelaySeconds = 30
	config.Tps.CheckIntervalSeconds = 5
	config.Tps.EmptyLimitDelaySeconds = 300
	
	// View Radius settings
	config.ViewRadius.Enabled = true
	config.ViewRadius.MinViewRadius = 2
	config.ViewRadius.DecreaseFactor = 0.75
	config.ViewRadius.IncreaseValue = 1
	config.ViewRadius.InitialDelaySeconds = 30
	config.ViewRadius.CheckIntervalSeconds = 5
	config.ViewRadius.RecoveryWaitTimeSeconds = 60
	config.ViewRadius.RequireNotifyPermission = false
	
	// GC Monitor
	config.ViewRadius.GcMonitor.Enabled = true
	config.ViewRadius.GcMonitor.HeapThresholdRatio = 0.85
	config.ViewRadius.GcMonitor.TriggerSequenceLength = 3
	config.ViewRadius.GcMonitor.WindowSeconds = 60
	
	// TPS Monitor
	config.ViewRadius.TpsMonitor.Enabled = true
	config.ViewRadius.TpsMonitor.TpsWaterMarkHigh = 0.75
	config.ViewRadius.TpsMonitor.TpsWaterMarkLow = 0.6
	config.ViewRadius.TpsMonitor.OnlyWorlds = []string{}
	config.ViewRadius.TpsMonitor.AdjustmentDelaySeconds = 20
	
	// Chunk Garbage Collection
	config.ChunkGarbageCollection.Enabled = true
	config.ChunkGarbageCollection.MinChunkCount = 128
	config.ChunkGarbageCollection.ChunkDropRatioThreshold = 0.8
	config.ChunkGarbageCollection.GarbageCollectionDelaySeconds = 300
	config.ChunkGarbageCollection.InitialDelaySeconds = 5
	config.ChunkGarbageCollection.CheckIntervalSeconds = 5
	
	return config
}


// DownloadPerformanceSaverPlugin downloads the Performance Saver plugin from GitHub releases
// If progressCallback is provided, it will be called with progress updates
func DownloadPerformanceSaverPlugin(ctx context.Context, progressCallback ProgressCallback) error {
	// GitHub releases API to get latest release
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", PerformanceSaverPluginRepo)
	
	if progressCallback != nil {
		progressCallback(0.0, "Fetching release information...")
	}
	
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch release info: status %d", resp.StatusCode)
	}
	
	var release struct {
		Assets []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("failed to parse release info: %w", err)
	}
	
	// Find the JAR file
	var jarURL string
	for _, asset := range release.Assets {
		if filepath.Ext(asset.Name) == ".jar" {
			jarURL = asset.BrowserDownloadURL
			break
		}
	}
	
	if jarURL == "" {
		return fmt.Errorf("no JAR file found in latest release")
	}
	
	// Create plugin directory
	sharedDir := GetSharedConfigDir()
	pluginDir := filepath.Join(sharedDir, "mods", PerformanceSaverPluginName)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugin directory: %w", err)
	}
	
	// Download using wget or curl with progress
	jarPath := filepath.Join(pluginDir, "PerformanceSaver.jar")
	
	// Try wget first, then curl, then fallback to http.Get
	if wgetPath, err := exec.LookPath("wget"); err == nil {
		return downloadWithWget(ctx, wgetPath, jarURL, jarPath, progressCallback)
	}
	
	if curlPath, err := exec.LookPath("curl"); err == nil {
		return downloadWithCurl(ctx, curlPath, jarURL, jarPath, progressCallback)
	}
	
	// Fallback to http.Get with manual progress tracking
	return downloadWithHTTP(ctx, jarURL, jarPath, progressCallback)
}

// downloadWithWget downloads using wget with progress bar
func downloadWithWget(ctx context.Context, wgetPath, url, destPath string, progressCallback ProgressCallback) error {
	if progressCallback != nil {
		progressCallback(0.0, "Downloading plugin...")
	}
	
	cmd := exec.CommandContext(ctx, wgetPath,
		"--progress=bar:force",
		"--output-document="+destPath,
		url,
	)
	
	// wget sends progress to stderr
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start wget: %w", err)
	}
	
	// Parse progress from stderr
	scanner := bufio.NewScanner(stderr)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			if progressCallback != nil {
				// Parse wget progress: "100%[=========>] 1,234,567  1.23M/s  in 5s"
				if idx := strings.Index(line, "%"); idx > 0 {
					var percent float64
					if _, err := fmt.Sscanf(line[:idx+1], "%f%%", &percent); err == nil {
						if percent >= 0 && percent <= 100 {
							progressCallback(percent/100.0, "Downloading plugin...")
						}
					}
				}
			}
		}
	}()
	
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("wget failed: %w", err)
	}
	
	if progressCallback != nil {
		progressCallback(1.0, "Download complete")
	}
	
	return nil
}

// downloadWithCurl downloads using curl with progress bar
func downloadWithCurl(ctx context.Context, curlPath, url, destPath string, progressCallback ProgressCallback) error {
	if progressCallback != nil {
		progressCallback(0.0, "Downloading plugin...")
	}
	
	file, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()
	
	cmd := exec.CommandContext(ctx, curlPath,
		"--progress-bar",
		"--output", destPath,
		url,
	)
	
	// curl sends progress to stderr
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start curl: %w", err)
	}
	
	// Parse progress from stderr
	scanner := bufio.NewScanner(stderr)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			if progressCallback != nil {
				// Parse curl progress: "########## 100.0%"
				if idx := strings.LastIndex(line, "%"); idx > 0 {
					var percent float64
					if _, err := fmt.Sscanf(line[idx-5:idx+1], "%f%%", &percent); err == nil {
						if percent >= 0 && percent <= 100 {
							progressCallback(percent/100.0, "Downloading plugin...")
						}
					}
				}
			}
		}
	}()
	
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("curl failed: %w", err)
	}
	
	if progressCallback != nil {
		progressCallback(1.0, "Download complete")
	}
	
	return nil
}

// downloadWithHTTP downloads using http.Get with manual progress tracking
func downloadWithHTTP(ctx context.Context, url, destPath string, progressCallback ProgressCallback) error {
	if progressCallback != nil {
		progressCallback(0.0, "Downloading plugin...")
	}
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download: status %d", resp.StatusCode)
	}
	
	file, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()
	
	// Track progress if content length is available
	total := resp.ContentLength
	var downloaded int64
	
	buf := make([]byte, 32*1024) // 32KB buffer
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := file.Write(buf[:n]); writeErr != nil {
				return fmt.Errorf("failed to write file: %w", writeErr)
			}
			downloaded += int64(n)
			
			if progressCallback != nil && total > 0 {
				percent := float64(downloaded) / float64(total)
				progressCallback(percent, "Downloading plugin...")
			}
		}
		
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}
	}
	
	if progressCallback != nil {
		progressCallback(1.0, "Download complete")
	}
	
	return nil
}

// InstallPerformanceSaverPlugin installs the Performance Saver plugin with default config
// If progressCallback is provided, it will be called with progress updates
func InstallPerformanceSaverPlugin(ctx context.Context, progressCallback ProgressCallback) error {
	sharedDir := GetSharedConfigDir()
	pluginDir := filepath.Join(sharedDir, "mods", PerformanceSaverPluginName)
	
	// Create plugin directory
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugin directory: %w", err)
	}
	
	// Download plugin JAR
	if err := DownloadPerformanceSaverPlugin(ctx, progressCallback); err != nil {
		return fmt.Errorf("failed to download plugin: %w", err)
	}
	
	if progressCallback != nil {
		progressCallback(0.9, "Creating plugin configuration...")
	}
	
	// Create default config
	config := CreateDefaultPerformanceSaverConfig()
	configPath := filepath.Join(pluginDir, "config.json")
	
	configFile, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer configFile.Close()
	
	encoder := json.NewEncoder(configFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}
