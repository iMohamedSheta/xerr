package xerr

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfigValues(t *testing.T) {
	cfg := DefaultConfig()
	assert.True(t, cfg.ShowSourceCode, "Default ShowSourceCode should be true")
	assert.Equal(t, 50, cfg.MaxFrames, "Default MaxFrames should be 50")
	assert.Equal(t, "development", cfg.Environment, "Default Environment should be development")
	assert.True(t, cfg.DebugMode, "Default DebugMode should be true")
	assert.Equal(t, 2, cfg.SkipFrames, "Default SkipFrames should be 2")
}

func TestNewErrorHandlerWithNilConfig(t *testing.T) {
	eh := New(nil)
	assert.NotNil(t, eh, "ErrorHandler should not be nil")
	assert.NotNil(t, eh.tpl, "Template should be initialized")
}

func TestNewErrorHandlerWithCustomConfig(t *testing.T) {
	cfg := &Config{ShowSourceCode: false, MaxFrames: 5}
	eh := New(cfg)
	assert.Equal(t, cfg, eh.config, "Custom config should be applied")
}

func TestHandleErrorRendersHTML(t *testing.T) {
	eh := New(nil)
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	eh.HandleError(w, r, "test error")

	resp := w.Result()
	body := w.Body.String()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode, "Should return 500 status")
	assert.Contains(t, body, "test error", "Body should contain error message")
	assert.Equal(t, "text/html; charset=utf-8", resp.Header.Get("Content-Type"))
}

func TestHandlePanicRecoversAndRenders(t *testing.T) {
	eh := New(nil)
	r := httptest.NewRequest(http.MethodGet, "/panic", nil)
	w := httptest.NewRecorder()

	func() {
		defer eh.HandlePanic(w, r)
		panic("panic test")
	}()

	body := w.Body.String()
	assert.Contains(t, body, "panic test", "Body should contain panic message")
}

func TestMiddlewareCatchesPanic(t *testing.T) {
	eh := New(nil)
	h := eh.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("middleware panic")
	}))
	r := httptest.NewRequest(http.MethodGet, "/panic", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	assert.Contains(t, w.Body.String(), "middleware panic")
}

func TestMiddlewareFuncCatchesPanic(t *testing.T) {
	eh := New(nil)
	hf := eh.MiddlewareFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("middleware func panic")
	})
	r := httptest.NewRequest(http.MethodGet, "/panic", nil)
	w := httptest.NewRecorder()
	hf(w, r)
	assert.Contains(t, w.Body.String(), "middleware func panic")
}

func TestCodeSnippetWithExistingFile(t *testing.T) {
	filename := "testfile.go"
	content := "package test\n\nfunc Test() {}\n"
	err := os.WriteFile(filename, []byte(content), 0644)
	assert.NoError(t, err)
	defer os.Remove(filename)

	eh := New(nil)
	snippet := eh.codeSnippet(filename, 2)
	assert.Contains(t, snippet, "func Test() {}")
}

func TestCodeSnippetWithNonExistentFile(t *testing.T) {
	eh := New(nil)
	snippet := eh.codeSnippet("nofile.go", 1)
	assert.Equal(t, "Could not read source file", snippet)
}

func TestStackFramesReturnsFrames(t *testing.T) {
	eh := New(&Config{MaxFrames: 10, SkipFrames: 0})
	frames := eh.stackFrames(nil)
	assert.Greater(t, len(frames), 0, "Should return at least one frame")
	assert.NotEmpty(t, frames[0].Function)
	assert.NotEmpty(t, frames[0].File)
}

func TestParseCodeLineTemplateFunc(t *testing.T) {
	parse := templateFuncs["parseCodeLine"].(func(string) map[string]string)

	line := ">> 10 | content here"
	parsed := parse(line)
	assert.Equal(t, "true", parsed["highlight"])
	assert.Equal(t, "10", parsed["number"])
	assert.Equal(t, " content here", parsed["content"])

	line2 := "   5 | regular content"
	parsed2 := parse(line2)
	assert.Equal(t, "false", parsed2["highlight"])
	assert.Equal(t, "5", parsed2["number"])
	assert.Equal(t, " regular content", parsed2["content"])
}

func TestLenTemplateFunc(t *testing.T) {
	lenFunc := templateFuncs["len"].(func(interface{}) int)
	assert.Equal(t, 3, lenFunc([]Frame{{}, {}, {}}))
	assert.Equal(t, 5, lenFunc("hello"))
	assert.Equal(t, 0, lenFunc(123))
}

func TestErrorHandler_HandlePanic_WithPanic(t *testing.T) {
	eh := New(nil)
	r := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()
	func() {
		defer eh.HandlePanic(rw, r)
		panic("panic test")
	}()
	assert.Equal(t, http.StatusInternalServerError, rw.Code)
	assert.Contains(t, rw.Body.String(), "panic test")
}

func TestErrorHandler_CodeSnippet_DisabledSource(t *testing.T) {
	eh := New(&Config{ShowSourceCode: false})
	snippet := eh.codeSnippet("anyfile.go", 10)
	assert.Equal(t, "Source code display disabled", snippet)
}

func TestErrorHandler_CodeSnippet_FileDoesNotExist(t *testing.T) {
	eh := New(nil)
	snippet := eh.codeSnippet("nonexistent.go", 10)
	assert.Equal(t, "Could not read source file", snippet)
}

func TestErrorHandler_CodeSnippet_FileBoundaries(t *testing.T) {
	filename := "temp_test.go"
	lines := []string{"line1", "line2", "line3"}
	err := os.WriteFile(filename, []byte(strings.Join(lines, "\n")), 0644)
	assert.NoError(t, err)
	defer os.Remove(filename)

	eh := New(nil)
	snippet := eh.codeSnippet(filename, 1)
	assert.Contains(t, snippet, "line1")
	snippetEnd := eh.codeSnippet(filename, 3)
	assert.Contains(t, snippetEnd, "line3")
}

func TestErrorHandler_StackFrames_NotEmpty(t *testing.T) {
	eh := New(nil)
	frames := eh.stackFrames(nil)
	assert.NotEmpty(t, frames)
	for _, f := range frames {
		assert.NotEmpty(t, f.Function)
		assert.NotEmpty(t, f.File)
	}
}

func TestMiddleware_CallsNextHandler(t *testing.T) {
	eh := New(nil)
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })
	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	eh.Middleware(next).ServeHTTP(rw, req)
	assert.True(t, called)
}

func TestMiddleware_PanicRecovery(t *testing.T) {
	eh := New(nil)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { panic("handler panic") })
	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	eh.Middleware(next).ServeHTTP(rw, req)
	assert.Equal(t, http.StatusInternalServerError, rw.Code)
	assert.Contains(t, rw.Body.String(), "handler panic")
}

func TestMiddlewareFunc_PanicRecovery(t *testing.T) {
	eh := New(nil)
	next := func(w http.ResponseWriter, r *http.Request) { panic("func panic") }
	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	eh.MiddlewareFunc(next)(rw, req)
	assert.Equal(t, http.StatusInternalServerError, rw.Code)
	assert.Contains(t, rw.Body.String(), "func panic")
}
