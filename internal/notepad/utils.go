package notepad

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func wantsJSON(c *gin.Context) bool {
	accept := c.GetHeader("Accept")
	requestPath := c.Request.URL.Path
	return strings.EqualFold(c.GetHeader("X-Requested-With"), "XMLHttpRequest") ||
		strings.Contains(accept, "application/json") ||
		strings.HasPrefix(requestPath, "/api") ||
		strings.HasPrefix(requestPath, "/save")
}

func wantsHTML(c *gin.Context) bool {
	return strings.Contains(c.GetHeader("Accept"), "text/html")
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func detectBaseDir() string {
	cwd, err := os.Getwd()
	if err == nil && hasProjectAssets(cwd) {
		return cwd
	}

	executable, err := os.Executable()
	if err == nil {
		baseDir := filepath.Dir(executable)
		if hasProjectAssets(baseDir) {
			return baseDir
		}
	}

	if err == nil {
		return cwd
	}
	return "."
}

func hasProjectAssets(baseDir string) bool {
	return pathExists(filepath.Join(baseDir, "views")) && pathExists(filepath.Join(baseDir, "public"))
}

func pathExists(target string) bool {
	_, err := os.Stat(target)
	return err == nil
}

func parseJWTExpiresIn(value string) (time.Duration, error) {
	match := expiresInPattern.FindStringSubmatch(strings.TrimSpace(value))
	if len(match) != 3 {
		return 0, fmt.Errorf("unsupported expiresIn value %q", value)
	}

	amount, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return 0, err
	}

	unit := strings.ToLower(match[2])
	multiplier := float64(time.Millisecond)
	switch unit {
	case "", "ms", "msec", "msecs", "millisecond", "milliseconds":
		multiplier = float64(time.Millisecond)
	case "s", "sec", "secs", "second", "seconds":
		multiplier = float64(time.Second)
	case "m", "min", "mins", "minute", "minutes":
		multiplier = float64(time.Minute)
	case "h", "hr", "hrs", "hour", "hours":
		multiplier = float64(time.Hour)
	case "d", "day", "days":
		multiplier = float64(24 * time.Hour)
	case "w", "week", "weeks":
		multiplier = float64(7 * 24 * time.Hour)
	case "y", "yr", "yrs", "year", "years":
		multiplier = float64(time.Duration(365.25 * 24 * float64(time.Hour)))
	default:
		return 0, fmt.Errorf("unsupported expiresIn unit %q", unit)
	}

	duration := time.Duration(amount * multiplier)
	if duration <= 0 {
		return 0, fmt.Errorf("expiresIn must be greater than zero")
	}
	return duration, nil
}

func uuidString() string {
	return uuid.NewString()
}
