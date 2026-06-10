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

package repository

import (
	"os"
	"path/filepath"
	"testing"

	"github-copilot-credit-count/internal/domain"
)

func TestParseWorkspaceURI(t *testing.T) {
	tests := []struct {
		input        string
		expectedPath string
		expectedName string
	}{
		{
			input:        "file:///c%3A/projekte/markinglogviewer",
			expectedPath: "c:/projekte/markinglogviewer",
			expectedName: "markinglogviewer",
		},
		{
			input:        "file:///C:/Users/test/workspace",
			expectedPath: "C:/Users/test/workspace",
			expectedName: "workspace",
		},
		{
			input:        "c:\\projekte\\some-app",
			expectedPath: "c:/projekte/some-app",
			expectedName: "some-app",
		},
	}

	for _, tt := range tests {
		p, n := parseWorkspaceURI(tt.input)
		if p != tt.expectedPath {
			t.Errorf("expected path %q, got %q for input %q", tt.expectedPath, p, tt.input)
		}
		if n != tt.expectedName {
			t.Errorf("expected name %q, got %q for input %q", tt.expectedName, n, tt.input)
		}
	}
}

func TestParseJSONSession(t *testing.T) {
	// Create a temporary test folder inside the workspace
	testDir := filepath.Join(".", "test_temp_json_credits")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	mockJSONContent := `{
		"version": 3,
		"requests": [
			{
				"timestamp": 1780670330207,
				"response": [
					{
						"kind": "thinking",
						"tokens": 192,
						"details": "Raptor mini (Preview) • 0.3 credits"
					}
				]
			},
			{
				"timestamp": 1780670350207,
				"response": [
					{
						"kind": "thinking",
						"tokens": 108
					},
					{
						"kind": "toolCall",
						"tokens": 200,
						"details": "GPT-5 mini • 1x"
					}
				]
			}
		]
	}`

	filePath := filepath.Join(testDir, "session.json")
	if err := os.WriteFile(filePath, []byte(mockJSONContent), 0644); err != nil {
		t.Fatalf("failed writing mock json: %v", err)
	}

	repo := NewCopilotLogRepository("")
	events := repo.parseJSONSession(filePath, "test-ws", "session")

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	// First request: total tokens 192, 0.3 credits
	if events[0].Tokens.Total != 192 {
		t.Errorf("expected 192 total tokens in first event, got %d", events[0].Tokens.Total)
	}
	if events[0].Tokens.AIC != 0.3 {
		t.Errorf("expected 0.3 AIC in first event, got %f", events[0].Tokens.AIC)
	}
	if events[0].Tokens.AIU != 300000000 {
		t.Errorf("expected 300,000,000 nano-AIU in first event, got %f", events[0].Tokens.AIU)
	}

	// Second request: 108 + 200 = 308 tokens, 1x details = 1.0 credit
	if events[1].Tokens.Total != 308 {
		t.Errorf("expected 308 total tokens in second event, got %d", events[1].Tokens.Total)
	}
	if events[1].Tokens.AIC != 1.0 {
		t.Errorf("expected 1.0 AIC in second event, got %f", events[1].Tokens.AIC)
	}
	if events[1].Tokens.AIU != 1000000000 {
		t.Errorf("expected 1,000,000,000 nano-AIU in second event, got %f", events[1].Tokens.AIU)
	}
}

