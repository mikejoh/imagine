# imagine 🧞

[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/mikejoh)](https://artifacthub.io/packages/search?repo=mikejoh)

`imagine` - The simplest [`ImagePolicyWebhook`](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#imagepolicywebhook) webhook example you'll ever find!

_Yes, you're correct, i'm studying for the CKS exam!_

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

Send two requests that includes two admission requests with Pod container images named: `nope:latest` and `nginx:latest`. We'll not allow images containing `nope` to be started basically:

```bash
curl --cacert ./certs/ca.crt https://localhost:4443
```

### Deploy in Kubernetes

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
