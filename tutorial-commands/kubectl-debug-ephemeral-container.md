runs only if 'spec/template/spec/securityContext/runAsNonRoot: false' was set
`kubectl debug -ti pods/personal-iot-controller-manager-6f7f7bc9d7-rs552 --image=busybox:1.28 --target=manager`

with copy target pod
`k debug -ti pods/personal-iot-controller-manager-6f7f7bc9d7-p5b42 --image=busybox:1.28 --share-processes --copy-to=iot-controller-debug`
to resume use 
`kubectl attach iot-controller-debug -c debugger-v956l -i -t`