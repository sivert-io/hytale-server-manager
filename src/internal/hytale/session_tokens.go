package hytale

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SessionTokens holds session and identity tokens for server authentication
// Per Server Provider Authentication Guide: https://support.hytale.com/hc/en-us/articles/45328341414043
type SessionTokens struct {
	SessionToken  string    `json:"session_token"`
	IdentityToken string    `json:"identity_token"`
	OwnerUUID     string    `json:"owner_uuid"` // Profile UUID
	ExpiresAt     time.Time `json:"expires_at"` // When tokens expire (1 hour TTL)
}

// GetSessionTokensPath returns the path to the session tokens file
func GetSessionTokensPath() string {
	return filepath.Join(GetSharedConfigDir(), ".session-tokens.json")
}

// LoadSessionTokens loads session tokens from the shared config directory
func LoadSessionTokens() (*SessionTokens, error) {
	tokensPath := GetSessionTokensPath()
	
	data, err := os.ReadFile(tokensPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("session tokens not found - need to authenticate first")
		}
		return nil, fmt.Errorf("failed to read session tokens: %w", err)
	}

	var tokens SessionTokens
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, fmt.Errorf("failed to parse session tokens: %w", err)
	}

	// Check if tokens are expired
	if time.Now().After(tokens.ExpiresAt) {
		return nil, fmt.Errorf("session tokens expired - need to refresh")
	}

	return &tokens, nil
}

// SaveSessionTokens saves session tokens to the shared config directory
func SaveSessionTokens(tokens *SessionTokens) error {
	tokensPath := GetSessionTokensPath()
	
	// Ensure directory exists
	dir := filepath.Dir(tokensPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create tokens directory: %w", err)
	}

	data, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session tokens: %w", err)
	}

	// Write with restrictive permissions (owner read/write only)
	if err := os.WriteFile(tokensPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write session tokens: %w", err)
	}

	return nil
}

// CreateGameSession creates a new game session and returns session/identity tokens
// This should be called after OAuth authentication to get tokens for server startup
// TODO: Implement actual API call to Hytale's game session endpoint
// For now, this is a placeholder that indicates tokens need to be obtained
func CreateGameSession(ctx context.Context, accessToken, profileUUID string) (*SessionTokens, error) {
	// TODO: Implement actual API call:
	// POST /game-session/new
	// Headers: Authorization: Bearer <accessToken>
	// Body: {"uuid": "<profileUUID>"}
	// Response: {"sessionToken": "...", "identityToken": "..."}
	
	// Placeholder implementation
	// In production, this would make an HTTP request to Hytale's API
	return nil, fmt.Errorf("game session creation not yet implemented - requires Hytale API integration")
}

// GetOrCreateSessionTokens gets existing valid tokens or creates new ones
// Returns tokens and whether they were newly created
func GetOrCreateSessionTokens(ctx context.Context, accessToken, profileUUID string) (*SessionTokens, bool, error) {
	// Try to load existing tokens
	tokens, err := LoadSessionTokens()
	if err == nil {
		// Tokens exist and are valid
		return tokens, false, nil
	}

	// Tokens don't exist or expired - create new ones
	newTokens, err := CreateGameSession(ctx, accessToken, profileUUID)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create game session: %w", err)
	}

	// Set expiry (1 hour from now, minus 5 minutes buffer for refresh)
	newTokens.ExpiresAt = time.Now().Add(55 * time.Minute)
	
	// Save tokens
	if err := SaveSessionTokens(newTokens); err != nil {
		return nil, false, fmt.Errorf("failed to save session tokens: %w", err)
	}

	return newTokens, true, nil
}

// AreSessionTokensValid checks if session tokens exist and are not expired
func AreSessionTokensValid() bool {
	tokens, err := LoadSessionTokens()
	return err == nil && tokens != nil && time.Now().Before(tokens.ExpiresAt)
}
