package engine

import (
	"fmt"
	"os"

	"skillyfiy/internal/model"
)

// PurgeReport records metrics and outcomes of the atomic purge operation.
type PurgeReport struct {
	SkillsUnlinked   int
	MCPServersPurged int
	ReclaimedTokens  int
	ReclaimedBytes   int64
	Errors           []string
	SkillFiles       []string
	MCPServers       []string
}

// ExecutePurge executes the deletion of selected skills and deregistration of MCP servers.
func ExecutePurge(items []model.AgentItem, dryRun bool) (*PurgeReport, error) {
	report := &PurgeReport{}

	// Separate selected skills and MCP items grouped by manifest source
	var selectedSkills []model.AgentItem
	mcpByManifest := make(map[string][]model.AgentItem)

	for _, item := range items {
		if !item.Selected {
			continue
		}

		switch item.Type {
		case model.TypeSkill:
			selectedSkills = append(selectedSkills, item)
		case model.TypeMCP:
			mcpByManifest[item.SourcePath] = append(mcpByManifest[item.SourcePath], item)
		}
	}

	// 1. Process Skills
	for _, skill := range selectedSkills {
		report.SkillFiles = append(report.SkillFiles, skill.SourcePath)
		report.ReclaimedTokens += skill.Tokens
		report.ReclaimedBytes += skill.Size

		if dryRun {
			report.SkillsUnlinked++
			continue
		}

		err := os.Remove(skill.SourcePath)
		if err != nil {
			if os.IsNotExist(err) {
				report.SkillsUnlinked++
			} else {
				report.Errors = append(report.Errors, fmt.Sprintf("Failed to remove %s: %v", skill.SourcePath, err))
			}
		} else {
			report.SkillsUnlinked++
		}
	}

	// 2. Process MCP Servers
	for manifestPath, serverItems := range mcpByManifest {
		var serverNames []string
		for _, s := range serverItems {
			serverNames = append(serverNames, s.Name)
			report.MCPServers = append(report.MCPServers, s.Name)
			report.ReclaimedTokens += s.Tokens
			report.ReclaimedBytes += s.Size
		}

		if dryRun {
			report.MCPServersPurged += len(serverNames)
			continue
		}

		if err := MutateAndSaveMCP(manifestPath, serverNames); err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("Failed to update MCP manifest %s: %v", manifestPath, err))
		} else {
			report.MCPServersPurged += len(serverNames)
		}
	}

	return report, nil
}
