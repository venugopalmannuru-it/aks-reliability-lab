# Reference assessment and clean-room boundary

Assessment date: 2026-09-04

## Reference selected

The supplied portfolio URL led to [`NagellaSivaKumar/javaproject-DevOpsLAB`](https://github.com/NagellaSivaKumar/javaproject-DevOpsLAB), a public fork of [`emirhandogandemir/java-DevOps-Project`](https://github.com/emirhandogandemir/java-DevOps-Project).

The repository describes a Spring Boot application containerized with Docker, integrated through Jenkins, and delivered with Helm and Argo CD. Its visible root contains Maven/Spring source, a Dockerfile, a Jenkinsfile, and a `helm-argocd` directory.

## Licence finding

GitHub did not show a licence file or detected licence for the fork at assessment time. It is therefore treated as copyrighted. No source code, prose, diagrams, screenshots, branding, assets, or Git history were copied.

## Safe conceptual inspiration

Only the generic delivery chain—small application, container image, CI automation, and Helm-based Kubernetes deployment—was retained as inspiration. Those are common DevOps patterns rather than repository-specific expression.

## Substantial differences

This project is a clean-room Go implementation with standard-library tests, explicit liveness/readiness behavior, graceful termination, Prometheus metrics, structured logs, a scratch non-root image, Docker Compose, raw Kubernetes resources, HPA behavior, a PodDisruptionBudget, NetworkPolicy, security contexts, separate environment values, GitHub Actions, SBOM/provenance generation, vulnerability/secret/misconfiguration scanning, and original operational documentation.

The project does not use Java, Maven, Spring Boot, Jenkins, Argo CD definitions, the reference chart, or the reference repository structure.
