# orders-svc

Orders API (in-memory); validates user IDs against users-svc. Part of the Backstage IDP lab.

## Run locally

```sh
go test ./...
go run ./cmd/server          # PORT=8080 by default
curl localhost:8080/healthz
```

## Layout

```
cmd/server/            main: config from env, routes, graceful shutdown
internal/              handlers, in-memory store, unit tests
Dockerfile             golang builder (cross-compiles) -> distroless static, non-root
charts/orders-svc/     Helm chart (default values); per-env values live in the gitops repo
.github/workflows/ci.yml
catalog-info.yaml, mkdocs.yml, docs/
```

## CI

- **PR:** `go vet`, `go test -race`, golangci-lint, `helm lint` / `helm template`, Docker build (no push).
- **Push to main:** the same, then
  1. push `ghcr.io/bhumi46/orders-svc:<git-sha>` (+ `:main`), linux/amd64 + linux/arm64;
  2. push the chart to `oci://ghcr.io/bhumi46/charts/orders-svc` if its `version` is new (bump `Chart.yaml` to publish chart changes);
  3. set `image.tag: <git-sha>` in `gitops/envs/dev/orders-svc/values.yaml` using the `GITOPS_TOKEN` secret. Staging is promoted by PR.

Charts are distributed as **OCI artifacts in GHCR**; Argo CD pulls the chart from there and the values from `gitops`.

## API

| Method | Path | |
|---|---|---|
| GET | `/healthz`, `/readyz` | probes |
| GET | `/api/orders` | list orders |
| GET | `/api/orders/{id}` | one order, 404 if missing |
| POST | `/api/orders` | `{"user_id","item","quantity"}` → 201; 422 if users-svc says the user doesn't exist; 502 if users-svc is unreachable |

Env: `PORT` (8080), `LOG_LEVEL`, `USERS_SVC_URL` (default `http://users-svc`, i.e. users-svc in the same namespace).

## Pinned versions

| What | Version |
|---|---|
| Go | 1.27.4 (`go.mod`) |
| Builder image | `golang:1.27.4-alpine@sha256:70bd7ddea8b6cbe5d6943cde279d6b67feb10cc924fb9100d1d8757c1575b280` |
| Runtime image | `gcr.io/distroless/static-debian13:nonroot@sha256:aafe993665bf6200db38a02c11a9351e782a84b6a45e7f33b6695638c88e3931` (uid 65532) |
| golangci-lint | v2.15.0 |
| Helm (CI) | v4.1.0 |
| yq (CI) | v4.54.3 |
| `actions/checkout` | v7.0.1 |
| `actions/setup-go` | v7.1.0 |
| `golangci/golangci-lint-action` | v9.3.0 |
| `azure/setup-helm` | v5.1.0 |
| `docker/setup-buildx-action` | v4.4.1 |
| `docker/login-action` | v4.8.0 |
| `docker/build-push-action` | v7.4.0 |
| Runner | `ubuntu-24.04` |
| Chart version | see `charts/*/Chart.yaml` (0.1.0) |
