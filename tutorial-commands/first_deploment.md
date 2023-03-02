#### first deployment
- `make manifests generate`
- `make install`
- `make docker-build`
- `kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.11.0/cert-manager.yaml `
- `make deploy`
- run in issuer error
- `kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.11.0/cert-manager.yaml`
- `make deploy` again

keep in mind, certmanagers start time could a bit lengthy
