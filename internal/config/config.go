// Package config resolves ORQEN's LLM/provider settings from, in order of
// precedence: OS environment, dotenv files, and ~/.orqen/config.json. It also
// owns the application's directory layout.
//
// Environment variables are ORQEN_* (ORQEN_HOME, ORQEN_LLM_PROVIDER,
// ORQEN_LLM_MODEL, ORQEN_LLM_API_KEY, ORQEN_LLM_BASE_URL, ORQEN_ENV_FILE).
// The legacy SPOOL_* names are still honoured as a fallback so existing
// setups keep working across the rename.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/monojitgoswami69/spool/internal/domain"
	"github.com/monojitgoswami69/spool/internal/safefs"
)

// Config is the persisted global configuration (never committed; 0600).
type Config struct {
	LLMProvider string `json:"llm_provider,omitempty"`
	LLMModel    string `json:"llm_model,omitempty"`
	LLMAPIKey   string `json:"llm_api_key,omitempty"`
	LLMBaseURL  string `json:"llm_base_url,omitempty"`
	GitHubPAT   string `json:"github_pat,omitempty"`
}

// Resolved is a fully-resolved provider configuration ready for the LLM client.
type Resolved struct {
	Provider string
	Model    string
	APIKey   string
	BaseURL  string
	Source   string
}

// Home returns the ORQEN home directory, honouring $ORQEN_HOME (with
// $SPOOL_HOME as a legacy fallback).
func Home() (string, error) {
	if h := strings.TrimSpace(os.Getenv("ORQEN_HOME")); h != "" {
		return filepath.Abs(h)
	}
	if h := strings.TrimSpace(os.Getenv("SPOOL_HOME")); h != "" {
		return filepath.Abs(h)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", domain.E(domain.KindConfig, "config.Home", "cannot determine home directory", err)
	}
	return filepath.Join(home, ".orqen"), nil
}

// ConfigPath returns the path to ~/.orqen/config.json.
func ConfigPath() (string, error) {
	h, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, "config.json"), nil
}

// ReposDir is where cloned repositories are kept.
func ReposDir() (string, error) {
	h, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, "repos"), nil
}

// ProjectsDir is where per-project generated manifests and state live.
func ProjectsDir() (string, error) {
	h, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, "projects"), nil
}

// ProjectDir returns the work directory for a project id.
func ProjectDir(id string) (string, error) {
	if !ValidProjectID(id) {
		return "", domain.E(domain.KindUnsafePath, "config.ProjectDir", "invalid project id "+id, nil)
	}
	base, err := ProjectsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, id), nil
}

