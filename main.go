package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	admission "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
)

type imageineOpts struct {
	key       string
	cert      string
	imageName string
	port      int
}

func main() {
	opts := imageineOpts{}

	// Parse command line arguments
	flag.StringVar(&opts.key, "key", "", "Path to the key file")
	flag.StringVar(&opts.cert, "cert", "", "Path to the cert file")
	flag.StringVar(&opts.imageName, "image-name", "", "Part of the image name used when validating")
	flag.IntVar(&opts.port, "port", 4443, "Port to listen on")
	flag.Parse()

	cert, err := tls.LoadX509KeyPair(opts.cert, opts.key)
	if err != nil {
		log.Fatalf("Failed to load key pair: %v", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{
			cert,
		},
	}

	mux := http.NewServeMux()
	mux.Handle("/", imagineHandler(opts.imageName))

	server := &http.Server{
		Addr:      fmt.Sprintf(":%d", opts.port),
		Handler:   mux,
		TLSConfig: tlsConfig,
	}

	log.Printf("Starting the webhook server on port %d", opts.port)
	err = server.ListenAndServeTLS("", "")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func imagineHandler(imageName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			if r.Method == http.MethodGet {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(fmt.Sprintf("This is the Imagine validating admission image policy webhook, checks if the image name contains: %s", imageName)))
				return
			}

			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("Method not allowed"))
		}

		var admissionReview admission.AdmissionReview
		if err := json.NewDecoder(r.Body).Decode(&admissionReview); err != nil {
			http.Error(w, "could not decode request body", http.StatusBadRequest)
			return
		}

		var pod corev1.Pod
		if err := json.Unmarshal(admissionReview.Request.Object.Raw, &pod); err != nil {
			http.Error(w, "could not decode pod spec", http.StatusBadRequest)
			return
		}

		// Check if the provided image name is in the Pod's containers
		var allowed bool
		for _, container := range pod.Spec.Containers {
			if !strings.Contains(container.Image, imageName) {
				allowed = true
				break
			}
		}

		admissionResponse := admission.AdmissionResponse{
			Allowed: allowed,
		}

		admissionReview.Response = &admissionResponse
		responseBytes, err := json.Marshal(admissionReview)
		if err != nil {
			http.Error(w, "could not encode response", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(responseBytes)
	}
}
