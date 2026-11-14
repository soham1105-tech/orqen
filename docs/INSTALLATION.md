# Installing ORQEN

This guide provides step-by-step instructions for installing and configuring **ORQEN** across Linux, macOS, and Windows environments.

---

## System Requirements

| Component | Minimum Version | Recommended | Notes |
|-----------|-----------------|-------------|-------|
| **Operating System** | Linux (kernel 5.4+), macOS 12+ (Monterey), Windows 10/11 | Ubuntu 22.04+, macOS 14+ (Sonoma), Windows 11 | x86_64 / amd64 or ARM64 / aarch64 |
| **Go** | 1.27+ | 1.27.4+ | Required only when building or installing from source |
| **Docker** | 20.10+ | Latest Docker Desktop or Docker Engine | Required for running stacks and containerized tests |
| **Docker Compose** | Compose v2 (2.20+) | Compose v2.29+ | Must be accessible as `docker compose` |
| **Git** | 2.30+ | Latest | Required for cloning repositories |
| **LLM Provider API Key** | (Optional at runtime) | Gemini / OpenAI / Anthropic / Groq / Ollama | Required for conversational agent, intent fallback, and AI healing |

---

## Installation Methods

### Method 1: Pre-built Binary Releases (Recommended for Users)

Download the pre-compiled binary for your operating system and architecture from the [GitHub Releases](https://github.com/soham1105-tech/orqen/releases) page.

#### Linux (x86_64 / amd64)

```bash
# Download latest release
VERSION=$(curl -s https://api.github.com/repos/soham1105-tech/orqen/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
curl -sSL "https://github.com/soham1105-tech/orqen/releases/download/${VERSION}/orqen_Linux_x86_64.tar.gz" | tar -xz
sudo mv orqen /usr/local/bin/
sudo chmod +x /usr/local/bin/orqen
```

#### Linux (ARM64 / aarch64)

```bash
VERSION=$(curl -s https://api.github.com/repos/soham1105-tech/orqen/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
curl -sSL "https://github.com/soham1105-tech/orqen/releases/download/${VERSION}/orqen_Linux_arm64.tar.gz" | tar -xz
sudo mv orqen /usr/local/bin/
sudo chmod +x /usr/local/bin/orqen
```

#### macOS (Apple Silicon / M1 / M2 / M3)

```bash
VERSION=$(curl -s https://api.github.com/repos/soham1105-tech/orqen/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
curl -sSL "https://github.com/soham1105-tech/orqen/releases/download/${VERSION}/orqen_Darwin_arm64.tar.gz" | tar -xz
sudo mv orqen /usr/local/bin/
sudo chmod +x /usr/local/bin/orqen
```

#### macOS (Intel)

```bash
VERSION=$(curl -s https://api.github.com/repos/soham1105-tech/orqen/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
curl -sSL "https://github.com/soham1105-tech/orqen/releases/download/${VERSION}/orqen_Darwin_x86_64.tar.gz" | tar -xz
sudo mv orqen /usr/local/bin/
sudo chmod +x /usr/local/bin/orqen
```

#### Windows (PowerShell)

1. Download `orqen_Windows_x86_64.zip` from [GitHub Releases](https://github.com/soham1105-tech/orqen/releases).
2. Extract `orqen.exe` to a directory in your `PATH` (e.g., `C:\Program Files\orqen\` or `$HOME\bin`).
3. Ensure the folder is included in your User or System `PATH` environment variable.

---

### Method 2: Via `go install`

If you have Go 1.27+ installed, you can install directly into your `$GOPATH/bin`:

```bash
go install -trimpath -ldflags "-s -w" github.com/soham1105-tech/orqen@latest
```

Ensure `$GOPATH/bin` (or `$(go env GOPATH)/bin`) is in your system `PATH`:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

---

### Method 3: From Source (Recommended for Contributors)

Clone the repository and use `make`:

```bash
git clone https://github.com/soham1105-tech/orqen.git
cd orqen

# Build local binary
make build

# Install to $(go env GOPATH)/bin
make install
```

---

### Method 4: Automated Shell Installer

On Unix-like systems (Linux and macOS):

```bash
curl -fsSL https://raw.githubusercontent.com/soham1105-tech/orqen/main/install.sh | bash
```

The script clones the latest source, builds it with your local Go compiler, and installs the binary to `~/.local/bin/orqen`.

---

### Method 5: Docker Container

You can also run ORQEN within a lightweight container:

```bash
docker build -f Dockerfile.release -t orqen:latest .
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock orqen:latest --help
```

---

## Verifying the Installation

After installing, verify the installation by checking the version and running preflight diagnostics:

```bash
# Verify binary is executable and check version
orqen --version

# Run the system diagnostic health check
orqen doctor
```

`orqen doctor` inspects your environment for:
1. Docker daemon connectivity and Compose v2 availability.
2. Git executable presence.
3. Configured LLM credentials and endpoint reachability.
4. Tunneling binaries (`cloudflared` or `ssh`) for ephemeral public sharing.

---

## Shell Autocompletion

ORQEN can generate command autocompletion scripts for popular shells.

### Bash

```bash
# Current session
source <(orqen completion bash)

# Persist across sessions
orqen completion bash > /etc/bash_completion.d/orqen
```

### Zsh

```bash
# If completion is enabled in your ~/.zshrc:
orqen completion zsh > "${fpath[1]}/_orqen"
```

### Fish

```bash
orqen completion fish > ~/.config/fish/completions/orqen.fish
```

### PowerShell (Windows)

```powershell
orqen.exe completion powershell | Out-String | Invoke-Expression

# To persist in your profile:
Add-Content $PROFILE "`n& orqen.exe completion powershell | Out-String | Invoke-Expression"
```

---

## Initial Configuration

Run the interactive setup wizard to configure your preferred LLM provider and credentials:

```bash
orqen setup
```

The wizard prompts for:
- **LLM Provider**: Gemini, OpenAI, Anthropic, Groq, or Ollama
- **Model**: Default or customized model identifier
- **API Key**: Securely masked input with immediate live connection verification
- **GitHub PAT** (Optional): Personal access token for cloning private repositories

See the [Configuration Guide](CONFIGURATION.md) for full details on environment variables and priority cascades.
