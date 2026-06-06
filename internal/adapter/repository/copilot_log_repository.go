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
	"bufio"
	"encoding/json"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"credit-count/internal/domain"
)

// Regex to match "0.3 credits" or "1x"
var detailsRegex = regexp.MustCompile(`(\d+(?:\.\d+)?)\s*(?:credits?|x)`)

// CopilotLogRepository implements domain.TokenRepository for VS Code storage.
type CopilotLogRepository struct {
	storagePath string
}

// NewCopilotLogRepository creates a new repository.
// If storagePath is empty, it uses the default VS Code workspace storage path for the current OS.
func NewCopilotLogRepository(storagePath string) *CopilotLogRepository {
	if storagePath == "" {
		configDir, err := os.UserConfigDir()
		if err != nil {
			// Fallback if user config directory cannot be determined
			if home, errHome := os.UserHomeDir(); errHome == nil {
				configDir = filepath.Join(home, ".config")
			}
		}
		storagePath = filepath.Join(configDir, "Code", "User", "workspaceStorage")
	}
	return &CopilotLogRepository{storagePath: storagePath}
}

// ScanSessions scans all workspaces for Copilot chat sessions, utilizing a cache.
func (r *CopilotLogRepository) ScanSessions() ([]domain.SessionEvent, map[string]domain.Workspace, error) {
	workspaces := make(map[string]domain.Workspace)
	var events []domain.SessionEvent

	// Check if workspace storage path exists.
	if _, err := os.Stat(r.storagePath); os.IsNotExist(err) {
		return nil, nil, err
	}

	cachePath := filepath.Join(r.storagePath, "credit-count-cache.json")
	oldCache := r.loadCache(cachePath)
	newCache := ScanCache{
		Files: make(map[string]FileCacheEntry),
	}

	// Read workspace directories.
	entries, err := os.ReadDir(r.storagePath)
	if err != nil {
		return nil, nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		workspaceID := entry.Name()
		workspaceDir := filepath.Join(r.storagePath, workspaceID)

		// 1. Resolve workspace name from workspace.json
		wsInfo := r.resolveWorkspace(workspaceDir, workspaceID)
		workspaces[workspaceID] = wsInfo

		// 2. Scan chatSessions directory
		sessionsDir := filepath.Join(workspaceDir, "chatSessions")
		if _, err := os.Stat(sessionsDir); os.IsNotExist(err) {
			continue
		}

		sessionEntries, err := os.ReadDir(sessionsDir)
		if err != nil {
			continue
		}

		for _, sEntry := range sessionEntries {
			if sEntry.IsDir() {
				continue
			}

			filePath := filepath.Join(sessionsDir, sEntry.Name())
			ext := strings.ToLower(filepath.Ext(sEntry.Name()))
			sessionID := strings.TrimSuffix(sEntry.Name(), filepath.Ext(sEntry.Name()))

			if ext != ".json" && ext != ".jsonl" {
				continue
			}

			fileInfo, err := os.Stat(filePath)
			if err != nil {
				continue
			}
			modTime := fileInfo.ModTime()
			size := fileInfo.Size()

			relPath, err := filepath.Rel(r.storagePath, filePath)
			if err != nil {
				relPath = filePath
			}
			relPath = strings.ReplaceAll(relPath, "\\", "/")

			var sessionEvents []domain.SessionEvent
			if cachedEntry, exists := oldCache.Files[relPath]; exists && cachedEntry.ModTime.Equal(modTime) && cachedEntry.Size == size {
				sessionEvents = cachedEntry.Events
			} else {
				if ext == ".json" {
					sessionEvents = r.parseJSONSession(filePath, workspaceID, sessionID)
				} else if ext == ".jsonl" {
					sessionEvents = r.parseJSONLSession(filePath, workspaceID, sessionID)
				}
			}

			newCache.Files[relPath] = FileCacheEntry{
				Path:    relPath,
				ModTime: modTime,
				Size:    size,
				Events:  sessionEvents,
			}

			events = append(events, sessionEvents...)
		}
	}

	// Persist updated cache
	_ = r.saveCache(cachePath, newCache)

	return events, workspaces, nil
}

// resolveWorkspace reads workspace.json to get the human-readable project path and name.
func (r *CopilotLogRepository) resolveWorkspace(dir, id string) domain.Workspace {
	ws := domain.Workspace{
		ID:   id,
		Path: "",
		Name: id, // Default fallback to ID
	}

	wsJSONPath := filepath.Join(dir, "workspace.json")
	file, err := os.Open(wsJSONPath)
	if err != nil {
		return ws
	}
	defer file.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return ws
	}

	var uriStr string
	if folder, ok := data["folder"].(string); ok {
		uriStr = folder
	} else if workspace, ok := data["workspace"].(string); ok {
		uriStr = workspace
	}

	if uriStr != "" {
		decodedPath, name := parseWorkspaceURI(uriStr)
		ws.Path = decodedPath
		ws.Name = name
	}

	return ws
}

