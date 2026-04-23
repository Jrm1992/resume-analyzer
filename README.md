# Resume Analyzer

Web app: score a PDF resume against a job description and get an AI-suggested rewrite. Talks to any **OpenAI-compatible `/v1/chat/completions` endpoint** — OpenAI, Anthropic (via their compat endpoint), OpenRouter, LiteLLM, Groq, etc.

## Quickstart

Set an API key, then build and run.

### OpenAI (default)

```bash
export LLM_API_KEY=sk-…
export LLM_MODEL=gpt-4o-mini            # or gpt-4o, gpt-4.1-mini, etc.
make run
```

### Anthropic (via OpenAI-compatible endpoint)

```bash
export LLM_BASE_URL=https://api.anthropic.com/v1
export LLM_API_KEY=sk-ant-…
export LLM_MODEL=claude-sonnet-4-5
make run
```

### OpenRouter (proxy to any provider)

```bash
export LLM_BASE_URL=https://openrouter.ai/api/v1
export LLM_API_KEY=sk-or-…
export LLM_MODEL=anthropic/claude-sonnet-4.5
make run
```

Then open http://localhost:8080.

## Configuration

| Env var           | Default                       | Required | Purpose                                          |
| ----------------- | ----------------------------- | -------- | ------------------------------------------------ |
| `LLM_API_KEY`     | —                             | **yes**  | Bearer token for the upstream API                |
| `LLM_BASE_URL`    | `https://api.openai.com/v1`   | no       | Chat Completions base (no `/chat/completions`)   |
| `LLM_MODEL`       | `gpt-4o-mini`                 | no       | Model identifier for the chosen provider         |
| `LLM_MAX_TOKENS`  | `4000`                        | no       | `max_tokens` in request body                     |
| `LLM_TIMEOUT_SEC` | `120`                         | no       | Per-job inference cap                            |
| `PORT`            | `8080`                        | no       | HTTP listen port                                 |
| `MAX_PDF_MB`      | `10`                          | no       | Upload size limit                                |
| `WORKERS`         | `2`                           | no       | Concurrent workers                               |
| `QUEUE_CAPACITY`  | `100`                         | no       | Queue buffer                                     |
| `JOB_TTL_MIN`     | `60`                          | no       | In-memory job cleanup threshold                  |

## Docker

```bash
LLM_API_KEY=sk-… docker compose up --build
```

Compose reads `LLM_API_KEY`, `LLM_BASE_URL`, `LLM_MODEL`, `LLM_MAX_TOKENS` from the host env.

## Development

```bash
make test         # unit tests
make test-race    # with -race
make cover        # coverage report
make docker       # build image
```

End-to-end test (hits your configured upstream and consumes API credits):

```bash
LLM_API_KEY=sk-… LLM_MODEL=gpt-4o-mini make test-integration
```

## Architecture

- Spec: `docs/superpowers/specs/2026-04-22-resume-analyzer-design.md`
- LLM swap (this iteration): `docs/superpowers/specs/2026-04-23-llm-providers-design.md`
