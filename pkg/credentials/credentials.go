package credentials

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// TerraformCredentials represents the structure of credentials.tfrc.json
type TerraformCredentials struct {
	Credentials map[string]struct {
		Token string `json:"token"`
	} `json:"credentials"`
}

// GetTerraformCredentialsPath returns the path to the Terraform credentials file
func GetTerraformCredentialsPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting user home directory: %w", err)
	}

	// The path is different on Windows
	if runtime.GOOS == "windows" {
		return filepath.Join(homeDir, "AppData", "Roaming", "terraform.d", "credentials.tfrc.json"), nil
	}

	return filepath.Join(homeDir, ".terraform.d", "credentials.tfrc.json"), nil
}

// LoadTerraformCredentials loads and parses the Terraform credentials file
func LoadTerraformCredentials() (*TerraformCredentials, error) {
	path, err := GetTerraformCredentialsPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty credentials if file doesn't exist
			return &TerraformCredentials{
				Credentials: make(map[string]struct {
					Token string `json:"token"`
				}),
			}, nil
		}
		return nil, fmt.Errorf("error reading credentials file: %w", err)
	}

	var creds TerraformCredentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("error parsing credentials file: %w", err)
	}

	return &creds, nil
}

// GetTokenForHost returns the token for a specific host from the credentials file
func GetTokenForHost(hostname string) (string, error) {
	creds, err := LoadTerraformCredentials()
	if err != nil {
		return "", err
	}

	if hostCreds, ok := creds.Credentials[hostname]; ok {
		return hostCreds.Token, nil
	}

	return "", nil
}
