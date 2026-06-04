# Testing & verification

## Automated checks

| Command | Purpose |
|---------|---------|
| `cd backend && go test -race ./...` | Unit tests (auth hashing, JWT, Fisher's exact) |
| `cd backend && golangci-lint run` | Static analysis (errcheck, gosec, staticcheck) |
| `cd frontend && dart analyze --fatal-infos` | Dart analyzer (strict casts) |
| `cd frontend && dart format --output=none --set-exit-if-changed .` | Format check |
| `cd frontend && flutter test --coverage` | Widget tests (incl. i18n switch) |
| `trivy fs .` | Dependency vulnerability scan |
| `trivy image ghcr.io/marvinf95/bitesense-backend:<tag>` | Image scan |

All of the above run in CI on every push (see `.github/workflows/ci.yml`).

## End-to-end acceptance (manual)

After deploying to a fresh environment, walk through:

1. **Register** new account → JWT returned, lands on `/meals`.
2. **Log a meal via text** with two items, each with a couple of allergen tags → appears in list with subtitle showing items.
3. **Take a photo** → vision pipeline returns parsed items → meal lands in list with `source = image`.
4. **Scan a barcode** of any packaged food → meal lands with `source = barcode` and OFF-canonical name.
5. **Log 3 symptoms** of the same type over different days → analytics tab shows at least a `WEAK_SIGNAL` after enough cross-pairings.
6. **Switch language** (Settings → Language → Deutsch) → all visible strings localised, including the date format.
7. **Export PDF** for the last 7 days → PDF opens with three sections; correlation table includes the suspect from step 5.
8. **Refresh token rotation**: wait > 15 minutes, perform any request → silently rotated, request succeeds.
9. **Delete account** from Settings → returns to login, attempting to log in fails with 401.
10. **Network policy**: from inside the cluster, `kubectl exec` into the frontend pod and try to reach `bitesense-backend` over a non-allowed port → connection refused.

## Sample fixtures

`backend/testdata/` is reserved for golden-file vision fixtures (intentionally empty in v0.1; real fixtures are added when integration tests land).
