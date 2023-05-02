[install metallb with help of the documentation](https://kind.sigs.k8s.io/docs/user/loadbalancer/)
- `kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.13.7/config/manifests/metallb-native.yaml`
- dig for ip address range `docker network inspect -f '{{.IPAM.Config}}' kind`
- create metallb-config.yaml, see `hack/metallb-config.yaml`
- `kubectl apply -f hack/metallb-config.yaml`
- ensure the volumes are set correctly:
  - the local host path volumes are set in `kind-cluster-config.yaml`, the `contiainerPath` matches the path in the volume
  - check if the `containerPath` exists in the node, e.g. `docker exec -ti personal-iot-worker ls /var/mqttdata`
  - check if the pv with storage class manual is declared in mosquitto.yaml
- deploy mosquitto `bin/kustomize build config/mosquitto | kubectl apply -f -`

Test the deployment:
- `k port-forward -n mqtt mosquitto-56b8cd6f69-d6lj4 1883`
- show connected clients `mqtt sub -h localhost -t '$SYS/broker/clients/connected' -u user -pw password`
- subscribe a topic `mqtt sub -h localhost -t 'cmnd/gosund_p1_1_12FCA5/POWER3' -u user -pw password`
- send messages `mqtt pub -h localhost -t cmnd/gosund_p1_1_12FCA5/POWER3 -m "OFF" -u user -pw password`