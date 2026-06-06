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

package ui

import (
	"context"

	"github-copilot-credit-count/internal/domain"
	"github-copilot-credit-count/internal/usecase"
)

// WailsAdapter coordinates frontend requests and business logic.
type WailsAdapter struct {
	ctx     context.Context
	usecase *usecase.TokenUsecase
}

// NewWailsAdapter creates a new Wails UI Adapter.
func NewWailsAdapter(u *usecase.TokenUsecase) *WailsAdapter {
	return &WailsAdapter{
		usecase: u,
	}
}

// Startup is called by Wails when the app starts.
func (a *WailsAdapter) Startup(ctx context.Context) {
	a.ctx = ctx
}

// ScanLogs scans all VS Code storage directories and caches session logs.
func (a *WailsAdapter) ScanLogs() error {
	return a.usecase.ScanAndCache()
}

// GetAvailableMonths returns the list of months found in the logs.
func (a *WailsAdapter) GetAvailableMonths() []string {
	return a.usecase.GetAvailableMonths()
}

// GetSummaryForMonth gets the detailed token breakdown for a specific month.
func (a *WailsAdapter) GetSummaryForMonth(month string) (domain.MonthSummary, error) {
	return a.usecase.GetSummaryForMonth(month)
}

// GetOverallTotal gets the total token usage across all sessions.
func (a *WailsAdapter) GetOverallTotal() domain.TokenUsage {
	return a.usecase.GetOverallTotal()
}
