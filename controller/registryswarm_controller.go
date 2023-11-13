package controller

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/stenstromen/registryswarm/model"
)

func GetRegistries(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		registries, err := model.GetRegistries(db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(registries); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func ProxyRequest(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for name, headers := range r.Header {
			for _, h := range headers {
				fmt.Printf("%v: %v\n", name, h)
			}
		}

		fmt.Printf("Received request for: %s\n", r.URL.Path)
		username := extractUsername(r.Header.Get("Authorization"))
		fmt.Printf("Extracted username: %s\n", username)
		if username == "" {
			if r.Method == "GET" && r.URL.Path == "/v2/" {
				issueAuthChallenge(w)
				return
			}
			fmt.Println("Authorization header missing or invalid")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		targetURL, err := model.GetBackendURL(db, username)
		if err != nil {
			fmt.Printf("No backend URL found for user: %s\n", username)
			http.Error(w, "Registry not found", http.StatusNotFound)
			return
		}
		if targetURL == nil {
			fmt.Printf("No backend URL found for user: %s\n", username)
			http.Error(w, "Registry not found", http.StatusNotFound)
			return
		}

		reverseProxy := httputil.NewSingleHostReverseProxy(targetURL)
		reverseProxy.ErrorLog = log.New(os.Stderr, "proxy error: ", log.LstdFlags)
		reverseProxy.Director = func(req *http.Request) {
			req.URL.Scheme = targetURL.Scheme
			req.URL.Host = targetURL.Host
			req.Header = r.Header.Clone()

			if req.URL.Path == "/v2/" {
				// This is a ping request; let it pass through as-is
				return
			}

			// Rewrite the URL path for push/pull operations
			if strings.HasPrefix(req.URL.Path, "/v2/") {
				splitPath := strings.SplitN(req.URL.Path, "/", 5)
				if len(splitPath) >= 4 && splitPath[2] == username {
					// Rewrite the path to exclude the username for the matched user
					req.URL.Path = fmt.Sprintf("/v2/%s", strings.Join(splitPath[3:], "/"))
				}
				// No additional path adjustment needed for other users
			}
		}

		fmt.Printf("Forwarding request to: %s\n", targetURL)
		reverseProxy.ServeHTTP(w, r)
	}
}

func extractUsername(authHeader string) string {
	if !strings.HasPrefix(authHeader, "Basic ") {
		return ""
	}
	encoded := strings.TrimPrefix(authHeader, "Basic ")
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return ""
	}
	credentials := strings.SplitN(string(decoded), ":", 2)
	if len(credentials) < 2 {
		return ""
	}
	return credentials[0]
}

func issueAuthChallenge(w http.ResponseWriter) {
	w.Header().Set("Www-Authenticate", `Basic realm="Registry Realm"`)
	w.WriteHeader(http.StatusUnauthorized)
}
