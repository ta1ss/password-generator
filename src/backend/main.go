package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"net/http"
	"passgen/passgen"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/logger"
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
		log.Fatal().Err(err).Msg("Error loading values")
	}
	passwordGenerator, err = passgen.NewPasswordGenerator(values)
	if err != nil {
		log.Fatal().Err(err).Msg("Error initiating NewPasswordGenerator")
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
	// Set zerolog to write JSON logs to stdout

	logging_type := os.Getenv("LOGGING_TYPE")
	if logging_type == "json" {
		log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	} else {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
		log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
	}

	mode := os.Getenv("GIN_MODE")
	if mode == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	if logging_type == "json" {
		r.Use(logger.SetLogger(logger.WithLogger(func(_ *gin.Context, l zerolog.Logger) zerolog.Logger {
			return l.Output(gin.DefaultWriter).With().Logger()
		})))
	} else {
		r.Use(logger.SetLogger(logger.WithLogger(func(_ *gin.Context, l zerolog.Logger) zerolog.Logger {
			return l.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Logger()
		})))
	}

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

		log.Info().Msg("Serving metrics on :9090/metrics")
		if err := metricsRouter.Run(":9090"); err != nil {
			log.Fatal().Err(err).Msg("Failed to run metrics server")
		}
	}()

	log.Info().Msg("Starting server on :8080")
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

	setValuesFromEnv()

	return values, nil
}

func setValuesFromEnv() {
	r := reflect.ValueOf(&values).Elem()

	for i := 0; i < r.NumField(); i++ {
		field := r.Type().Field(i)
		envValue := os.Getenv(field.Name)

		if envValue != "" {
			switch field.Type.Kind() {
			case reflect.String:
				r.Field(i).SetString(envValue)
			case reflect.Int:
				intValue, err := strconv.Atoi(envValue)
				if err == nil {
					r.Field(i).SetInt(int64(intValue))
	}
			}
		}
	}
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
