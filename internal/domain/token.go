// Copyright (c) 2026 Heino Stömmer.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package domain

import "time"

// Workspace represents a VS Code workspace directory structure.
type Workspace struct {
	ID   string `json:"id"`
	Path string `json:"path"`
	Name string `json:"name"`
}

// TokenUsage holds prompt, completion (output) and total token counts, as well as AI Credits (AIC), AI Usage (AIU), and prompt request counts.
type TokenUsage struct {
	Prompt     int     `json:"prompt"`
	Completion int     `json:"completion"`
	Total      int     `json:"total"`
	AIC        float64 `json:"aic"`      // AI Credits (e.g. 0.30 or 1.00)
	AIU        float64 `json:"aiu"`      // AI Usage in nano-AIU (e.g. 300,000,000)
	Requests   int     `json:"requests"` // Number of individual prompt requests
}

// SessionEvent represents a single token usage event extracted from logs.
type SessionEvent struct {
	WorkspaceID string     `json:"workspaceId"`
	SessionID   string     `json:"sessionId"`
	Timestamp   time.Time  `json:"timestamp"`
	Tokens      TokenUsage `json:"tokens"`
}

// WorkspaceSummary aggregates token usage for a single workspace in a month.
type WorkspaceSummary struct {
	Workspace Workspace  `json:"workspace"`
	Tokens    TokenUsage `json:"tokens"`
}

// MonthSummary represents the aggregated token count for a specific month.
type MonthSummary struct {
	Month       string             `json:"month"` // Format: YYYY-MM
	TotalTokens TokenUsage         `json:"totalTokens"`
	Workspaces  []WorkspaceSummary `json:"workspaces"`
}

// TokenRepository defines the port/interface for retrieving raw copilot session events.
type TokenRepository interface {
	ScanSessions() ([]SessionEvent, map[string]Workspace, error)
}
