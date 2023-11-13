package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestHandlers(t *testing.T) {
	router := setupRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	testCases := []struct {
		name         string
		path         string
		expectedCode int
	}{
		{"JSONHandler", "/json", http.StatusOK},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", server.URL+tc.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.expectedCode {
				t.Errorf("Expected status code %d for %s, got %d", tc.expectedCode, tc.name, resp.StatusCode)
			}
		})
	}
}

func setupRouter() *gin.Engine {
	r := gin.New()
	r.Use(PrometheusMiddleware)
	r.Use(gin.Recovery())

	r.Static("/assets", "./assets")

	r.GET("/json", jsonHandler)

	return r
}
