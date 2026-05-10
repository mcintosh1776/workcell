package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitCmdFakeRuntimeHappyPath(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "workcell.yaml")
	
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	defer os.Chdir(origDir)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}

	if err := initCmd([]string{"--runtime", "fake"}); err != nil {
		t.Fatalf("initCmd failed: %v", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	expected := "backend: fake"
	if !contains(string(content), expected) {
		t.Errorf("config missing %q, got: %q", expected, string(content))
	}
}

func TestInitCmdExistingConfigPathCollision(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "workcell.yaml")
	
	if err := os.WriteFile(configPath, []byte("existing"), 0o644); err != nil {
		t.Fatalf("failed to create existing config: %v", err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	defer os.Chdir(origDir)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}

	err = initCmd([]string{"--runtime", "fake"})
	if err == nil {
		t.Fatal("expected error for existing config, got nil")
	}
	if !contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got: %v", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	if string(content) != "existing" {
		t.Errorf("config was overwritten, got: %q", string(content))
	}
}

func TestInitCmdUnsupportedRuntime(t *testing.T) {
	tmpDir := t.TempDir()
	
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	defer os.Chdir(origDir)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}

	err = initCmd([]string{"--runtime", "docker"})
	if err == nil {
		t.Fatal("expected error for unsupported runtime, got nil")
	}
	if !contains(err.Error(), "unsupported runtime") {
		t.Errorf("expected 'unsupported runtime' error, got: %v", err)
	}

	configPath := filepath.Join(tmpDir, "workcell.yaml")
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Error("config file should not exist for unsupported runtime")
	}
}

func TestInitCmdPodmanRuntime(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "workcell.yaml")
	
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	defer os.Chdir(origDir)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}

	if err := initCmd([]string{"--runtime", "podman"}); err != nil {
		t.Fatalf("initCmd failed: %v", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	expected := "backend: podman"
	if !contains(string(content), expected) {
		t.Errorf("config missing %q, got: %q", expected, string(content))
	}
}

func TestInitCmdIncusRuntime(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "workcell.yaml")
	
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	defer os.Chdir(origDir)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}

	if err := initCmd([]string{"--runtime", "incus"}); err != nil {
		t.Fatalf("initCmd failed: %v", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	expected := "backend: incus"
	if !contains(string(content), expected) {
		t.Errorf("config missing %q, got: %q", expected, string(content))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsInternal(s, substr))
}

func containsInternal(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}