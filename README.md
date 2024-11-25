# imagine 🧞

[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/mikejoh)](https://artifacthub.io/packages/search?repo=mikejoh)

`imagine` - The simplest [`ImagePolicyWebhook`](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#imagepolicywebhook) webhook example you'll ever find!

## Introduction

In the Kubernetes API request journey, `imagine` plays a role in the validating admission phase, where the red dot in the diagram marks our position. This project showcases how simple it can be to integrate a custom webhook into Kubernetes and validate container images effectively.

<p align="center">
  <img src="https://github.com/user-attachments/assets/6ca9ad4e-87ee-4c2a-82c2-2becf4c1e58f" />
</p>

## Deploy in Kubernetes

_I'm kind of assuming that you've deployed a small cluster using `kubeadm`. I've done that on-top of KVM during development of this project, see [this](https://github.com/mikejoh/k8s-on-kvm) repository on how that was done._

1. Deploy `imagine` using Helm, this deploys the latest released version of `imagine`:

```bash
make deploy
```

2. Generate the needed certificate to be used with `imagine`, this will be used by `imagine` as the server-side certificate of the webhook:

```bash
make k8s-gen-cert
```

3. Copy the following files to a specific directory on all control-plane nodes:

```bash
files/admissionconfig.yaml
files/config.yaml
```

shall be copied to `/etc/kubernetes/imagine`, make sure you've created it first.

4. Now change the `kube-apiserver` static Pod manifest (in `/etc/kubernetes/manifests`), you'll need to add the following configuration:

* `ImagePolicyWebhook` to the `--enable-admission-plugins`flag
* `--admission-control-config-file=/etc/kubernetes/imagine/admissionconfig.yaml`

and:

```bash
...
spec:
  containers:
  - volumeMounts:
    - mountPath: /etc/kubernetes/imagine
      name: imagine
      readOnly: true
...
  volumes:
  - hostPath:
      path: /etc/kubernetes/imagine
      type: DirectoryOrCreate
    name: imagine
```

Wait for a while to let the `kube-apiserver` start up again, check the status with e.g. `crictl`.

You can now run the simplest of tests:

```bash
kubectl run --image nginx nginx-1
pod/nginx-1 created
```

vs

```bash
kubectl run --image nope nginx-2
Error from server (Forbidden): pods "nginx-2" is forbidden: image policy webhook backend denied one or more images: image name contains disallowed string: nope
```

And from the perspective of `imagine` these requests looks something like this:

```bash
imagine-67d4f474cf-4hf8m 2024/11/25 20:22:40 Received request: POST /
imagine-67d4f474cf-4hf8m 2024/11/25 20:22:40 Raw JSON request body: {"kind":"ImageReview","apiVersion":"imagepolicy.k8s.io/v1alpha1","metadata":{"creationTimestamp":null},"s
pec":{"containers":[{"image":"nginx"}],"namespace":"default"},"status":{"allowed":false}}
imagine-67d4f474cf-4hf8m 2024/11/25 20:22:40 Image: nginx, Allowed: true, Reason: Image name is allowed
imagine-67d4f474cf-4hf8m 2024/11/25 20:22:53 Received request: POST /
imagine-67d4f474cf-4hf8m 2024/11/25 20:22:53 Raw JSON request body: {"kind":"ImageReview","apiVersion":"imagepolicy.k8s.io/v1alpha1","metadata":{"creationTimestamp":null},"s
pec":{"containers":[{"image":"nope"}],"namespace":"default"},"status":{"allowed":false}}
imagine-67d4f474cf-4hf8m 2024/11/25 20:22:53 Image: nope, Allowed: false, Reason: image name contains disallowed string: nope
```

Congratulations, you're now done! 🎉

### Troubleshooting

* Check the logs of `imagine`, it logs incoming requests verbatim, and also the status and reason of the validating admission control request.
* Check the logs of the `kube-apiserver`.

## Development

Check that you have all necessary tools installed:

```bash
make check-tools
```

### Local testing

Generate certificates:

1. Generates the CA private key
2. Generates the CA certificate
3. Generates the private key to be used by the webhook HTTP server
4. Generates the webhook HTTP server CSR
5. Signs the CSR with the CA's key and certificate to issue the webhook HTTP server certificate

```bash
make gen-certs
```

You can now start the webhook HTTP server:

```bash
make run
```

Send two requests that includes two admission requests with Pod container images named: `nope:latest` and `nginx:latest`. We'll not allow container images containing `nope` to be started basically:

```bash
make send
```
