package notepad

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Run() error {
	baseDir := detectBaseDir()

	app, err := newApp(baseDir)
	if err != nil {
		return err
	}

	router := newRouter(app, baseDir)
	port := envOrDefault("PORT", defaultPort)
	return router.Run(":" + port)
}

func newApp(baseDir string) (*App, error) {
	jwtSecret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET is required")
	}

	sessionTTL, err := parseJWTExpiresIn(envOrDefault("JWT_SESSION_EXPIRES_IN", defaultSessionExpiresIn))
	if err != nil {
		return nil, fmt.Errorf("parse JWT_SESSION_EXPIRES_IN: %w", err)
	}

	rememberTTL, err := parseJWTExpiresIn(envOrDefault("JWT_REMEMBER_EXPIRES_IN", defaultRememberExpiresIn))
	if err != nil {
		return nil, fmt.Errorf("parse JWT_REMEMBER_EXPIRES_IN: %w", err)
	}

	store := &Store{
		notesDir:   filepath.Join(baseDir, "notes"),
		uploadsDir: filepath.Join(baseDir, "uploads"),
		metaFile:   filepath.Join(baseDir, "notes", "meta.json"),
	}

	if err := store.bootstrap(); err != nil {
		return nil, err
	}

	return &App{
		store:       store,
		password:    envOrDefault("PASSWORD", defaultPassword),
		jwtSecret:   []byte(jwtSecret),
		sessionTTL:  sessionTTL,
		rememberTTL: rememberTTL,
	}, nil
}
