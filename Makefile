check-tools:
	@echo "Checking if you have the required tools installed:"
	@echo
	@which go
	@which openssl
	@which curl

gen-certs:
	@echo "Generating certificates in certs directory"
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
	curl -k -X POST -H "Content-Type: application/json" --data @files/admission_req_1.json https://localhost:4443/
	@echo
	@echo

	@echo "Sending a request that includes a Pod container with denied image name:"
	@echo
	curl -k -X POST -H "Content-Type: application/json" --data @files/admission_req_2.json https://localhost:4443/
	@echo

PHONY: gen-certs run send