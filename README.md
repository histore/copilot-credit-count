# README

## About

This is the official Wails Vanilla template.

You can configure the project by editing `wails.json`. More information about the project settings can be found
here: https://wails.io/docs/reference/project-config

## Live Development

To run in live development mode, run `wails dev` in the project directory. This will run a Vite development
server that will provide very fast hot reload of your frontend changes. If you want to develop in a browser
and have access to your Go methods, there is also a dev server that runs on http://localhost:34115. Connect
to this in your browser, and you can call your Go code from devtools.

## Building

To build a redistributable, production-ready desktop package, use the following commands depending on your target platform.

### Standard Build

Run the main build command in the project root:
```bash
wails build
```

This will automatically bundle the frontend assets and compile a production binary for your current operating system.

### Platform Specifics

#### Windows
Requirements:
* WebView2 Runtime (typically pre-installed on Windows 10/11)
* Go and Node.js

To build for Windows:
```bash
wails build
```

#### Linux
Requirements:
* CGO compilation tools (`build-essential`, `pkg-config`)
* GTK3 and WebKit2Gtk development libraries

On Debian/Ubuntu-based distributions, install the requirements with:
```bash
sudo apt update
sudo apt install build-essential pkg-config libgtk-3-dev libwebkit2gtk-4.0-dev
```
Then build the binary:
```bash
wails build
```

#### macOS
To build a macOS application bundle (`.app`):
```bash
wails build
```
*Note: To distribute macOS applications, the binary should ideally be code-signed and notarized.*

### Cross-Compilation

Since Wails relies on CGO and links native platform libraries (like WebKit2Gtk on Linux, WebView2 on Windows, or Cocoa on macOS), you cannot simply cross-compile by setting standard Go environment variables like `GOOS` and `GOARCH`.

Instead, the recommended approaches are:
1. **CI/CD Build Pipelines (Highly Recommended)**: Use platforms like GitHub Actions or GitLab CI to build natively on runners of each respective operating system (Windows, Ubuntu, macOS). You can use community actions such as `dAppServer/wails-build-action`.
2. **Docker / Containerization**: Use Docker containers with pre-configured cross-compiler toolchains and libraries.

## License

This project is licensed under the Apache License, Version 2.0. See the [LICENSE](LICENSE) file for the full text.
