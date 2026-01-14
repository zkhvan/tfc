package tfconfig

import (
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// TerraformConfig represents the parsed terraform.cloud block from state.tf
type TerraformConfig struct {
	Organization string
	Workspace    WorkspaceConfig
}

// WorkspaceConfig represents the workspace configuration
type WorkspaceConfig struct {
	Name    string
	Project string // Optional
}

// IsValid returns true if both organization and workspace name are present
func (c *TerraformConfig) IsValid() bool {
	return c != nil && c.Organization != "" && c.Workspace.Name != ""
}

// ReadConfig attempts to read and parse state.tf from the given directory.
// Returns nil if file doesn't exist or cannot be parsed (silent fallback).
func ReadConfig(dir string) *TerraformConfig {
	filePath := filepath.Join(dir, "state.tf")

	// Silent fallback: file doesn't exist
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}

	// Silent fallback: parse error
	file, diags := hclsyntax.ParseConfig(content, filePath, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil
	}

	// Parse the terraform.cloud block
	cfg, err := decodeConfig(file.Body)
	if err != nil {
		return nil
	}

	// Silent fallback: invalid/incomplete config
	if !cfg.IsValid() {
		return nil
	}

	return cfg
}

// decodeConfig extracts the terraform.cloud block from the HCL body
func decodeConfig(body hcl.Body) (*TerraformConfig, error) {
	// Define the structure we're looking for in HCL
	type terraformBlock struct {
		Cloud *struct {
			Organization string `hcl:"organization"`
			Workspaces   *struct {
				Name    string `hcl:"name,optional"`
				Project string `hcl:"project,optional"`
			} `hcl:"workspaces,block"`
		} `hcl:"cloud,block"`
	}

	type rootSchema struct {
		Terraform *terraformBlock `hcl:"terraform,block"`
	}

	var root rootSchema
	diags := gohcl.DecodeBody(body, nil, &root)
	if diags.HasErrors() {
		return nil, diags
	}

	// Extract the data we need
	if root.Terraform == nil || root.Terraform.Cloud == nil {
		return nil, nil
	}

	cfg := &TerraformConfig{
		Organization: root.Terraform.Cloud.Organization,
	}

	if root.Terraform.Cloud.Workspaces != nil {
		cfg.Workspace = WorkspaceConfig{
			Name:    root.Terraform.Cloud.Workspaces.Name,
			Project: root.Terraform.Cloud.Workspaces.Project,
		}
	}

	return cfg, nil
}
