package notepad

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func newTestApp(t *testing.T) *App {
	t.Helper()

	baseDir := t.TempDir()
	store := &Store{
		notesDir:   filepath.Join(baseDir, "notes"),
		uploadsDir: filepath.Join(baseDir, "uploads"),
		metaFile:   filepath.Join(baseDir, "notes", "meta.json"),
	}
	if err := store.bootstrap(); err != nil {
		t.Fatalf("bootstrap store: %v", err)
	}

	return &App{
		store:       store,
		password:    "password",
		jwtSecret:   []byte("secret"),
		sessionTTL:  time.Hour,
		rememberTTL: 30 * 24 * time.Hour,
	}
}

func newTestRouter(t *testing.T, app *App) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	return newRouter(app, projectRoot(t))
}

func projectRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	root := wd
	for {
		if hasProjectAssets(root) {
			return root
		}

		parent := filepath.Dir(root)
		if parent == root {
			t.Fatalf("project root not found from %s", wd)
		}
		root = parent
	}
}

func authCookie(t *testing.T, app *App) *http.Cookie {
	t.Helper()
	token, err := app.signAuthToken(false)
	if err != nil {
		t.Fatalf("sign auth token: %v", err)
	}
	return &http.Cookie{Name: authCookieName, Value: token, Path: "/"}
}

func performRequest(router *gin.Engine, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}
