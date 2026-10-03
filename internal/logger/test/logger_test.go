package logger_test

import (
	"testing"

	"agri-api/internal/logger"
)

func TestNewLogger(t *testing.T) {
	for _, env := range []string{"development", "production"} {
		log := logger.NewLogger(env)
		if log == nil || log.SugaredLogger == nil {
			t.Fatalf("logger nil pentru %s", env)
		}
		log.Infof("mesaj de test (%s)", env)
	}
}
