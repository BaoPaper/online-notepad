package notepad

import (
	"testing"
	"time"
)

func TestParseJWTExpiresIn(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Duration
		wantErr bool
	}{
		{name: "days", input: "1d", want: 24 * time.Hour},
		{name: "hours", input: "2h", want: 2 * time.Hour},
		{name: "words", input: "2 days", want: 48 * time.Hour},
		{name: "milliseconds string", input: "120", want: 120 * time.Millisecond},
		{name: "invalid", input: "soon", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseJWTExpiresIn(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseJWTExpiresIn(%q): %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("parseJWTExpiresIn(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestSignAndVerifyAuthToken(t *testing.T) {
	app := newTestApp(t)

	token, err := app.signAuthToken(true)
	if err != nil {
		t.Fatalf("signAuthToken: %v", err)
	}

	claims := app.verifyAuthToken(token)
	if claims == nil {
		t.Fatal("expected claims, got nil")
	}
	if claims["role"] != "user" {
		t.Fatalf("role = %v, want user", claims["role"])
	}
	if app.verifyAuthToken(token+"broken") != nil {
		t.Fatal("expected invalid token to return nil claims")
	}
}
