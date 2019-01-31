# Examples

Here you can find examples on how to run the exporter as a sidecar alongside an EMQ container.

## Prerequisites

* A running kubernetes cluster (e.g. minikube)
* A prometheus server running and configured

## Usage

To install the manifests in the cluster run:

```bash
kubectl apply -f manifests/
```

This will create the following resources:
* StatefulSet with 2 EMQ brokers, configured to auto discover via the kubernetes api
* 2 Services - a headless one for the statefulset and another for the dashboard, api etc.
* Secret to hold the admin credentials and erlang node cookie
* RBAC service account, role and role binding to allow discovery

Open up prometheus and see the metrics flow!

### Notes

* All metrics exported to prometheus will be prefixed with `emq_`
* Some fields specify the `k8s` namespace in which `emq` was deployed (some env vars and the `rolebinding`). 
If you're installing in another namespace, make sure to change those
* The image used can be found in [docker hub](https://hub.docker.com/r/emqx/emqx). Feel free to use which ever image you'd like
* If you change the metrics port (e.g passing `--web.listen-address=PORT`), don't forget to change to annotation and port in the container spec
