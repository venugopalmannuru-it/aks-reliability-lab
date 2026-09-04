# Repository guidance for coding agents

## Scope

This is a portfolio lab, not evidence of a production deployment. Keep claims factual and distinguish tested local behavior from cloud-dependent examples.

## Design constraints

- Keep the Go service dependency-free unless a new dependency has a clear operational benefit.
- Preserve `/healthz`, `/readyz`, and `/metrics` contracts.
- Never commit credentials or concrete secret values. Reference an externally managed Kubernetes Secret.
- Keep the container non-root and compatible with a read-only root filesystem.
- Keep raw manifests and Helm behavior aligned when changing runtime or reliability settings.
- Use least-privilege GitHub Actions permissions.

## Required checks

Run `make validate` after application or deployment changes. Run `make docker-build` and exercise `/readyz` when Docker is available. If a required tool is unavailable, name the skipped command in the handoff.

## Documentation

Update the README whenever commands, values, endpoints, or operational behavior change. Do not describe an external deployment as tested without evidence.
