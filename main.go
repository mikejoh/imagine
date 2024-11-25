package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	imagepolicy "k8s.io/api/imagepolicy/v1alpha1"
)

type imagineOpts struct {
	key       string
	cert      string
	imageName string
	port      int
}

func main() {
	opts := imagineOpts{}

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

	log.Printf("Starting the imagine server on port %d", opts.port)
	err = server.ListenAndServeTLS("", "")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func imagineHandler(imageName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request: %s %s", r.Method, r.URL.Path)
		if r.Method != http.MethodPost {
			if r.Method == http.MethodGet {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(fmt.Sprintf("This is the Imagine validating admission image policy webhook, checks if the image name contains: %s", imageName)))
				return
			}

			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("Method not allowed"))
			return
		}

		var imageReview imagepolicy.ImageReview

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Failed to read request body: %v", err)
			http.Error(w, "could not read request body", http.StatusBadRequest)
			return
		}

		log.Printf("Raw JSON request body: %s", string(body))
		if err := json.Unmarshal(body, &imageReview); err != nil {
			log.Printf("Failed to decode request body: %v", err)
			http.Error(w, "could not decode request body", http.StatusBadRequest)
			return
		}

		var allowed bool
		for _, container := range imageReview.Spec.Containers {
			if strings.Contains(container.Image, imageName) {
				allowed = true
				break
			}
		}

		imageReview.Status.Allowed = allowed

		responseBytes, err := json.Marshal(imageReview)
		if err != nil {
			log.Printf("Failed to encode response: %v", err)
			http.Error(w, "could not encode response", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(responseBytes); err != nil {
			log.Printf("Failed to write response: %v", err)
		}
	}
}
