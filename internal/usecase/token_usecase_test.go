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

package usecase

import (
	"math"
	"testing"
	"time"

	"github-copilot-credit-count/internal/domain"
)

// MockTokenRepository implementation.
type MockTokenRepository struct {
	events     []domain.SessionEvent
	workspaces map[string]domain.Workspace
}

func (m *MockTokenRepository) ScanSessions() ([]domain.SessionEvent, map[string]domain.Workspace, error) {
	return m.events, m.workspaces, nil
}

func TestTokenUsecase(t *testing.T) {
	// Setup mock data
	ws1 := domain.Workspace{ID: "ws-1", Path: "c:/projekte/app1", Name: "app1"}
	ws2 := domain.Workspace{ID: "ws-2", Path: "c:/projekte/app2", Name: "app2"}

	timeJune := time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC)
	timeMay := time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)

	mockEvents := []domain.SessionEvent{
		{
			WorkspaceID: "ws-1",
			SessionID:   "sess-1",
			Timestamp:   timeJune,
			Tokens:      domain.TokenUsage{Prompt: 100, Completion: 50, Total: 150, AIC: 0.3, AIU: 300000000},
		},
		{
			WorkspaceID: "ws-2",
			SessionID:   "sess-2",
			Timestamp:   timeJune,
			Tokens:      domain.TokenUsage{Prompt: 200, Completion: 100, Total: 300, AIC: 1.0, AIU: 1000000000},
		},
		{
			WorkspaceID: "ws-1",
			SessionID:   "sess-3",
			Timestamp:   timeMay,
			Tokens:      domain.TokenUsage{Prompt: 50, Completion: 20, Total: 70, AIC: 0.1, AIU: 100000000},
		},
	}

	mockWorkspaces := map[string]domain.Workspace{
		"ws-1": ws1,
		"ws-2": ws2,
	}

	mockRepo := &MockTokenRepository{events: mockEvents, workspaces: mockWorkspaces}
	uc := NewTokenUsecase(mockRepo)

	// Test ScanAndCache
	if err := uc.ScanAndCache(); err != nil {
		t.Fatalf("ScanAndCache failed: %v", err)
	}

	// Test GetAvailableMonths
	months := uc.GetAvailableMonths()
	if len(months) != 2 {
		t.Fatalf("expected 2 months, got %d", len(months))
	}
	if months[0] != "2026-06" || months[1] != "2026-05" {
		t.Errorf("months sorted incorrectly: %v", months)
	}

	// Test GetSummaryForMonth (June 2026)
	summary, err := uc.GetSummaryForMonth("2026-06")
	if err != nil {
		t.Fatalf("GetSummaryForMonth failed: %v", err)
	}

	// June totals: 150+300 = 450 tokens, 0.3+1.0 = 1.3 AIC, 1.3B AIU
	if summary.TotalTokens.Total != 450 {
		t.Errorf("expected 450 total tokens in June, got %d", summary.TotalTokens.Total)
	}
	if math.Abs(summary.TotalTokens.AIC-1.3) > 1e-9 {
		t.Errorf("expected 1.3 AIC in June, got %f", summary.TotalTokens.AIC)
	}
	if math.Abs(summary.TotalTokens.AIU-1300000000) > 1e-9 {
		t.Errorf("expected 1,300,000,000 nano-AIU in June, got %f", summary.TotalTokens.AIU)
	}

	if len(summary.Workspaces) != 2 {
		t.Fatalf("expected 2 workspaces in June summary, got %d", len(summary.Workspaces))
	}

	// Should be sorted by total tokens descending: ws-2 (300) then ws-1 (150)
	if summary.Workspaces[0].Workspace.ID != "ws-2" {
		t.Errorf("expected ws-2 to be first workspace in summary, got %s", summary.Workspaces[0].Workspace.ID)
	}
	if math.Abs(summary.Workspaces[0].Tokens.AIC-1.0) > 1e-9 {
		t.Errorf("expected ws-2 to have 1.0 AIC, got %f", summary.Workspaces[0].Tokens.AIC)
	}
	if summary.Workspaces[1].Workspace.ID != "ws-1" {
		t.Errorf("expected ws-1 to be second workspace in summary, got %s", summary.Workspaces[1].Workspace.ID)
	}
	if math.Abs(summary.Workspaces[1].Tokens.AIC-0.3) > 1e-9 {
		t.Errorf("expected ws-1 to have 0.3 AIC, got %f", summary.Workspaces[1].Tokens.AIC)
	}

	// Test GetOverallTotal
	overall := uc.GetOverallTotal()
	if math.Abs(overall.AIC-1.4) > 1e-9 {
		t.Errorf("expected overall AIC to be 1.4, got %f", overall.AIC)
	}
}
