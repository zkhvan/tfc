package init_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	initCmd "github.com/zkhvan/tfc/cmd/tfc/init"
	"github.com/zkhvan/tfc/internal/test"
	"github.com/zkhvan/tfc/pkg/cmdutil"
	"github.com/zkhvan/tfc/pkg/iolib"
	"github.com/zkhvan/tfc/pkg/tfconfig"
)

func TestInit_generates_state_tf(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	f := &cmdutil.Factory{
		IOStreams: &iolib.IOStreams{
			Out:    stdout,
			ErrOut: stderr,
		},
		TFEClient:       nil,
		TerraformConfig: func() *tfconfig.TerraformConfig { return nil },
	}

	cmd := initCmd.NewCmdInit(f)
	cmd.SetArgs([]string{"-W", "my-org/my-workspace"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dir, "state.tf"))
	if err != nil {
		t.Fatalf("reading state.tf: %v", err)
	}

	want := `terraform {
  cloud {
    organization = "my-org"

    workspaces {
      name    = "my-workspace"
    }
  }
}
`
	if string(got) != want {
		t.Errorf("state.tf content got:\n%s\nwant:\n%s", string(got), want)
	}

	test.Buffer(t, stderr, "Wrote state.tf\n")
	test.BufferEmpty(t, stdout)
}

func TestInit_generates_state_tf_with_project(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	f := &cmdutil.Factory{
		IOStreams: &iolib.IOStreams{
			Out:    stdout,
			ErrOut: stderr,
		},
		TFEClient:       nil,
		TerraformConfig: func() *tfconfig.TerraformConfig { return nil },
	}

	cmd := initCmd.NewCmdInit(f)
	cmd.SetArgs([]string{"-W", "my-org/my-workspace", "--project", "my-project"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dir, "state.tf"))
	if err != nil {
		t.Fatalf("reading state.tf: %v", err)
	}

	want := `terraform {
  cloud {
    organization = "my-org"

    workspaces {
      name    = "my-workspace"
      project = "my-project"
    }
  }
}
`
	if string(got) != want {
		t.Errorf("state.tf content got:\n%s\nwant:\n%s", string(got), want)
	}
}

func TestInit_errors_when_state_tf_exists(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	// Create existing state.tf
	err := os.WriteFile(filepath.Join(dir, "state.tf"), []byte("existing"), 0600)
	if err != nil {
		t.Fatalf("creating state.tf: %v", err)
	}

	f := &cmdutil.Factory{
		IOStreams: &iolib.IOStreams{
			Out:    new(bytes.Buffer),
			ErrOut: new(bytes.Buffer),
		},
		TFEClient:       nil,
		TerraformConfig: func() *tfconfig.TerraformConfig { return nil },
	}

	cmd := initCmd.NewCmdInit(f)
	cmd.SetArgs([]string{"-W", "my-org/my-workspace"})
	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "state.tf already exists; use --force to overwrite" {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify original content is preserved
	got, _ := os.ReadFile(filepath.Join(dir, "state.tf"))
	if string(got) != "existing" {
		t.Errorf("state.tf was modified, got: %s", string(got))
	}
}

func TestInit_force_overwrites_existing(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	// Create existing state.tf
	err := os.WriteFile(filepath.Join(dir, "state.tf"), []byte("existing"), 0600)
	if err != nil {
		t.Fatalf("creating state.tf: %v", err)
	}

	f := &cmdutil.Factory{
		IOStreams: &iolib.IOStreams{
			Out:    new(bytes.Buffer),
			ErrOut: new(bytes.Buffer),
		},
		TFEClient:       nil,
		TerraformConfig: func() *tfconfig.TerraformConfig { return nil },
	}

	cmd := initCmd.NewCmdInit(f)
	cmd.SetArgs([]string{"-W", "my-org/my-workspace", "--force"})
	err = cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := os.ReadFile(filepath.Join(dir, "state.tf"))
	if string(got) == "existing" {
		t.Error("state.tf was not overwritten")
	}
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })
}
