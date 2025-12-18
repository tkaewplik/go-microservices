package main

import (
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/tkaewplik/go-microservices/pkg/middleware"
)

type Gateway struct {
	authServiceURL    *url.URL
	paymentServiceURL *url.URL
}

func NewGateway(authURL, paymentURL string) (*Gateway, error) {
	authServiceURL, err := url.Parse(authURL)
	if err != nil {
		return nil, err
	}

	paymentServiceURL, err := url.Parse(paymentURL)
	if err != nil {
		return nil, err
	}

	return &Gateway{
		authServiceURL:    authServiceURL,
		paymentServiceURL: paymentServiceURL,
	}, nil
}

func (g *Gateway) ProxyRequest(target *url.URL) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		proxy := httputil.NewSingleHostReverseProxy(target)
		
		// Customize director to preserve headers and path
		director := proxy.Director
		proxy.Director = func(req *http.Request) {
			director(req)
			req.Host = target.Host
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
		}

		// Error handler
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("Proxy error: %v", err)
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte("Service unavailable"))
		}

		proxy.ServeHTTP(w, r)
	}
}

func (g *Gateway) HandleRequest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Route to auth service
	if strings.HasPrefix(path, "/auth/") {
		r.URL.Path = strings.TrimPrefix(path, "/auth")
		g.ProxyRequest(g.authServiceURL)(w, r)
		return
	}

	// Route to payment service
	if strings.HasPrefix(path, "/payment/") {
		r.URL.Path = strings.TrimPrefix(path, "/payment")
		g.ProxyRequest(g.paymentServiceURL)(w, r)
		return
	}

	// Health check
	if path == "/health" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"status":"ok"}`)
		return
	}

	http.NotFound(w, r)
}

func main() {
	authServiceURL := getEnv("AUTH_SERVICE_URL", "http://localhost:8081")
	paymentServiceURL := getEnv("PAYMENT_SERVICE_URL", "http://localhost:8082")

	gateway, err := NewGateway(authServiceURL, paymentServiceURL)
	if err != nil {
		log.Fatalf("Failed to create gateway: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", gateway.HandleRequest)

	handler := middleware.CORS(mux)

	port := getEnv("PORT", "8080")
	log.Printf("API Gateway starting on port %s", port)
	log.Printf("Routing /auth/* to %s", authServiceURL)
	log.Printf("Routing /payment/* to %s", paymentServiceURL)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
