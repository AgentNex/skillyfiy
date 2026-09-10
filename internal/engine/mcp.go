package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"skillyfiy/internal/model"
)

// ServerConfig captures the common MCP server configuration schema.
type ServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// ParseMCP reads an MCP configuration JSON manifest and extracts AgentItem representations.
func ParseMCP(manifestPath string) ([]model.AgentItem, error) {
	var items []model.AgentItem

	info, err := os.Stat(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return items, nil
		}
		return nil, fmt.Errorf("failed to stat MCP manifest: %w", err)
	}
	if info.IsDir() {
		// Look for common manifest files in directory
		candidates := []string{
			filepath.Join(manifestPath, "mcp_config.json"),
			filepath.Join(manifestPath, "settings.json"),
			filepath.Join(manifestPath, "config.json"),
			filepath.Join(manifestPath, ".mcp.json"),
		}
		for _, cand := range candidates {
			if candInfo, err := os.Stat(cand); err == nil && !candInfo.IsDir() {
				return ParseMCP(cand)
			}
		}
		return nil, fmt.Errorf("no valid MCP manifest (mcp_config.json, settings.json) found in directory: %s", manifestPath)
	}

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read MCP manifest: %w", err)
	}

	parsed := gjson.ParseBytes(data)
	mcpServers := parsed.Get("mcpServers")

	// If mcpServers key is not found, check if root is directly a map of servers
	if !mcpServers.Exists() || !mcpServers.IsObject() {
		altServers := parsed.Get("servers")
		if altServers.Exists() && altServers.IsObject() {
			mcpServers = altServers
		} else {
			return items, nil
		}
	}

	absPath, err := filepath.Abs(manifestPath)
	if err != nil {
		absPath = manifestPath
	}

	mcpServers.ForEach(func(key, value gjson.Result) bool {
		serverName := key.String()
		rawServerJSON := value.Raw

		var prettyBuf bytes.Buffer
		if err := json.Indent(&prettyBuf, []byte(rawServerJSON), "", "  "); err == nil {
			rawServerJSON = prettyBuf.String()
		}

		size := int64(len(rawServerJSON))
		tokens := int(size / 4)
		if size > 0 && tokens == 0 {
			tokens = 1
		}

		command := value.Get("command").String()
		serverURL := value.Get("serverUrl").String()
		if serverURL == "" {
			serverURL = value.Get("url").String()
		}

		var args []string
		for _, arg := range value.Get("args").Array() {
			args = append(args, arg.String())
		}
		envCount := len(value.Get("env").Map())

		descParts := []string{}
		if command != "" {
			descParts = append(descParts, command)
		}
		if len(args) > 0 {
			descParts = append(descParts, strings.Join(args, " "))
		}
		if len(descParts) == 0 && serverURL != "" {
			descParts = append(descParts, serverURL)
		}
		description := strings.Join(descParts, " ")
		if description == "" {
			description = fmt.Sprintf("MCP Server [%s]", serverName)
		}
		if len(description) > 160 {
			description = description[:157] + "..."
		}

		item := model.AgentItem{
			ID:          "mcp:" + serverName,
			Name:        serverName,
			Type:        model.TypeMCP,
			SourcePath:  absPath,
			Description: description,
			Tokens:      tokens,
			Size:        size,
			Details: map[string]string{
				"Category":  "MCP Server",
				"Server":    serverName,
				"Command":   command,
				"Arguments": strings.Join(args, " "),
				"Env Keys":  fmt.Sprintf("%d defined", envCount),
				"Manifest":  absPath,
				"Size":      model.FormatBytes(size),
				"Tokens":    fmt.Sprintf("~%d", tokens),
			},
			RawPreview: rawServerJSON,
			Selected:   false,
		}

		items = append(items, item)
		return true
	})

	return items, nil
}

// BackupManifest creates a <path>.bak copy of the manifest file.
func BackupManifest(manifestPath string) error {
	bakPath := manifestPath + ".bak"

	src, err := os.Open(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to open source manifest for backup: %w", err)
	}
	defer src.Close()

	dst, err := os.OpenFile(bakPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open backup destination: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to copy manifest to backup: %w", err)
	}

	return dst.Sync()
}

// MutateAndSaveMCP removes the specified server keys from manifestPath using sjson,
// formats with 2-space indentation, and writes atomically.
func MutateAndSaveMCP(manifestPath string, serverNames []string) error {
	if len(serverNames) == 0 {
		return nil
	}

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest for mutation: %w", err)
	}

	// Backup original before mutation
	if err := BackupManifest(manifestPath); err != nil {
		return fmt.Errorf("backup failed prior to mutation: %w", err)
	}

	// Determine prefix: "mcpServers" or "servers"
	parsed := gjson.ParseBytes(data)
	prefix := "mcpServers"
	if !parsed.Get("mcpServers").Exists() && parsed.Get("servers").Exists() {
		prefix = "servers"
	}

	mutated := data
	for _, server := range serverNames {
		keyPath := prefix + "." + server
		updated, err := sjson.DeleteBytes(mutated, keyPath)
		if err != nil {
			return fmt.Errorf("failed to delete server key '%s': %w", server, err)
		}
		mutated = updated
	}

	// Format with 2-space indentation
	var indented bytes.Buffer
	if err := json.Indent(&indented, mutated, "", "  "); err == nil {
		mutated = append(indented.Bytes(), '\n')
	}

	// Atomic write via temporary file
	dir := filepath.Dir(manifestPath)
	tmpFile, err := os.CreateTemp(dir, "skillyfiy-mcp-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file for atomic write: %w", err)
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName)

	if _, err := tmpFile.Write(mutated); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write mutated manifest to temp file: %w", err)
	}
	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to sync temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := os.Chmod(tmpName, 0644); err != nil {
		return fmt.Errorf("failed to set permissions on temp file: %w", err)
	}

	if err := os.Rename(tmpName, manifestPath); err != nil {
		return fmt.Errorf("failed to replace manifest file atomically: %w", err)
	}

	return nil
}
