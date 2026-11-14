# ORQEN Configuration Guide

This guide details how ORQEN discovers, validates, and manages configuration settings across system environments, project workspaces, and runtime containers.

---

## Configuration Precedence

ORQEN resolves configuration settings using a strict three-tier precedence cascade:

```
┌──────────────────────────────────────────────┐
│  1. Operating System Environment Variables   │  (Highest Priority)
├──────────────────────────────────────────────┤
│  2. Dotenv Files (.env)                      │
│     - $ORQEN_ENV_FILE                        │
│     - ./.env                                 │
│     - ~/.orqen/.env                          │
├──────────────────────────────────────────────┤
│  3. Persistent Configuration File            │  (Lowest Priority)
│     - ~/.orqen/config.json                   │
└──────────────────────────────────────────────┘
```

If a setting is defined in multiple layers, the higher layer always wins.

---

## Supported Environment Variables

| Variable | Default | Purpose |
|----------|---------|---------|
| `ORQEN_HOME` | `~/.orqen` | Root directory for state, shallow clones, and project metadata |
| `ORQEN_LLM_PROVIDER` | `gemini` | Chosen AI provider: `gemini`, `openai`, `anthropic`, `groq`, or `ollama` |
| `ORQEN_LLM_MODEL` | Provider default | Model identifier (e.g. `gemini-3.5-flash-lite`, `gpt-4o-mini`, `claude-3-5-sonnet-20241022`) |
| `ORQEN_LLM_API_KEY` | *(None)* | Universal fallback API key for the active provider |
| `GEMINI_API_KEY` | *(None)* | Google Gemini API key |
| `OPENAI_API_KEY` | *(None)* | OpenAI API key |
| `ANTHROPIC_API_KEY` | *(None)* | Anthropic API key |
| `GROQ_API_KEY` | *(None)* | Groq API key |
| `ORQEN_LLM_BASE_URL` | *(None)* | Custom endpoint base URL (required for Ollama or private proxies) |
| `GITHUB_PAT` / `GH_TOKEN` | *(None)* | Personal Access Token for cloning private repositories via HTTPS |
| `ORQEN_ENV_FILE` | *(None)* | Path to an explicit dotenv file to load |

---

## Configuration File (`~/.orqen/config.json`)

When you run `orqen setup` or `orqen config set`, settings are persisted in `~/.orqen/config.json` with strict POSIX `0600` permissions (user read/write only).

### Schema

```json
{
  "provider": "gemini",
  "model": "gemini-3.5-flash-lite",
  "api_key": "AIzaSy...",
  "base_url": "",
  "github_pat": "ghp_..."
}
```

### Scripting with `orqen config`

```bash
# Display resolved configuration with credentials safely masked
orqen config show

# Print configuration file location
orqen config path

# Set specific configuration keys programmatically
orqen config set provider openai
orqen config set model gpt-4o-mini
orqen config set api-key sk-...
```

---

## Managing Project Environment Variables

Each project onboarded with ORQEN maintains its own isolated environment variables:

1. **Discovered Variables**: Detected during repository scanning from code, `.env.example`, and framework settings.
2. **Generated Variables**: Cryptographically secure random strings for secrets (`POSTGRES_PASSWORD`, `SECRET_KEY`, `JWT_SECRET`, etc.).
3. **Inferred Service URLs**: Dynamic connection strings linking services to databases and caches (e.g. `DATABASE_URL=postgres://user:pass@postgres:5432/db`).
4. **User Overrides**: Stored in `~/.orqen/projects/<project-id>/env-overrides.json`.

### CLI Commands for Project Variables

```bash
# List all variables for a project
orqen env my-project

# Set or update an environment variable (automatically triggers container recreation)
orqen env my-project set STRIPE_API_KEY=sk_test_12345

# Remove an override
orqen env my-project unset STRIPE_API_KEY

# Open interactive terminal editor with masked inputs
orqen env my-project config
```

---

## Security Best Practices

1. **Never commit `.env` files**: Ensure `.env` is listed in your `.gitignore`.
2. **Review credentials**: Use `orqen config show` to confirm credentials are read from expected sources.
3. **Use granular GitHub PATs**: When configuring `GITHUB_PAT`, grant only read access to repository contents.
4. **Audit permissions**: Ensure `~/.orqen` and all files inside maintain restrictive permissions (`0700` for directories, `0600` for files).
