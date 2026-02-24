# secctl - Kubernetes Secrets Interactive Editor

A small TUI tool to browse and edit Kubernetes secrets without YAML or base64.

[![Test](https://github.com/g4s8/secctl/actions/workflows/test.yml/badge.svg)](https://github.com/g4s8/secctl/actions/workflows/test.yml)
![GitHub License](https://img.shields.io/github/license/g4s8/secctl)
![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/g4s8/secctl/total)

## Demo

![demo.gif](demo.gif)

## Features

- **Interactive selection** - Browse namespaces, secrets, and keys with search support
- **External editor support** - Edit secrets in your preferred editor (vim, nano, emacs, etc.)
- **Search** - Fuzzy search on every step
- **Diff** - Preview diff and confirm save

## Installation

### Download pre-built binaries
Download the latest release from [GitHub Releases](https://github.com/g4s8/secctl/releases):

### Using homebrew

```
brew install g4s8/tap/secctl
```

### Using go install
```bash
go install github.com/g4s8/secctl@latest
```

### Build from source
```bash
git clone https://github.com/g4s8/secctl
cd secctl
make install
```

## Configuration

### Environment Variables
- `EDITOR` - Default text editor to use when `--editor` is not specified
- `KUBECONFIG` - Default kubeconfig path when `--kubeconfig` is not specified

### Command-line Flags
```
-editor string
    Path to the text editor (default: $EDITOR)
-kubeconfig string
    Path to the kubeconfig file (default: $KUBECONFIG or ~/.kube/config)
```