// parseJSONSession parses older single-JSON chat session files.
func (r *CopilotLogRepository) parseJSONSession(filePath, workspaceID, sessionID string) []domain.SessionEvent {
	file, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer file.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return nil
	}

	requestsRaw, ok := data["requests"].([]interface{})
	if !ok {
		return nil
	}

	var events []domain.SessionEvent
	for _, reqRaw := range requestsRaw {
		reqMap, ok := reqRaw.(map[string]interface{})
		if !ok {
			continue
		}

		// Parse timestamp
		ts := parseTimestamp(reqMap["timestamp"])

		// Extract tokens and credits recursively from the request map
		prompt, comp, total, aic, aiu := r.extractTokensFromMap(reqMap)

		// If total is 0 but prompt or comp are set
		if total == 0 && (prompt > 0 || comp > 0) {
			total = prompt + comp
		}

		// If no tokens or credits found, don't create empty event
		if total == 0 && aic == 0 {
			continue
		}

		events = append(events, domain.SessionEvent{
			WorkspaceID: workspaceID,
			SessionID:   sessionID,
			Timestamp:   ts,
			Tokens: domain.TokenUsage{
				Prompt:     prompt,
				Completion: comp,
				Total:      total,
				AIC:        aic,
				AIU:        aiu,
				Requests:   1,
			},
		})
	}

	return events
}

// parseJSONLSession parses newer line-delimited JSONL chat session files.
func (r *CopilotLogRepository) parseJSONLSession(filePath, workspaceID, sessionID string) []domain.SessionEvent {
	file, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer file.Close()

	var events []domain.SessionEvent
	reader := bufio.NewReader(file)

	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			if err == io.EOF {
				break
			}
			continue
		}

		var eventMap map[string]interface{}
		if err := json.Unmarshal([]byte(line), &eventMap); err != nil {
			if err == io.EOF {
				break
			}
			continue
		}

		// Extract timestamp
		var ts time.Time
		if tVal, ok := eventMap["timestamp"]; ok {
			ts = parseTimestamp(tVal)
		} else if tsVal, ok := eventMap["ts"]; ok {
			ts = parseTimestamp(tsVal)
		}

		if ts.IsZero() {
			ts = time.Now() // Fallback
		}

		// Look for root-level token and credit keys first to avoid double counting nested tool call tokens.
		prompt, comp, total, aic, aiu := r.extractRootTokens(eventMap)

		// If no root tokens/credits are found, fall back to recursive scanning.
		if total == 0 && prompt == 0 && comp == 0 && aic == 0 {
			prompt, comp, total, aic, aiu = r.extractTokensFromMap(eventMap)
		}

		if total == 0 && (prompt > 0 || comp > 0) {
			total = prompt + comp
		}

		// Skip events with no tokens and no credits
		if total > 0 || aic > 0 {
			events = append(events, domain.SessionEvent{
				WorkspaceID: workspaceID,
				SessionID:   sessionID,
				Timestamp:   ts,
				Tokens: domain.TokenUsage{
					Prompt:     prompt,
					Completion: comp,
					Total:      total,
					AIC:        aic,
					AIU:        aiu,
					Requests:   1,
				},
			})
		}

		if err == io.EOF {
			break
		}
	}

	return events
}

// parseDetails parses string fields (like details) to extract AIC and AIU values.
func (r *CopilotLogRepository) parseDetails(val interface{}) (aic, aiu float64) {
	str, ok := val.(string)
	if !ok {
		return 0, 0
	}
	matches := detailsRegex.FindStringSubmatch(str)
	if len(matches) > 1 {
		if creditVal, err := strconv.ParseFloat(matches[1], 64); err == nil {
			return creditVal, creditVal * 1000000000 // Convert to nano-AIU
		}
	}
	return 0, 0
}

// extractRootTokens looks for token properties directly on the root of the event map.
func (r *CopilotLogRepository) extractRootTokens(m map[string]interface{}) (prompt, comp, total int, aic, aiu float64) {
	for k, v := range m {
		kLower := strings.ToLower(k)
		switch kLower {
		case "prompttokens", "prompt_tokens":
			prompt += toInt(v)
		case "completiontokens", "completion_tokens", "outputtokens", "output_tokens":
			comp += toInt(v)
		case "totaltokens", "total_tokens", "tokens":
			total += toInt(v)
		case "details":
			aicVal, aiuVal := r.parseDetails(v)
			aic += aicVal
			aiu += aiuVal
		}
	}
	return prompt, comp, total, aic, aiu
}

