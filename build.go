//usr/bin/env go run "$0" "$@"; exit
//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

// Central version definition
const version = "1.6.1"

func main() {
	fmt.Printf("Starting build process for version %s...\n", version)

	// 1. Update wails.json
	fmt.Println("Updating wails.json...")
	if err := updateWailsJSON(version); err != nil {
		fmt.Printf("Failed to update wails.json: %v\n", err)
		os.Exit(1)
	}

	// 2. Build Wails app (GUI & Installer)
	fmt.Println("Building Wails application...")
	cmdWails := exec.Command("wails", "build", "-clean")
	cmdWails.Stdout = os.Stdout
	cmdWails.Stderr = os.Stderr
	if err := cmdWails.Run(); err != nil {
		fmt.Printf("Wails build failed: %v\n", err)
		os.Exit(1)
	}

	// 3. Build Go CLI
	fmt.Println("Building CLI application...")
	ldflags := fmt.Sprintf("-X main.Version=%s", version)

	// Determine output name based on OS (add .exe on windows)
	outPath := "bin/github-copilot-credit-count-cli"
	if os.Getenv("GOOS") == "windows" || (os.Getenv("GOOS") == "" && os.PathSeparator == '\\') {
		outPath += ".exe"
	}

	cmdGo := exec.Command("go", "build", "-ldflags", ldflags, "-o", outPath, "./cmd/github-copilot-credit-count-cli")
	cmdGo.Stdout = os.Stdout
	cmdGo.Stderr = os.Stderr
	if err := cmdGo.Run(); err != nil {
		fmt.Printf("CLI build failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Build completed successfully. Executables are in the build/bin and bin/ directories.")
}

func updateWailsJSON(v string) error {
	data, err := os.ReadFile("wails.json")
	if err != nil {
		return err
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	info, ok := config["info"].(map[string]interface{})
	if !ok {
		info = make(map[string]interface{})
		config["info"] = info
	}

	info["productVersion"] = v

	// Ensure required fields for NSIS
	if _, exists := info["companyName"]; !exists {
		info["companyName"] = "histore"
	}
	if _, exists := info["productName"]; !exists {
		info["productName"] = "github-copilot-credit-count"
	}

	newData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile("wails.json", newData, 0644)
}