func TestParseJSONLSession(t *testing.T) {
	// Create a temporary test folder inside the workspace
	testDir := filepath.Join(".", "test_temp_jsonl_credits")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	mockJSONLContent := `{"ts": 1780670330000, "promptTokens": 100, "completionTokens": 50, "details": "Raptor mini • 0.3 credits"}
{"timestamp": "2026-06-05T14:38:45.000Z", "toolCallRounds": [{"tokens": 128, "details": "Claude Haiku 4.5 • 1x"}]}`

	filePath := filepath.Join(testDir, "session.jsonl")
	if err := os.WriteFile(filePath, []byte(mockJSONLContent), 0644); err != nil {
		t.Fatalf("failed writing mock jsonl: %v", err)
	}

	repo := NewCopilotLogRepository("")
	events := repo.parseJSONLSession(filePath, "test-ws", "session")

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	var event1, event2 domain.SessionEvent
	if events[0].Tokens.Total == 150 {
		event1 = events[0]
		event2 = events[1]
	} else {
		event1 = events[1]
		event2 = events[0]
	}

	// Line 1: 100 prompt, 50 comp, 150 total, 0.3 AIC, 300,000,000 AIU
	if event1.Tokens.Prompt != 100 || event1.Tokens.Completion != 50 || event1.Tokens.Total != 150 {
		t.Errorf("first event token counts mismatch: %+v", event1.Tokens)
	}
	if event1.Tokens.AIC != 0.3 || event1.Tokens.AIU != 300000000 {
		t.Errorf("first event credits mismatch: AIC=%f, AIU=%f", event1.Tokens.AIC, event1.Tokens.AIU)
	}

	// Line 2: 128 total tokens, 1.0 AIC, 1,000,000,000 AIU
	if event2.Tokens.Total != 128 {
		t.Errorf("second event token count mismatch, expected 128 total, got %d", event2.Tokens.Total)
	}
	if event2.Tokens.AIC != 1.0 || event2.Tokens.AIU != 1000000000 {
		t.Errorf("second event credits mismatch: AIC=%f, AIU=%f", event2.Tokens.AIC, event2.Tokens.AIU)
	}
}

func TestScanSessionsWithCache(t *testing.T) {
	// Create a temporary storage directory
	storageDir, err := os.MkdirTemp("", "github-copilot-credit-count-test-*")
	if err != nil {
		t.Fatalf("failed to create temp storage dir: %v", err)
	}
	defer os.RemoveAll(storageDir)

	wsID := "workspace1"
	wsDir := filepath.Join(storageDir, wsID)
	sessionsDir := filepath.Join(wsDir, "chatSessions")
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		t.Fatalf("failed to create sessions dir: %v", err)
	}

	// Create workspace.json
	wsJSONContent := `{"folder": "file:///C:/Users/test/workspace"}`
	if err := os.WriteFile(filepath.Join(wsDir, "workspace.json"), []byte(wsJSONContent), 0644); err != nil {
		t.Fatalf("failed to write workspace.json: %v", err)
	}

	// Create a log file with 1 event
	logFile := filepath.Join(sessionsDir, "session1.jsonl")
	logContent := `{"ts": 1780670330000, "promptTokens": 100, "completionTokens": 50, "details": "Raptor mini • 0.3 credits"}`
	if err := os.WriteFile(logFile, []byte(logContent), 0644); err != nil {
		t.Fatalf("failed to write log file: %v", err)
	}

	repo := NewCopilotLogRepository(storageDir)

	// First scan (should create cache and parse file)
	events, _, err := repo.ScanSessions()
	if err != nil {
		t.Fatalf("first scan failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Tokens.Total != 150 {
		t.Errorf("expected 150 tokens, got %d", events[0].Tokens.Total)
	}

	// Verify cache file was created
	cachePath := filepath.Join(storageDir, "github-copilot-credit-count-cache-v2.json")
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Errorf("expected cache file to be created at %s", cachePath)
	}

	// Get initial file attributes
	fi, err := os.Stat(logFile)
	if err != nil {
		t.Fatalf("failed to stat log file: %v", err)
	}
	initialModTime := fi.ModTime()

	// Modify the file content on disk, but force the same ModTime and size so it hits the cache.
	// Since size must match, we write content of exactly the same byte length.
	// New content represents 500 tokens, but since it hits the cache, it should still yield the old 150 token event.
	modifiedContent := `{"ts": 1780670330000, "promptTokens": 300, "completionTokens": 50, "details": "Raptor mini • 0.3 credits"}`
	if len(modifiedContent) != len(logContent) {
		t.Fatalf("test configuration error: modified content length (%d) does not match original (%d)", len(modifiedContent), len(logContent))
	}
	if err := os.WriteFile(logFile, []byte(modifiedContent), 0644); err != nil {
		t.Fatalf("failed to write modified content: %v", err)
	}

	// Restore mod time
	if err := os.Chtimes(logFile, initialModTime, initialModTime); err != nil {
		t.Fatalf("failed to restore mod time: %v", err)
	}

	// Second scan (Cache Hit expected - should yield the original 150 token event, not the new 500 token one)
	events2, _, err := repo.ScanSessions()
	if err != nil {
		t.Fatalf("second scan failed: %v", err)
	}
	if len(events2) != 1 {
		t.Fatalf("expected 1 event on second scan, got %d", len(events2))
	}
	if events2[0].Tokens.Total != 150 {
		t.Errorf("expected cache hit yielding 150 tokens, but got %d (which indicates a cache miss/reparse)", events2[0].Tokens.Total)
	}

	// Now modify the file size to force a cache miss (new length, different content)
	differentSizeContent := `{"ts": 1780670330000, "promptTokens": 500, "completionTokens": 500, "details": "Raptor mini • 0.3 credits"}` // different length
	if err := os.WriteFile(logFile, []byte(differentSizeContent), 0644); err != nil {
		t.Fatalf("failed to write different size content: %v", err)
	}
	// Restore mod time again, but size is different now.
	if err := os.Chtimes(logFile, initialModTime, initialModTime); err != nil {
		t.Fatalf("failed to restore mod time: %v", err)
	}

	// Third scan (Cache Miss expected due to size mismatch - should yield 1000 tokens)
	events3, _, err := repo.ScanSessions()
	if err != nil {
		t.Fatalf("third scan failed: %v", err)
	}
	if len(events3) != 1 {
		t.Fatalf("expected 1 event on third scan, got %d", len(events3))
	}
	if events3[0].Tokens.Total != 1000 {
		t.Errorf("expected cache miss yielding 1000 tokens, but got %d", events3[0].Tokens.Total)
	}
}

