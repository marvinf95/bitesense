# Deployment to Pi-Homelab (k3s)

This guide assumes the [Pi-Homelab](../../pi-homelab/PLAN.md) is already up: k3s installed, Traefik enabled, Tailscale Kubernetes Operator configured, Sealed Secrets controller running.

## 1. Build images (on the dev notebook)

```bash
# Backend
docker buildx build \
  --platform linux/arm64 \
  -t ghcr.io/marvinf95/bitesense-backend:v0.1.0 \
  --push ./backend

# Frontend
docker buildx build \
  --platform linux/arm64 \
  -t ghcr.io/marvinf95/bitesense-frontend:v0.1.0 \
  --push ./frontend
```

(Or let GitHub Actions do it: `.github/workflows/build-images.yml` triggers on tag push.)

## 2. Seal real secrets

```bash
# Pull the cluster's public key once.
kubeseal --fetch-cert > /tmp/sealed-pubkey.pem

kubectl create secret generic bitesense-secrets \
  --namespace bitesense \
  --from-literal=BITESENSE_JWT_SECRET="$(openssl rand -hex 32)" \
  --from-literal=GEMINI_API_KEY="$GEMINI_API_KEY" \
  --from-literal=ANTHROPIC_API_KEY="$ANTHROPIC_API_KEY" \
  --dry-run=client -o yaml \
| kubeseal --cert /tmp/sealed-pubkey.pem -o yaml > k8s/overlays/prod/sealed-secrets.yaml
```

Then drop the example file from base by patching the prod overlay (see `k8s/overlays/prod/kustomization.yaml`).

## 3. Apply the overlay

```bash
kubectl apply -k k8s/overlays/prod
kubectl -n bitesense rollout status deploy/bitesense-backend
kubectl -n bitesense rollout status deploy/bitesense-frontend
```

## 4. Verify

```bash
# Healthy backend
kubectl -n bitesense port-forward svc/bitesense-backend 8080:80 &
curl http://localhost:8080/readyz

# Reachable over Tailscale (via Tailscale K8s operator)
tailscale status | grep bitesense
curl https://bitesense.<your-tailnet>.ts.net/livez
```

## 5. Backups

The Pi-Homelab restic backup job already covers `/var/lib/rancher/k3s/storage/`. Confirm the `bitesense-data` PVC subdirectory is included in the next nightly run.

## Rollback

```bash
kubectl -n bitesense rollout undo deploy/bitesense-backend
kubectl -n bitesense rollout undo deploy/bitesense-frontend
```