// ValidProjectID rejects ids that could escape the projects directory.
func ValidProjectID(id string) bool {
	if id == "" || id == "." || id == ".." {
		return false
	}
	if strings.ContainsAny(id, `/\`) || strings.Contains(id, "..") {
		return false
	}
	for _, r := range id {
		if !(r == '-' || r == '_' || r == '.' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			return false
		}
	}
	return true
}

// Exists reports whether a config file is present.
func Exists() bool {
	p, err := ConfigPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(p)
	return err == nil
}

// LoadOptional returns the config, or an empty config on any error. It never
// fails, so it is safe on code paths that can operate without an LLM.
func LoadOptional() *Config {
	c, err := load(false)
	if err != nil || c == nil {
		return &Config{}
	}
	return c
}

// Load strictly loads and validates config.json.
func Load() (*Config, error) {
	return load(true)
}

func load(strict bool) (*Config, error) {
	p, err := ConfigPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			if strict {
				return nil, domain.E(domain.KindConfig, "config.Load",
					fmt.Sprintf("config not found at %s (run `orqen doctor` or set GEMINI_API_KEY)", p), err)
			}
			return nil, nil
		}
		return nil, domain.E(domain.KindConfig, "config.Load", "read config", err)
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, domain.E(domain.KindConfig, "config.Load", "parse config JSON", err)
	}
	if strict {
		if err := c.Validate(); err != nil {
			return nil, err
		}
	}
	return &c, nil
}

// Validate checks that the provider/model/key triple is usable.
func (c *Config) Validate() error {
	if strings.TrimSpace(c.LLMProvider) == "" {
		return domain.E(domain.KindConfig, "config.Validate", "llm_provider is empty", nil)
	}
	if strings.TrimSpace(c.LLMModel) == "" {
		return domain.E(domain.KindConfig, "config.Validate", "llm_model is empty", nil)
	}
	if NormalizeProvider(c.LLMProvider) != "ollama" && strings.TrimSpace(c.LLMAPIKey) == "" {
		return domain.E(domain.KindConfig, "config.Validate", "llm_api_key is empty", nil)
	}
	return nil
}

// Save writes the config atomically with 0600 permissions.
func Save(c *Config) error {
	if c == nil {
		return domain.E(domain.KindConfig, "config.Save", "nil config", nil)
	}
	p, err := ConfigPath()
	if err != nil {
		return err
	}
	if err := safefs.EnsureDir(filepath.Dir(p), 0o700); err != nil {
		return domain.E(domain.KindIO, "config.Save", "create home", err)
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return domain.E(domain.KindConfig, "config.Save", "marshal", err)
	}
	data = append(data, '\n')
	if err := safefs.WriteFileAtomic(p, data, 0o600); err != nil {
		return domain.E(domain.KindIO, "config.Save", "write", err)
	}
	return nil
}

// envFiles returns the dotenv files ORQEN will consult for runtime settings.
func envFiles() []string {
	var files []string
	if h, err := Home(); err == nil {
		files = append(files, filepath.Join(h, ".env"))
	}
	explicit := firstNonEmpty(
		os.Getenv("ORQEN_ENV_FILE"),
		os.Getenv("SPOOL_ENV_FILE"), // legacy name
	)
	if explicit != "" {
		files = append(files, explicit)
	}
	if wd, err := os.Getwd(); err == nil {
		files = append(files, filepath.Join(wd, ".env"))
	}
	return files
}

// lookupEnv resolves the first non-empty value among OS env / dotenv lookups
// for the given keys, in order. It backs the ORQEN_* variables with their
// legacy SPOOL_* equivalents.
func lookupEnv(dotEnv map[string]string, keys ...string) string {
	for _, key := range keys {
		if v, ok := os.LookupEnv(key); ok && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
		if v, ok := dotEnv[key]; ok && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func providerEnvKey(provider string) string {
	switch NormalizeProvider(provider) {
	case "gemini":
		return "GEMINI_API_KEY"
	case "openai":
		return "OPENAI_API_KEY"
	case "anthropic":
		return "ANTHROPIC_API_KEY"
	case "groq":
		return "GROQ_API_KEY"
	default:
		return ""
	}
}

// DefaultModel returns the provider's default model.
func DefaultModel(provider string) string {
	switch NormalizeProvider(provider) {
	case "gemini":
		return "gemini-3.5-flash-lite"
	case "openai":
		return "gpt-4o-mini"
	case "anthropic":
		return "claude-3-5-sonnet-20241022"
	case "groq":
		return "llama-3.3-70b-versatile"
	case "ollama":
		return "llama3.1"
	default:
		return ""
	}
}

// NormalizeProvider canonicalizes provider aliases.
func NormalizeProvider(p string) string {
	switch strings.ToLower(strings.TrimSpace(p)) {
	case "gemini", "google", "google-gemini", "googleai":
		return "gemini"
	case "openai", "open-ai":
		return "openai"
	case "anthropic", "claude":
		return "anthropic"
	case "groq":
		return "groq"
	case "ollama", "local":
		return "ollama"
	default:
		return strings.ToLower(strings.TrimSpace(p))
	}
}

// Resolve builds a fully-resolved provider configuration, merging OS env,
// dotenv, and config.json. It returns domain.ErrNotConfigured when no usable
// provider/key can be found.
func Resolve() (*Resolved, error) {
	cfg := LoadOptional()
	dotEnv := LoadDotEnv(envFiles()...)

	provider := NormalizeProvider(firstNonEmpty(
		lookupEnv(dotEnv, "ORQEN_LLM_PROVIDER", "SPOOL_LLM_PROVIDER"),
		cfg.LLMProvider,
	))
	if provider == "" {
		// Infer from whichever key is present.
		for _, p := range []string{"gemini", "openai", "anthropic", "groq"} {
			if lookupEnv(dotEnv, providerEnvKey(p)) != "" {
				provider = p
				break
			}
		}
	}
	if provider == "" {
		return nil, domain.E(domain.KindConfig, "config.Resolve", "no LLM provider configured", domain.ErrNotConfigured)
	}

	model := firstNonEmpty(lookupEnv(dotEnv, "ORQEN_LLM_MODEL", "SPOOL_LLM_MODEL"), cfg.LLMModel, DefaultModel(provider))

	source := "config"
	apiKey := ""
	if pk := providerEnvKey(provider); pk != "" {
		if v := lookupEnv(dotEnv, pk); v != "" {
			apiKey, source = v, "env:"+pk
		}
	}
	if apiKey == "" {
		if v := firstNonEmpty(
			lookupEnv(dotEnv, "ORQEN_LLM_API_KEY", "SPOOL_LLM_API_KEY"),
			lookupEnv(dotEnv, "LLM_API_KEY"),
		); v != "" {
			apiKey, source = v, "env:generic"
		}
	}
	if apiKey == "" && cfg.LLMAPIKey != "" {
		apiKey, source = cfg.LLMAPIKey, "config"
	}

	baseURL := firstNonEmpty(lookupEnv(dotEnv, "ORQEN_LLM_BASE_URL", "SPOOL_LLM_BASE_URL"), cfg.LLMBaseURL)

	if provider != "ollama" && apiKey == "" {
		return nil, domain.E(domain.KindConfig, "config.Resolve",
			fmt.Sprintf("provider %q has no API key (set %s or run `orqen setup`)", provider, providerEnvKey(provider)),
			domain.ErrNotConfigured)
	}

	return &Resolved{Provider: provider, Model: model, APIKey: apiKey, BaseURL: baseURL, Source: source}, nil
}

// GitHubToken resolves a PAT from OS env / dotenv / config.json.
func GitHubToken() string {
	cfg := LoadOptional()
	dotEnv := LoadDotEnv(envFiles()...)
	return firstNonEmpty(lookupEnv(dotEnv, "GITHUB_PAT"), lookupEnv(dotEnv, "GH_TOKEN"), cfg.GitHubPAT)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// Dir returns a resolved subdirectory of the ORQEN home, creating it 0700.
func Dir(parts ...string) (string, error) {
	h, err := Home()
	if err != nil {
		return "", err
	}
	p := filepath.Join(append([]string{h}, parts...)...)
	if err := safefs.EnsureDir(p, 0o700); err != nil {
		return "", err
	}
	return p, nil
}

// DotEnvFiles returns the dotenv files ORQEN consults at runtime.
func DotEnvFiles() []string { return envFiles() }

// LoadEnvIntoProcess loads dotenv files into the process environment without
// overriding variables that are already set. This makes .env work for every
// command, not just those that call Resolve.
func LoadEnvIntoProcess() {
	for k, v := range LoadDotEnv(envFiles()...) {
		if strings.TrimSpace(os.Getenv(k)) == "" {
			_ = os.Setenv(k, v)
		}
	}
}
