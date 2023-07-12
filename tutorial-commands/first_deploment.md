#### first deployment
- generate manifests and boilerplate code `make manifests generate`
- build CRDs and apply to cluster `make install`
- build docker image and push it to the docker repo, load docker image to cluster `make docker-build docker-push kind-load`
- install cert-manager in cluster `D`, be aware the cert-manager will take a moment to start
- deploy the application `make deploy`

