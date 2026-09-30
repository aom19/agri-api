package config

import "testing"

func TestLoadConfig_DefaultsAndOverrides(t *testing.T) {
	t.Setenv("APP_NAME", "")
	t.Setenv("SERVER_PORT", "9090")
	t.Setenv("REDIS_DB", "3")
	t.Setenv("SMTP_PORT", "nu-e-numar")
	t.Setenv("WEATHER_SNAPSHOT_INTERVAL", "0")

	cfg := LoadConfig()
	if cfg.AppName != "agri-api" || cfg.AppEnv != "development" {
		t.Errorf("valorile implicite: %+v", cfg)
	}
	if cfg.ServerPort != "9090" || cfg.RedisDB != 3 || cfg.WeatherSnapshotInterval != "0" {
		t.Errorf("valorile din mediu nu au fost preluate: %+v", cfg)
	}
	if cfg.SMTPPort != 1025 {
		t.Errorf("un întreg invalid trebuie să cadă pe valoarea implicită, am %d", cfg.SMTPPort)
	}
	if cfg.DBHost != "localhost" || cfg.AccessTokenTTL != "15m" || cfg.FieldOperationOverdueCheckInterval != "1m" {
		t.Errorf("alte valori implicite: %+v", cfg)
	}
}

func TestGetEnvHelpers(t *testing.T) {
	t.Setenv("AGRI_TEST_STR", "x")
	t.Setenv("AGRI_TEST_INT", "7")
	if getEnv("AGRI_TEST_STR", "d") != "x" || getEnv("AGRI_TEST_MISSING", "d") != "d" {
		t.Error("getEnv")
	}
	if getEnvInt("AGRI_TEST_INT", 1) != 7 || getEnvInt("AGRI_TEST_MISSING", 1) != 1 {
		t.Error("getEnvInt")
	}
}
