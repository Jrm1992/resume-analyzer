# Resume Analyzer

Local-first web app: score a PDF resume against a job description and get an AI-suggested rewrite. Uses Ollama for inference — no cloud LLMs.

## Quickstart

1. Install and run Ollama: https://ollama.com
2. Pull a model: `ollama pull llama3.1:8b`
3. Build and run:
   ```bash
   make run
   ```
4. Open http://localhost:8080

## Configuration

| Env var           | Default                   | Purpose                 |
| ----------------- | ------------------------- | ----------------------- |
| `PORT`            | `8080`                    | HTTP listen port        |
| `OLLAMA_URL`      | `http://localhost:11434`  | Ollama base URL         |
| `OLLAMA_MODEL`    | `llama3.1:8b`             | Ollama model tag        |
| `MAX_PDF_MB`      | `10`                      | Upload size limit       |
| `LLM_TIMEOUT_SEC` | `120`                     | Per-job inference cap   |
| `WORKERS`         | `2`                       | Concurrent workers      |
| `QUEUE_CAPACITY`  | `100`                     | Queue buffer            |
| `JOB_TTL_MIN`     | `60`                      | Cleanup threshold       |

## Development

```bash
make test         # unit tests
make test-race    # with -race
make cover        # coverage report
make docker       # build image
```

End-to-end test (needs Ollama running):

```bash
ollama pull qwen2:0.5b
OLLAMA_MODEL=qwen2:0.5b make test-integration
```

## Architecture

See `docs/superpowers/specs/2026-04-22-resume-analyzer-design.md`.
