package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WorkspaceManager handles file operations within the workspace
type WorkspaceManager struct {
	workspaceDir string
}

// NewWorkspaceManager creates a new workspace manager
func NewWorkspaceManager() (*WorkspaceManager, error) {
	workspaceDir := "workspace"
	
	if err := os.MkdirAll(workspaceDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create workspace directory: %w", err)
	}
	
	return &WorkspaceManager{
		workspaceDir: workspaceDir,
	}, nil
}

// WriteFile writes content to a markdown file in the workspace
func (w *WorkspaceManager) WriteFile(filename, content string) error {
	if !strings.HasSuffix(filename, ".md") {
		return fmt.Errorf("only .md files are allowed")
	}
	
	// Ensure filename doesn't contain path traversal
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		return fmt.Errorf("invalid filename: %s", filename)
	}
	
	filePath := filepath.Join(w.workspaceDir, filename)
	return os.WriteFile(filePath, []byte(content), 0644)
}

// ReadFile reads content from a markdown file in the workspace
func (w *WorkspaceManager) ReadFile(filename string) (string, error) {
	if !strings.HasSuffix(filename, ".md") {
		return "", fmt.Errorf("only .md files are allowed")
	}
	
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		return "", fmt.Errorf("invalid filename: %s", filename)
	}
	
	filePath := filepath.Join(w.workspaceDir, filename)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	
	return string(content), nil
}

// GetWorkspaceDir returns the workspace directory path
func (w *WorkspaceManager) GetWorkspaceDir() string {
	return w.workspaceDir
}