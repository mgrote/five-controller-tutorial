#### init project
- `kubebuilder init --domain frup.org --owner "mgrote" --repo github.com/mgrote/personal-iot --component-config`

#### kubebuilder create objects for api
- `kubebuilder create api --group personal-iot --version v1alpha1 --kind Powerstrip`
- `kubebuilder create api --group personal-iot --version v1alpha1 --kind Poweroutlet`
- `kubebuilder create api --group personal-iot --version v1alpha1 --kind Location`

#### kubebuilder add custom config to controller config file
- `- kubebuilder create api --group personal-iot --version v1alpha1 --kind MQTTControllerConfig --resource --controller=false --make=false`

#### kubebuilder create webhook
- `kubebuilder create webhook --group personal-iot --version v1alpha1 --defaulting --programmatic-validation --kind Poweroutlet`

#### build boilerplate and manifests
- `make generate`
- `make manifests`
