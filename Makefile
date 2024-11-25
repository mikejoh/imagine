VERSION := $(shell git describe --tags --always --dirty)

check-tools:
	@echo "Checking if you have the required tools installed.."
	@echo
	@which go
	@which openssl
	@which curl
	@which docker

gen-certs:
	@echo "Generating certificates in certs directory.."
	mkdir -p certs
	openssl genpkey -algorithm RSA -out certs/ca.key -pkeyopt rsa_keygen_bits:2048
	openssl req -x509 -new -nodes -key certs/ca.key -sha256 -days 3650 -out certs/ca.crt -subj "/C=US/ST=State/L=City/O=Organization/OU=Unit/CN=CA"
	openssl genpkey -algorithm RSA -out certs/webhook.key -pkeyopt rsa_keygen_bits:2048
	openssl req -new -key certs/webhook.key -subj "/CN=localhost" -addext "subjectAltName=IP:127.0.0.1" -out certs/webhook.csr
	openssl x509 -req -in certs/webhook.csr -CA certs/ca.crt -CAkey certs/ca.key -CAcreateserial -out certs/webhook.crt -days 365 -sha256

run:
	@echo "Running the webhook server"
	go run ./main.go -key=certs/webhook.key -cert=certs/webhook.crt -image-name nope

send:
	@echo "Sending a request that includes a Pod container with allowed image name:"
	@echo
	curl -k -X POST -H "Content-Type: application/json" --data @files/image_review_1.json https://localhost:4443/
	@echo
	@echo

	@echo "Sending a request that includes a Pod container with denied image name:"
	@echo
	curl -k -X POST -H "Content-Type: application/json" --data @files/image_review_2.json https://localhost:4443/
	@echo

docker-build:
	@echo "Building the container image.."
	docker build -t mikejoh/imagine:$(VERSION) .

docker-push:
	@echo "Pushing the container image.."
	docker push mikejoh/imagine:$(VERSION)

release:
	@echo "Checking if the working directory is clean.."
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "Error: Working directory is dirty. Commit or stash your changes before releasing."; \
		exit 1; \
	fi
	@echo "Checking if the current commit has a tag..."
	@if [ -z "$$(git describe --tags --exact-match 2>/dev/null)" ]; then \
		echo "Error: No tag found on the current commit. Please tag the commit before releasing."; \
		exit 1; \
	fi
	@echo "Building and pushing the container image for release..."
	make docker-build
	make docker-push
	@echo "Release completed successfully."

deploy:
	@echo "Rendering deployment manifest.."
	helm repo add mikejoh https://mikejoh.github.io/helm-charts/
	helm repo update mikejoh
	helm \
		upgrade \
		imagine \
		--install \
		--create-namespace \
		--namespace imagine \
		mikejoh/imagine

k8s-gen-cert:
	@echo "Generating certificate for use with Kubernetes.."
	@SERVICE_IP=$$(kubectl get svc -n imagine -o jsonpath='{.items[0].spec.clusterIP}'); \
	openssl req \
		-new \
		-key certs/webhook.key \
		-subj "/CN=system:node:imagine/O=system:nodes" \
		-addext "subjectAltName = DNS:imagine.imagine.svc.cluster.local,DNS:imagine.imagine.svc,DNS:imagine.imagine.pod.cluster.local,IP:$$SERVICE_IP" \
		-out certs/k8s-webhook.csr
	
	@SIGNING_REQUEST=$$(cat certs/k8s-webhook.csr | base64 | tr -d '\n'); \
	export SIGNING_REQUEST=$$SIGNING_REQUEST; \
	envsubst < files/csr-template.yaml | kubectl apply -f -

	kubectl certificate approve imagine
	kubectl get csr imagine -o=jsonpath={.status.certificate} | base64 --decode > certs/k8s-webhook.crt

	kubectl create secret tls imagine-tls -n imagine --cert=certs/k8s-webhook.crt --key=certs/webhook.key


.PHONY: gen-certs run send docker-build docker-push release check-tools deploy k8s-gen-cert