func TestParseJSONLSessionWithNestedRequestsAndDeduplication(t *testing.T) {
	testDir := filepath.Join(".", "test_temp_jsonl_nested")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Line 1: Request 1 exists with 10 tokens (completedAt June)
	// Line 2: Request 1 is updated to 50 tokens, Request 2 is added with 100 tokens (completedAt July)
	mockJSONLContent := `{"kind":0,"v":{"requests":[{"requestId":"req1","modelState":{"value":1,"completedAt":1780670330000},"promptTokens":10,"completionTokens":0}]}}
{"kind":2,"v":{"requests":[{"requestId":"req1","modelState":{"value":1,"completedAt":1780670330000},"promptTokens":20,"completionTokens":30},{"requestId":"req2","modelState":{"value":1,"completedAt":1783262330000},"promptTokens":50,"completionTokens":50}]}}`

	filePath := filepath.Join(testDir, "session.jsonl")
	if err := os.WriteFile(filePath, []byte(mockJSONLContent), 0644); err != nil {
		t.Fatalf("failed writing mock jsonl: %v", err)
	}

	repo := NewCopilotLogRepository("")
	events := repo.parseJSONLSession(filePath, "test-ws", "session")

	if len(events) != 2 {
		t.Fatalf("expected 2 unique request events, got %d", len(events))
	}

	var req1, req2 *domain.SessionEvent
	for i := range events {
		if events[i].Timestamp.Unix() == 1780670330 {
			req1 = &events[i]
		} else if events[i].Timestamp.Unix() == 1783262330 {
			req2 = &events[i]
		}
	}

	if req1 == nil {
		t.Fatalf("could not find req1 by timestamp")
	}
	if req2 == nil {
		t.Fatalf("could not find req2 by timestamp")
	}

	// Verify req1 has the updated tokens (20 prompt + 30 completion = 50 total)
	if req1.Tokens.Prompt != 20 || req1.Tokens.Completion != 30 || req1.Tokens.Total != 50 {
		t.Errorf("req1 tokens mismatch: %+v", req1.Tokens)
	}

	// Verify req2 has correct tokens (50 prompt + 50 completion = 100 total)
	if req2.Tokens.Prompt != 50 || req2.Tokens.Completion != 50 || req2.Tokens.Total != 100 {
		t.Errorf("req2 tokens mismatch: %+v", req2.Tokens)
	}
}

