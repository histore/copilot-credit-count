# Github Copilot Credit Count

A lightweight, local desktop application built with **Go**, **Wails v2**, and **Vanilla JS** that scans, aggregates, and visualizes token consumption and credit usage from your local GitHub Copilot chat session logs.

## Features

- **Local Scan & Analysis**: Automatically locates VS Code workspace storage on your Windows system and extracts Copilot token metrics from chat history files (`.json`/`.jsonl`).
- **Token & Credit Insights**: Tracks Prompt Tokens, Completion Tokens, Total Tokens, AI Credits (AIC), AI Usage (AIU), and individual request counts.
- **Workspace Isolation**: Shows exactly which project or workspace is consuming how many tokens and credits.
- **Monthly Statistics**: Filters usage data by month with detailed workspace-specific breakdowns.
- **Performance Caching**: Implements a JSON-based filesystem caching layer (`credit-count-cache.json`) to skip unmodified files during subsequent scans, accelerating load times.
- **Internationalization (i18n)**: Fully localized interface support. Displays in English by default and automatically switches to German when the system locale is set to German. Manual toggle available in the header.
- **Sleek UX**: Dynamic design with responsive styling, a glassmorphic dark-theme card layout, loading indicators, and informative tooltips.

## Architecture

This application strictly adheres to **Clean Architecture** and **Clean Code** principles:
- **Domain (`internal/domain`)**: Defines domain entities (`Workspace`, `TokenUsage`, `SessionEvent`, `MonthSummary`) and boundary interfaces (`TokenRepository`). Free of external dependencies.
- **Use Cases (`internal/usecase`)**: Contains core orchestrator logic for caching, sorting, and aggregating metrics, utilizing Go's native synchronization primitives (`sync.RWMutex`).
- **Adapters (`internal/adapter`)**:
  - `repository`: Implements the file parser (`JSON`/`JSONL`), directory scanning, and caching storage.
  - `ui`: Direct glue layer for Wails runtime binding and context registration.

## Prerequisites

Before building or running the application, ensure your development machine meets the following requirements:

### General Requirements
- **Go**: Version 1.20 or higher.
- **Node.js**: Version 16 or higher (along with `npm`).
- **Wails CLI**: Installed via Go:
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  ```

### Platform-Specific Requirements

#### Windows
- **WebView2 Runtime**: Typically pre-installed on Windows 10/11.
- **NSIS**: Required if you want to build a Windows installer package.

#### Linux
- **CGO compilation tools**: `build-essential` and `pkg-config`.
- **GTK3 and WebKit2Gtk development libraries**:
  On Debian/Ubuntu-based distributions, install them via:
  ```bash
  sudo apt update
  sudo apt install build-essential pkg-config libgtk-3-dev libwebkit2gtk-4.0-dev
  ```

#### macOS
- **Xcode Command Line Tools**: Install using `xcode-select --install`.

## Live Development

To run the application in live development mode:
1. Ensure all general and platform-specific prerequisites are installed.
2. Execute the following command in the root directory:
   ```bash
   wails dev
   ```
This compiles the Go backend and launches Vite in hot-reload mode for frontend development.

## Building

To build a standalone, production-ready desktop package:

### Standard Build
Run the build command in the project root:
```bash
wails build
```
The resulting executable will be placed in the `build/bin/` directory.

### Cross-Compilation

Since Wails relies on CGO and links native platform libraries (such as WebKit2Gtk on Linux, WebView2 on Windows, or Cocoa on macOS), you cannot simply cross-compile by setting standard Go environment variables like `GOOS` and `GOARCH`.

Instead, the recommended approaches are:
1. **CI/CD Build Pipelines (Highly Recommended)**: Use platforms like GitHub Actions or GitLab CI to build natively on runners of each respective operating system (Windows, Ubuntu, macOS). You can use community actions such as `dAppServer/wails-build-action`.
2. **Docker / Containerization**: Use Docker containers with pre-configured cross-compiler toolchains and libraries.

## License

This project is licensed under the Apache License, Version 2.0. See the [LICENSE](LICENSE) file for details.
