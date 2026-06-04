package application

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/djalben/istok-agent-core/internal/ports"
)

// mockLLM implements ports.LLMProvider for deterministic testing.
// Simulates real-world LLM behavior including truncation on heavy files.
type mockLLM struct {
	mu             sync.Mutex
	calls          []mockLLMCall // recorded calls for assertions
	callCount      atomic.Int64  // total calls made
	truncateAfter  int           // simulate truncation: cut response after N fully-closed <file> blocks
	maxFilesPerReq int           // track max files requested in a single call (for chunk size assertion)
}

type mockLLMCall struct {
	Model        string
	SystemPrompt string
	UserPrompt   string
	MaxTokens    int
	Reasoning    bool
}

func newMockLLM() *mockLLM {
	return &mockLLM{
		truncateAfter: 0, // 0 = no truncation
	}
}

// withTruncation configures the mock to truncate XML output after N complete files.
// Simulates LLM hitting max_tokens limit mid-response.
func (m *mockLLM) withTruncation(afterNFiles int) *mockLLM {
	m.truncateAfter = afterNFiles
	return m
}

func (m *mockLLM) Complete(ctx context.Context, req ports.LLMRequest) (*ports.LLMResponse, error) {
	// Check context
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	m.callCount.Add(1)

	m.mu.Lock()
	m.calls = append(m.calls, mockLLMCall{
		Model:        req.Model,
		SystemPrompt: req.SystemPrompt,
		UserPrompt:   req.UserPrompt,
		MaxTokens:    req.MaxTokens,
		Reasoning:    req.Reasoning,
	})
	m.mu.Unlock()

	// Route based on prompt content
	lower := strings.ToLower(req.UserPrompt + req.SystemPrompt)

	// Coder responses (check FIRST — coder prompts also contain words like "plan")
	if strings.Contains(lower, "files to generate in this batch") ||
		strings.Contains(lower, "xml artifact") ||
		strings.Contains(lower, "<file path=") {
		files := m.extractRequestedFiles(req.UserPrompt)
		content := m.generateXMLArtifacts(files)
		return &ports.LLMResponse{
			Content:    content,
			TokensUsed: 2000,
			Model:      req.Model,
		}, nil
	}

	// Architect/Planner responses
	if strings.Contains(lower, "architect") || strings.Contains(lower, "manifest") ||
		strings.Contains(lower, "plan") || strings.Contains(lower, "бизнес-аналитик") {
		return &ports.LLMResponse{
			Content:    m.architectResponse(),
			TokensUsed: 500,
			Model:      req.Model,
		}, nil
	}

	// Default: simple text
	return &ports.LLMResponse{
		Content:    "OK",
		TokensUsed: 10,
		Model:      req.Model,
	}, nil
}

// architectResponse returns a plausible architect/planner response.
func (m *mockLLM) architectResponse() string {
	return `{"architecture":"React SPA","components":["App","Router","Pages"],"steps":["Setup project","Create components","Add routing"]}`
}

// extractRequestedFiles parses "FILES TO GENERATE IN THIS BATCH:" from user prompt.
func (m *mockLLM) extractRequestedFiles(userPrompt string) []string {
	marker := "FILES TO GENERATE IN THIS BATCH:"
	idx := strings.Index(userPrompt, marker)
	if idx == -1 {
		return []string{"index.html"}
	}

	rest := userPrompt[idx+len(marker):]
	// Files end at next section (RULES: or empty double-newline)
	endIdx := strings.Index(rest, "\nRULES:")
	if endIdx == -1 {
		endIdx = strings.Index(rest, "\nCRITICAL")
	}
	if endIdx == -1 {
		endIdx = len(rest)
	}
	block := strings.TrimSpace(rest[:endIdx])

	var files []string
	for _, line := range strings.Split(block, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}

	// Track max files per request for chunk size assertion
	m.mu.Lock()
	if len(files) > m.maxFilesPerReq {
		m.maxFilesPerReq = len(files)
	}
	m.mu.Unlock()

	return files
}

// generateXMLArtifacts creates mock XML artifact output.
// If truncateAfter > 0, simulates LLM truncation by cutting mid-file.
func (m *mockLLM) generateXMLArtifacts(files []string) string {
	var sb strings.Builder
	for i, f := range files {
		// If truncation is enabled and we've written enough complete files, truncate mid-file
		if m.truncateAfter > 0 && i >= m.truncateAfter {
			// Write a partial, unclosed file block to simulate truncation
			sb.WriteString(fmt.Sprintf("<file path=\"%s\">\n// This file was truncated by max_tok", f))
			break
		}

		sb.WriteString(fmt.Sprintf("<file path=\"%s\">\n", f))
		sb.WriteString(m.generateMockCode(f))
		sb.WriteString("\n</file>\n\n")
	}
	return sb.String()
}

// generateMockCode returns realistic-looking code for a given file path.
func (m *mockLLM) generateMockCode(filepath string) string {
	lower := strings.ToLower(filepath)
	switch {
	case strings.HasSuffix(lower, ".tsx"):
		name := extractComponentName(filepath)
		return fmt.Sprintf(`import React from 'react';

interface %sProps {
  title?: string;
}

export const %s: React.FC<%sProps> = ({ title = "Default" }) => {
  return (
    <div className="p-4">
      <h2 className="text-xl font-bold">{title}</h2>
      <p>%s component content</p>
    </div>
  );
};

export default %s;`, name, name, name, name, name)

	case strings.HasSuffix(lower, ".ts"):
		return fmt.Sprintf(`// %s
export const config = {
  apiUrl: '/api/v1',
  timeout: 5000,
};`, filepath)

	case strings.HasSuffix(lower, ".css"):
		return `@tailwind base;
@tailwind components;
@tailwind utilities;`

	case strings.HasSuffix(lower, ".json"):
		return `{
  "name": "taxigo",
  "version": "1.0.0"
}`

	default:
		return fmt.Sprintf("// Generated: %s\nexport default {};", filepath)
	}
}

// extractComponentName gets a PascalCase component name from a file path.
func extractComponentName(path string) string {
	parts := strings.Split(path, "/")
	filename := parts[len(parts)-1]
	name := strings.TrimSuffix(filename, ".tsx")
	name = strings.TrimSuffix(name, ".ts")
	if len(name) > 0 {
		return strings.ToUpper(name[:1]) + name[1:]
	}
	return "Component"
}
