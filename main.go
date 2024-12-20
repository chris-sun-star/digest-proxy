package main

import (
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/icholy/digest"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	// Read these from environment variables
	targetURL = os.Getenv("TARGET_URL")
	username  = os.Getenv("USERNAME")
	password  = os.Getenv("PASSWORD")
	port      = os.Getenv("PORT")
)

func main() {
	initLogger()
	if targetURL == "" || username == "" || password == "" {
		log.Fatal().Msg("Missing required environment variables: TARGET_URL, USERNAME, PASSWORD")
	}

	// Default port if not set
	if port == "" {
		port = "3000"
	}

	r := mux.NewRouter()

	r.PathPrefix("/").HandlerFunc(proxyHandler)

	log.Info().Msgf("Proxy server running on port %s, targeting service at: %s", port, targetURL)

	// CORS middleware
	corsMiddleware := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}), // Replace with your Swagger UI address
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	if err := http.ListenAndServe(":"+port, corsMiddleware(r)); err != nil {
		log.Fatal().Msgf("Failed to start server: %v", err)
	}
}

func initLogger() {
	zerolog.TimeFieldFormat = time.DateTime
	log.Logger = zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
		FormatLevel: func(i interface{}) string {
			return "[" + i.(string) + "]"
		},
		FormatMessage: func(i interface{}) string {
			return i.(string)
		},
		NoColor: true,
	}).With().Timestamp().Logger()
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	t := &digest.Transport{
		Username: username,
		Password: password,
	}

	// Prepare the Digest Authentication client
	client := &http.Client{Transport: t}

	reqURL := targetURL + r.URL.Path
	log.Info().Msgf("Proxying to URL: %s", reqURL)

	// copy query parameters
	if r.URL.RawQuery != "" {
		reqURL += "?" + r.URL.RawQuery
	}

	// Copy the incoming request body to send it to the target server
	body := io.Reader(r.Body)
	if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
		body = r.Body
	}

	// Create the request to the target URL
	req, err := http.NewRequest(r.Method, reqURL, body)
	if err != nil {
		http.Error(w, "Failed to create request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Add headers from the original request (excluding host and content-length)
	for key, values := range r.Header {
		if key != "Host" && key != "Content-Length" {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	}

	// Send the request to the target server
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to proxy request: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	log.Info().Msgf("Target server response: %d", resp.StatusCode)
	// Copy the response headers from the target server
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Set the status code
	w.WriteHeader(resp.StatusCode)

	// Copy the response body to the client
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		http.Error(w, "Failed to copy response: "+err.Error(), http.StatusInternalServerError)
	}
}
