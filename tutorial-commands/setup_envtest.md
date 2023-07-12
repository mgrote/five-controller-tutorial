#### setup envtest
- `make envtest` to download k8s api and etcd
- `bin/setup-envtest use 1.25.0`
- `cp -r ~/.local/share/kubebuilder-envtest/k8s bin/`
- set path to executables with `KUBEBUILDER_ASSETS` in `controllers/suite_test.go`
- check if your local gingko binary version fits the requested version in `go.mod`, if not upgrade   

error if ginkgo version are not fitting:
```
/usr/bin/ginkgo -v "--focus=Power strip controller"
Ginkgo detected a version mismatch between the Ginkgo CLI and the version of Ginkgo imported by your packages:
  Ginkgo CLI Version:
    2.8.0
  Mismatched package versions found:
    2.7.0 used by controllers
```