package main

import (
	"fmt"
	"net/http"

	"github.com/iMohamedSheta/xerr"
)

// ===== EXAMPLE USAGE =====

// Example using as middleware
func ExampleMiddleware() {
	// Create error handler with custom config
	config := &xerr.Config{
		ShowSourceCode: true,
		MaxFrames:      20,
		Environment:    "production",
		DebugMode:      false,
	}
	errorHandler := xerr.NewErrorHandler(config)

	// Use as middleware
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		panic("Something went wrong!")
	})

	http.ListenAndServe(":8080", errorHandler.Middleware(mux))
}

// Example handling errors manually
func ExampleManualError() {
	errorHandler := xerr.NewErrorHandler(nil) // Use default config

	http.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				errorHandler.HandleError(w, r, rec)
				return
			}
		}()

		// Your application logic here
		panic("Database connection failed")
	})
}

// Example with custom error types
func ExampleCustomError() {
	errorHandler := xerr.NewErrorHandler(nil)

	http.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
		// Handle specific errors
		if err := someOperation(); err != nil {
			errorHandler.HandleError(w, r, fmt.Errorf("operation failed: %w", err))
			return
		}

		w.Write([]byte("Success"))
	})
}

func someOperation() error {
	return fmt.Errorf("simulated error")
}

// Example main function showing different usage patterns
func main() {
	// Method 1: Simple middleware usage
	errorHandler := xerr.NewErrorHandler(xerr.DefaultConfig())

	mux := http.NewServeMux()
	mux.HandleFunc("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("boom! something failed")
	})

	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		errorHandler.HandleError(w, r, fmt.Errorf("custom error occurred"))
	})

	mux.HandleFunc("/xerr", func(w http.ResponseWriter, r *http.Request) {
		errorHandler.HandleError(w, r, doAction())
	})

	mux.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("assets/css"))))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello! Try /panic or /error to see the error pages"))
	})

	fmt.Println("Server starting on http://localhost:8085")
	fmt.Println("Routes:")
	fmt.Println("  /        - Hello message")
	fmt.Println("  /panic   - Trigger a panic")
	fmt.Println("  /error   - Trigger a custom error")
	fmt.Println("  /xerr    - Trigger a xerr error which is a custom error with stacktrace and type ex xerr.Error(...)")

	if err := http.ListenAndServe(":8085", errorHandler.Middleware(mux)); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}

func doAction() error {
	return xerr.New("custom error occurred", xerr.ErrUnknown, nil)
}