// extractTokensFromMap recursively walks the map to aggregate all token and credit fields.
func (r *CopilotLogRepository) extractTokensFromMap(m map[string]interface{}) (prompt, comp, total int, aic, aiu float64) {
	for k, v := range m {
		kLower := strings.ToLower(k)
		switch kLower {
		case "prompttokens", "prompt_tokens":
			prompt += toInt(v)
		case "completiontokens", "completion_tokens", "outputtokens", "output_tokens":
			comp += toInt(v)
		case "totaltokens", "total_tokens", "tokens":
			total += toInt(v)
		case "details":
			aicVal, aiuVal := r.parseDetails(v)
			aic += aicVal
			aiu += aiuVal
		default:
			// Recurse into nested structures
			p, c, t, ac, au := r.extractTokensRecursive(v)
			prompt += p
			comp += c
			total += t
			aic += ac
			aiu += au
		}
	}
	return prompt, comp, total, aic, aiu
}

// extractTokensRecursive checks the type of the value and recurses accordingly.
func (r *CopilotLogRepository) extractTokensRecursive(val interface{}) (prompt, comp, total int, aic, aiu float64) {
	switch v := val.(type) {
	case map[string]interface{}:
		return r.extractTokensFromMap(v)
	case []interface{}:
		for _, item := range v {
			p, c, t, ac, au := r.extractTokensRecursive(item)
			prompt += p
			comp += c
			total += t
			aic += ac
			aiu += au
		}
	}
	return prompt, comp, total, aic, aiu
}

// Helper: parseTimestamp parses interface values representing timestamps.
func parseTimestamp(val interface{}) time.Time {
	if val == nil {
		return time.Time{}
	}

	switch v := val.(type) {
	case float64:
		// Millisecond Unix timestamp
		sec := int64(v / 1000)
		nsec := int64(v) % 1000 * 1000000
		return time.Unix(sec, nsec)
	case string:
		// ISO/RFC3339 string
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t
		}
		if t, err := time.Parse("2006-01-02T15:04:05.000Z", v); err == nil {
			return t
		}
		if t, err := time.Parse("2006-01-02 15:04:05.000 MST", v); err == nil {
			return t
		}
	}

	return time.Time{}
}

// Helper: parseWorkspaceURI parses the file URL into a path and name.
func parseWorkspaceURI(uriStr string) (path string, name string) {
	path = uriStr
	if strings.HasPrefix(uriStr, "file://") {
		u, err := url.Parse(uriStr)
		if err == nil {
			path = u.Path
			// Windows file:/// URI format handling (e.g., file:///c%3A/path -> /c:/path)
			if len(path) > 3 && path[0] == '/' && path[2] == ':' {
				path = path[1:]
			}
			// Decode percent encoding
			decodedPath, err := url.PathUnescape(path)
			if err == nil {
				path = decodedPath
			}
		}
	}
	
	path = filepath.Clean(path)
	path = strings.ReplaceAll(path, "\\", "/")

	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		name = parts[len(parts)-1]
	}
	if name == "" {
		name = path
	}
	return path, name
}

// Helper: convert interface to int.
func toInt(val interface{}) int {
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case float64:
		return int(v)
	case float32:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	}
	return 0
}

// FileCacheEntry holds filesystem metadata and aggregated session events for a log file.
type FileCacheEntry struct {
	Path    string                `json:"path"`
	ModTime time.Time             `json:"modTime"`
	Size    int64                 `json:"size"`
	Events  []domain.SessionEvent `json:"events"`
}

// ScanCache acts as a repository cache for copilot logs.
type ScanCache struct {
	Files map[string]FileCacheEntry `json:"files"`
}

// loadCache loads the cached file mappings from disk.
func (r *CopilotLogRepository) loadCache(cachePath string) ScanCache {
	cache := ScanCache{
		Files: make(map[string]FileCacheEntry),
	}
	file, err := os.Open(cachePath)
	if err != nil {
		return cache
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&cache); err != nil {
		return ScanCache{Files: make(map[string]FileCacheEntry)}
	}
	return cache
}

// saveCache writes the cached file mappings back to disk.
func (r *CopilotLogRepository) saveCache(cachePath string, cache ScanCache) error {
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		return err
	}

	file, err := os.Create(cachePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(cache)
}
