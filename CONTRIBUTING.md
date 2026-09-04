# Contributing

Thank you for improving the lab. Open an issue before a large architectural change so its learning objective stays clear.

## Development workflow

1. Create a short-lived branch from `main`.
2. Make one coherent change and add or update tests.
3. Run `make validate`.
4. If Docker is available, run `make docker-build` and verify `/healthz`, `/readyz`, and `/metrics`.
5. Update documentation for any user-visible or operational change.
6. Open a pull request explaining the reliability or security trade-off.

Use Conventional Commit-style messages when practical, for example `feat: add latency histogram` or `fix: align probe timeout with server deadline`.

Never add real tokens, kubeconfigs, cloud account identifiers, customer information, or employer-specific configuration. Report suspected vulnerabilities privately to the repository owner rather than opening a public issue with exploit details.
