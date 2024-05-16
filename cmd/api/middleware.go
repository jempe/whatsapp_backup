package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"net/http"

	"golang.org/x/time/rate"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Defer a function to recover from any panics.
		defer func() {
			// Use the builtin recover function to check if there has been a panic or not.
			if err := recover(); err != nil {
				// Set a "Connection: close" header on the response.
				w.Header().Set("Connection", "close")
				// Call the app.serverError helper method to return a 500 Internal
				// Server response.
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	limiter := rate.NewLimiter(2, 4)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Call the RateLimit method on the limiter to check if the request is permitted
		// or not.
		if !limiter.Allow() {
			// If the request is not allowed, call the app.rateLimitExceeded helper
			// method to return a 429 Too Many Requests response.
			app.rateLimitExceededResponse(w, r)
			return
		}

		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}

func (app *application) basicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		username, password, ok := r.BasicAuth()
		if ok {
			usernameHash := sha256.Sum256([]byte(username))
			passwordHash := sha256.Sum256([]byte(password))
			expectedUsernameHash := sha256.Sum256([]byte(app.config.basicAuth.username))
			expectedPasswordHash := sha256.Sum256([]byte(app.config.basicAuth.password))

			usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1)
			passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)

			if usernameMatch && passwordMatch {
				next.ServeHTTP(w, r)
				return
			} else {
				app.serverErrorResponse(w, r, fmt.Errorf("Unauthorized"))
			}
		}
		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

func (app *application) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow CORS here By * or specific origin
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
		next.ServeHTTP(w, r)
	})
}
