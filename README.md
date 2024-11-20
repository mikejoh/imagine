# imagine 🧞

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

---

_But wait! There will be more, if you stay tuned for a while i'll add the rest also. We're missing the Kubernetes related things!_ :smiling_face_with_tear:
