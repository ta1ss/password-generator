package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"

	"net/http"
	"passgen/passgen"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v3"
)

const (
	defaultNumPasswords = 1
	maxNumPasswords     = 1000
)

var (
	passwordData      []passgen.Password
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Number of HTTP requests",
		},
		[]string{"path", "method"},
	)
	passwordGenerator *passgen.PasswordGenerator
	values            passgen.Values
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	values, err := loadValues()
	if err != nil {
		log.Fatal("Error loading values:", err)
	}
	passwordGenerator, err = passgen.NewPasswordGenerator(values)
	if err != nil {
		log.Fatal("Error initiating NewPasswordGenerator:", err)
	}
}

func PrometheusMiddleware(c *gin.Context) {
	path := c.FullPath()
	method := c.Request.Method
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		httpRequestsTotal.WithLabelValues(path, method).Inc()
	}))
	defer timer.ObserveDuration()
	c.Next()
}

func setSecurityHeaders(c *gin.Context) {
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-Frame-Options", "DENY")
	c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; frame-ancestors 'none'; form-action 'self';")
}

func main() {

	mode := os.Getenv("GIN_MODE")
	if mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()
	r.Use(PrometheusMiddleware)
	r.Use(gin.Recovery())
	r.Use(cors.Default())

	r.Use(setSecurityHeaders)

	r.UseH2C = true // Enable HTTP/2

	r.Static("/static", "./static")

	r.NoRoute(func(c *gin.Context) {
		c.File("./index.html")
	})

	r.GET("/json", jsonHandler)

	r.GET("/api/v1/passwords", jsonHandler)

	r.GET("/api/v1/config/maxPasswordLength", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"maxPasswordLength": values.MAX_PASSWORD_LENGTH})
	})
	r.GET("/api/v1/config/minPasswordLength", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"minPasswordLength": values.MIN_PASSWORD_LENGTH})
	})
	server := &http.Server{
		Addr:         ":8080",
		Handler:      r.Handler(),
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	go func() {
		metricsRouter := gin.New()
		metricsRouter.GET("/metrics", gin.WrapH(promhttp.Handler()))

		fmt.Printf("Serving metrics on :9090/metrics\n")
		metricsRouter.Run(":9090")
	}()

	fmt.Printf("Starting server on :8080\n")
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

func loadValues() (passgen.Values, error) {

	yamlFile, err := os.ReadFile("values/values.yaml")
	if err != nil {
		return values, err
	}

	err = yaml.Unmarshal(yamlFile, &values)
	if err != nil {
		return values, err
	}

	// Override with environment variables
	values.MIN_PASSWORD_LENGTH = getEnvAsInt("MIN_PASSWORD_LENGTH", values.MIN_PASSWORD_LENGTH)
	values.MAX_PASSWORD_LENGTH = getEnvAsInt("MAX_PASSWORD_LENGTH", values.MAX_PASSWORD_LENGTH)
	values.BETWEEN_SYMBOLS = getEnv("BETWEEN_SYMBOLS", values.BETWEEN_SYMBOLS)
	values.INSIDE_SYMBOLS = getEnv("INSIDE_SYMBOLS", values.INSIDE_SYMBOLS)
	values.PASSWORD_PER_ROUTINE = getEnvAsInt("PASSWORD_PER_ROUTINE", values.PASSWORD_PER_ROUTINE)
	values.WORDLIST_PATH = getEnv("WORDLIST_PATH", values.WORDLIST_PATH)

	return values, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if len(valueStr) == 0 {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getQueryParameterAsInt(c *gin.Context, key string, defaultValue int) int {
	valueStr := c.Query(key)
	if len(valueStr) == 0 {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func jsonHandler(c *gin.Context) {
	numPasswords := getQueryParameterAsInt(c, "num", defaultNumPasswords)
	minPasswordLength := getQueryParameterAsInt(c, "minPasswordLength", values.MIN_PASSWORD_LENGTH)
	maxPasswordLength := getQueryParameterAsInt(c, "maxPasswordLength", values.MAX_PASSWORD_LENGTH)

	// Validate input
	if numPasswords < 1 || numPasswords > maxNumPasswords {
		c.String(http.StatusBadRequest, fmt.Sprintf("Invalid number of passwords: %d", numPasswords))
		return
	}
	if minPasswordLength < values.MIN_PASSWORD_LENGTH {
		c.String(http.StatusBadRequest, fmt.Sprintf("Invalid minPasswordLength: %d", minPasswordLength))
		return
	}

	if minPasswordLength > maxPasswordLength {
		c.String(http.StatusBadRequest, fmt.Sprintf("Minimal password length must be smaller than maximal length. minPasswordLength:%d; maxPasswordLength:%d.", minPasswordLength, maxPasswordLength))
		return
	}
	if maxPasswordLength > values.MAX_PASSWORD_LENGTH {
		c.String(http.StatusBadRequest, fmt.Sprintf("Invalid maxPasswordLength: %d", maxPasswordLength))
		return
	}
	passwords, err := passwordGenerator.GeneratePasswords(numPasswords, minPasswordLength, maxPasswordLength)

	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error generating passwords: %s", err.Error()))
		return
	}

	passwordData = passwords
	jsonBytes, err := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(passwordData)

	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error forming JSON: %s", err.Error()))
		return
	}

	c.Data(http.StatusOK, "application/json", jsonBytes)
}
