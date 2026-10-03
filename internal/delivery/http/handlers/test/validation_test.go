package handlers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"agri-api/internal/delivery/http/handlers"
)

func TestValidationErrors(t *testing.T) {
	type payload struct {
		Email  string `validate:"required,email"`
		Name   string `validate:"min=3"`
		Code   string `validate:"max=2"`
		Status string `validate:"oneof=active inactive"`
		Age    string `validate:"numeric"`
	}
	err := validator.New().Struct(payload{Name: "ab", Code: "abc", Status: "x", Age: "abc"})
	errs := handlers.ValidationErrors(err)
	expected := map[string]string{
		"Email":  "obligatoriu",
		"Name":   "minim 3",
		"Code":   "maxim 2",
		"Status": "unul din: active inactive",
		"Age":    "invalid (numeric)",
	}
	for field, fragment := range expected {
		if !strings.Contains(errs[field], fragment) {
			t.Errorf("%s: %q nu conține %q", field, errs[field], fragment)
		}
	}
	err = validator.New().Struct(payload{Email: "nu-e-email", Name: "abc", Code: "ab", Status: "active", Age: "1"})
	if !strings.Contains(handlers.ValidationErrors(err)["Email"], "email valid") {
		t.Errorf("email invalid: %v", handlers.ValidationErrors(err))
	}

	generic := handlers.ValidationErrors(errors.New("json invalid"))
	if generic["error"] != "json invalid" {
		t.Errorf("eroare generică: %v", generic)
	}
}

func TestHealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)
	handlers.HealthCheck(c)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"status":"ok"`) {
		t.Errorf("HealthCheck: %d %s", rec.Code, rec.Body.String())
	}
}
