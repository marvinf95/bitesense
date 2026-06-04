# Deployment

Two deployment flows are supported:

1. [Pi-Homelab](#pi-homelab-flow-on-pi-buildkit) — Marvin's actual production setup. On-Pi BuildKit, no external registry, Tailscale Ingress.
2. [Portable](#portable-flow-ghcr--traefik) — anyone else who wants to host BiteSense on a regular k8s cluster with an external image registry.

---

## Pi-Homelab flow (on-Pi BuildKit)

Assumptions about the target cluster:

- k3s with kube-router NetworkPolicy enforcement (`v1.31.4+k3s1` confirmed)
- Sealed Secrets controller in namespace `sealed-secrets`
- Tailscale Kubernetes Operator (creates `ingressClassName: tailscale` ingresses → `<host>.<tailnet>.ts.net` with auto Let's Encrypt cert)
- BuildKit running on the Pi exposed via the `pibuild` helper (writes into the k3s containerd `k8s.io` namespace)
- LAN SSH to the Pi with `ssh marvin@pi-homelab.fritz.box`, NOPASSWD sudo

If any of these aren't in place, see the [pi-homelab repo](../../pi-homelab/) for setup notes.

### 1. Seal the secrets (once per cluster)

```bash
kubeseal --fetch-cert \
  --controller-namespace sealed-secrets \
  --controller-name sealed-secrets-controller \
  > /tmp/sealed-pubkey.pem

kubectl create secret generic bitesense-secrets \
  --namespace bitesense \
  --from-literal=BITESENSE_JWT_SECRET="$(openssl rand -hex 32)" \
  --from-literal=GEMINI_API_KEY="$GEMINI_API_KEY" \
  --from-literal=ANTHROPIC_API_KEY="$ANTHROPIC_API_KEY" \
  --dry-run=client -o yaml \
| kubeseal --cert /tmp/sealed-pubkey.pem -o yaml \
  > k8s/overlays/prod/sealed-secrets.yaml
```

Then point the overlay at the real SealedSecret by uncommenting the `patches` block in `k8s/overlays/prod/kustomization.yaml`.

### 2. Build + deploy via Makefile

The repo ships a `Makefile` that wraps the entire flow. From the dev notebook:

```bash
# Builds Flutter web on the notebook → rsync to Pi → pibuild backend + frontend →
# kustomize edit set image → kubectl apply -k → wait for rollout.
make pi-deploy
```

The default values are tuned for Marvin's setup; override anything you need:

```bash
make pi-deploy \
  PI_HOST=pi-homelab.fritz.box \
  PI_PATH=/home/marvin/bitesense \
  TAG=$(git rev-parse --short HEAD) \
  API_URL=https://bitesense.tailb969ce.ts.net \
  KUBE_OVERLAY=prod
```

What the targets do under the hood:

| Target | Action |
|--------|--------|
| `make frontend-build-web` | `flutter build web --release` → `frontend/web/` (consumed by the slim Dockerfile) |
| `make pi-sync` | `rsync` backend/, frontend/ (incl. the bundled `web/`), k8s/ to the Pi |
| `make pi-build-backend` | SSH + `pibuild backend bitesense-backend:$TAG` |
| `make pi-build-frontend` | SSH + `pibuild frontend bitesense-frontend:$TAG` |
| `make pi-apply` | SSH + `kustomize edit set image` + `kubectl apply -k` + `rollout status` |

### 3. Verify

```bash
# Cluster-side
ssh marvin@pi-homelab.fritz.box "kubectl -n bitesense get pods,svc,ingress"

# Tailnet-side
curl -fsSL https://bitesense.tailb969ce.ts.net/livez
```

The Pi's pod-crash alert (`/usr/local/bin/pi-pod-alert.sh`) automatically watches the new namespace and pings WhatsApp via OpenClaw on any `CrashLoopBackOff` / `ImagePullBackOff` / `OOMKilled`.

### 4. Rollback

```bash
ssh marvin@pi-homelab.fritz.box \
  "kubectl -n bitesense rollout undo deploy/bitesense-backend && \
   kubectl -n bitesense rollout undo deploy/bitesense-frontend"
```

### 5. Backups

The Pi's restic + rclone job (`/usr/local/bin/pi-backup.sh`, runs nightly via `pi-backup.timer`) already covers `/var/lib/rancher/k3s/storage` — the BiteSense PVC is included automatically. No app-specific configuration needed.

---

## Portable flow (GHCR + Traefik)

For external clusters that have:

- A real container registry (default: GHCR at `ghcr.io/marvinf95`)
- Traefik with `IngressRoute` CRDs
- Some certificate solution (cert-manager, ACME via Traefik, manual TLS Secret)

### Build + push images

```bash
docker login ghcr.io                     # once
make image-build-portable TAG=v0.1.0     # multi-arch buildx → push to GHCR
```

### Apply with the portable overlay

The `overlays/portable` overlay swaps:

- `Ingress (ingressClassName: tailscale)` → Traefik `IngressRoute` (edit the host in `ingressroute.yaml`)
- `imagePullPolicy: Never` → `IfNotPresent`
- Local image names → fully-qualified GHCR names

```bash
# Seal your secrets the same way as above, then:
kubectl apply -k k8s/overlays/portable
```

CI (GitHub Actions) handles the image build automatically on push to `main` and tag pushes — see [.github/workflows/build-images.yml](../.github/workflows/build-images.yml).
