package logger

import "testing"

func TestNewLogger(t *testing.T) {
	for _, env := range []string{"development", "production"} {
		log := NewLogger(env)
		if log == nil || log.SugaredLogger == nil {
			t.Fatalf("logger nil pentru %s", env)
		}
		log.Infof("mesaj de test (%s)", env)
	}
}
