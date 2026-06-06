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

package main

import (
	"embed"

	"credit-count/internal/adapter/repository"
	"credit-count/internal/adapter/ui"
	"credit-count/internal/usecase"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Initialize clean architecture layers
	repo := repository.NewCopilotLogRepository("")
	uc := usecase.NewTokenUsecase(repo)
	wailsAdapter := ui.NewWailsAdapter(uc)

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "Github Copilot Credit Count",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 15, G: 23, B: 42, A: 1}, // Sleek dark slate color
		OnStartup:        wailsAdapter.Startup,
		Bind: []interface{}{
			wailsAdapter,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
