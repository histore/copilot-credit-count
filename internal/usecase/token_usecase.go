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
	"fmt"
	"sort"
	"sync"

	"github-copilot-credit-count/internal/domain"
)

// TokenUsecase handles the core business logic of token count aggregation.
type TokenUsecase struct {
	repo       domain.TokenRepository
	mu         sync.RWMutex
	events     []domain.SessionEvent
	workspaces map[string]domain.Workspace
	isScanned  bool
}

// NewTokenUsecase creates a new usecase instance.
func NewTokenUsecase(repo domain.TokenRepository) *TokenUsecase {
	return &TokenUsecase{
		repo:       repo,
		workspaces: make(map[string]domain.Workspace),
	}
}

// ScanAndCache triggers a repository scan and caches the results in memory.
func (u *TokenUsecase) ScanAndCache() error {
	u.mu.Lock()
	defer u.mu.Unlock()

	events, workspaces, err := u.repo.ScanSessions()
	if err != nil {
		return fmt.Errorf("failed scanning sessions: %w", err)
	}

	u.events = events
	u.workspaces = workspaces
	u.isScanned = true
	return nil
}

// GetAvailableMonths returns a sorted list of months in YYYY-MM format (latest first).
func (u *TokenUsecase) GetAvailableMonths() []string {
	u.mu.RLock()
	defer u.mu.RUnlock()

	monthMap := make(map[string]bool)
	for _, event := range u.events {
		if event.Timestamp.IsZero() {
			continue
		}
		monthStr := event.Timestamp.Format("2006-01")
		monthMap[monthStr] = true
	}

	var months []string
	for month := range monthMap {
		months = append(months, month)
	}

	// Sort months descending (latest month first)
	sort.Slice(months, func(i, j int) bool {
		return months[i] > months[j]
	})

	return months
}

// GetSummaryForMonth compiles token metrics for a specific month (YYYY-MM).
func (u *TokenUsecase) GetSummaryForMonth(month string) (domain.MonthSummary, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	if !u.isScanned {
		return domain.MonthSummary{}, fmt.Errorf("scan has not been executed yet")
	}

	summary := domain.MonthSummary{
		Month: month,
	}

	// Map workspace ID to aggregated tokens for this month
	wsTokenMap := make(map[string]domain.TokenUsage)

	for _, event := range u.events {
		if event.Timestamp.IsZero() {
			continue
		}
		eventMonth := event.Timestamp.Format("2006-01")
		if eventMonth != month {
			continue
		}

		// Accumulate total month tokens and credits
		summary.TotalTokens.Prompt += event.Tokens.Prompt
		summary.TotalTokens.Completion += event.Tokens.Completion
		summary.TotalTokens.Total += event.Tokens.Total
		summary.TotalTokens.AIC += event.Tokens.AIC
		summary.TotalTokens.AIU += event.Tokens.AIU
		summary.TotalTokens.Requests += event.Tokens.Requests

		// Accumulate workspace tokens and credits
		wsID := event.WorkspaceID
		wsUsage := wsTokenMap[wsID]
		wsUsage.Prompt += event.Tokens.Prompt
		wsUsage.Completion += event.Tokens.Completion
		wsUsage.Total += event.Tokens.Total
		wsUsage.AIC += event.Tokens.AIC
		wsUsage.AIU += event.Tokens.AIU
		wsUsage.Requests += event.Tokens.Requests
		wsTokenMap[wsID] = wsUsage
	}

	// Build the workspace summary list
	for wsID, usage := range wsTokenMap {
		wsInfo, exists := u.workspaces[wsID]
		if !exists {
			wsInfo = domain.Workspace{
				ID:   wsID,
				Name: wsID,
			}
		}

		summary.Workspaces = append(summary.Workspaces, domain.WorkspaceSummary{
			Workspace: wsInfo,
			Tokens:    usage,
		})
	}

	// Sort workspaces by total tokens descending
	sort.Slice(summary.Workspaces, func(i, j int) bool {
		return summary.Workspaces[i].Tokens.Total > summary.Workspaces[j].Tokens.Total
	})

	return summary, nil
}

// GetOverallTotal returns total token counts accumulated across all logs.
func (u *TokenUsecase) GetOverallTotal() domain.TokenUsage {
	u.mu.RLock()
	defer u.mu.RUnlock()

	var total domain.TokenUsage
	for _, event := range u.events {
		total.Prompt += event.Tokens.Prompt
		total.Completion += event.Tokens.Completion
		total.Total += event.Tokens.Total
		total.AIC += event.Tokens.AIC
		total.AIU += event.Tokens.AIU
		total.Requests += event.Tokens.Requests
	}
	return total
}
