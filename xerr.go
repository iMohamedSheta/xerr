package xerr

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// HTML templates for the error page
var errorTemplate = []string{
	filepath.Join(packageRoot(), "assets", "templates", "error.html"),
}

// the executed template to show the error page
const execTemplate = "error.html"

// Frame represents a single stack frame
type Frame struct {
	Function string
	File     string
	Line     int
	Snippet  string
}

// ErrorData contains all the information needed to render an error page
type ErrorData struct {
	Error     string
	Frames    []Frame
	Timestamp time.Time
	Method    string
	URL       string
	UserAgent string
	GoVersion string
	OS        string
	Arch      string
	Request   *http.Request
}

// Config holds configuration options for the error handler
type Config struct {
	ShowSourceCode bool   // Whether to show source code snippets
	MaxFrames      int    // Maximum number of stack frames to display
	Environment    string // Environment name (development, production, etc.)
	DebugMode      bool   // Whether debug mode is enabled
	SkipFrames     int    // Number of frames to skip from the top
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		ShowSourceCode: true,
		MaxFrames:      50,
		Environment:    "development",
		DebugMode:      true,
		SkipFrames:     2, // Skip the panic, recover, and this function
	}
}

// ErrorHandler handles errors and renders the error page
type ErrorHandler struct {
	config *Config
	tpl    *template.Template
}

// New creates a new ErrorHandler with the given configuration
func New(config *Config) *ErrorHandler {
	if config == nil {
		config = DefaultConfig()
	}

	tpl := template.Must(
		template.New("error").Funcs(templateFuncs).ParseFiles(errorTemplate...),
	)

	return &ErrorHandler{
		config: config,
		tpl:    tpl,
	}
}

// HandleError renders an error page for the given error and writes it to the ResponseWriter
func (eh *ErrorHandler) HandleError(w http.ResponseWriter, r *http.Request, err interface{}) {
	data := &ErrorData{
		Error:     fmt.Sprintf("%v", err),
		Frames:    eh.stackFrames(),
		Timestamp: time.Now(),
		GoVersion: strings.TrimPrefix(runtime.Version(), "go"),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		Request:   r,
	}

	if r != nil {
		data.Method = r.Method
		data.URL = r.URL.String()
		data.UserAgent = r.UserAgent()
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)

	if renderErr := eh.tpl.ExecuteTemplate(w, execTemplate, data); renderErr != nil {
		// Fallback to plain text if template rendering fails
		fmt.Fprintf(w, "Error: %v\n\nTemplate rendering failed: %v", err, renderErr)
	}
}

// HandlePanic is a convenience method for handling panics in HTTP handlers
func (eh *ErrorHandler) HandlePanic(w http.ResponseWriter, r *http.Request) {
	if rec := recover(); rec != nil {
		eh.HandleError(w, r, rec)
	}
}

// Middleware returns an HTTP middleware that catches panics and renders error pages
func (eh *ErrorHandler) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				eh.HandleError(w, r, rec)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// MiddlewareFunc returns an HTTP middleware function that catches panics and renders error pages
func (eh *ErrorHandler) MiddlewareFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				eh.HandleError(w, r, rec)
			}
		}()
		next(w, r)
	}
}

// codeSnippet extracts a code snippet around the given line in the file
func (eh *ErrorHandler) codeSnippet(file string, line int) string {
	if !eh.config.ShowSourceCode {
		return "Source code display disabled"
	}

	data, err := os.ReadFile(file)
	if err != nil {
		return "Could not read source file"
	}

	lines := strings.Split(string(data), "\n")
	start := line - 15
	if start < 0 {
		start = 0
	}
	end := line + 20
	if end > len(lines) {
		end = len(lines)
	}

	var b strings.Builder
	for i := start; i < end; i++ {
		prefix := "   "
		if i+1 == line {
			prefix = ">> "
		}
		fmt.Fprintf(&b, "%s%4d | %s\n", prefix, i+1, lines[i])
	}
	return b.String()
}

// stackFrames extracts stack frames from the current goroutine
func (eh *ErrorHandler) stackFrames() []Frame {
	pcs := make([]uintptr, eh.config.MaxFrames)
	n := runtime.Callers(eh.config.SkipFrames+1, pcs)
	iter := runtime.CallersFrames(pcs[:n])

	var frames []Frame
	for {
		fr, more := iter.Next()
		if fr.File != "" {
			// Skip standard library and module cache files
			if strings.Contains(fr.File, "/go/src/") || strings.Contains(fr.File, "/pkg/mod/") {
				if !more {
					break
				}
				continue
			}

			frame := Frame{
				Function: fr.Function,
				File:     fr.File,
				Line:     fr.Line,
				Snippet:  eh.codeSnippet(fr.File, fr.Line),
			}
			frames = append(frames, frame)

			if len(frames) >= eh.config.MaxFrames {
				break
			}
		}
		if !more {
			break
		}
	}
	return frames
}

// Template functions for the HTML template
var templateFuncs = template.FuncMap{
	"split":     strings.Split,
	"contains":  strings.Contains,
	"trimSpace": strings.TrimSpace,
	"parseCodeLine": func(line string) map[string]string {
		result := map[string]string{
			"number":    "",
			"content":   "",
			"highlight": "false",
		}

		if strings.TrimSpace(line) == "" {
			return result
		}

		// Check if this is a highlighted line (starts with ">>")
		if strings.HasPrefix(line, ">>") {
			result["highlight"] = "true"
			// Extract line number and content from ">> 123 | content"
			parts := strings.SplitN(line[2:], "|", 2)
			if len(parts) == 2 {
				result["number"] = strings.TrimSpace(parts[0])
				result["content"] = parts[1]
			}
		} else {
			// Regular line format "   123 | content"
			parts := strings.SplitN(line, "|", 2)
			if len(parts) == 2 {
				result["number"] = strings.TrimSpace(parts[0])
				result["content"] = parts[1]
			}
		}

		return result
	},
	"len": func(v interface{}) int {
		switch s := v.(type) {
		case []Frame:
			return len(s)
		case string:
			return len(s)
		default:
			return 0
		}
	},
}

func packageRoot() string {
	_, file, _, _ := runtime.Caller(0) // path to this source file
	return filepath.Dir(file)
}
