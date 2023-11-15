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
	wordList []string
	values   passgen.Values
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
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
	var err error
	mode := os.Getenv("GIN_MODE")
	if mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	values, err = loadValues()
	if err != nil {
		log.Fatal("Error loading values:", err)
	}
	once.Do(func() {
		wordList, err = loadWordsFromFile(values.WORDLIST_PATH)
		if err != nil {
			log.Fatalf("Error loading wordlist: %v", err)
		}
	})
	r := gin.New()
	r.Use(PrometheusMiddleware)
	r.Use(gin.Recovery())
	r.Use(cors.Default())

	r.Use(setSecurityHeaders)

	r.Static("/static", "./static")

	r.NoRoute(func(c *gin.Context) {
		c.File("./index.html")
	})

	r.GET("/json", jsonHandler)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      r,
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

	return values, nil
}
func loadWordsFromFile(filename string) ([]string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Printf("File %s not found.\n", filename)
		return nil, err
	}
	return strings.Split(string(data), "\n"), nil
}

func jsonHandler(c *gin.Context) {
	numPasswordsStr := c.Query("num")
	numPasswords, err := strconv.Atoi(numPasswordsStr)
	if err != nil || numPasswords < 1 || numPasswords > maxNumPasswords {
		numPasswords = defaultNumPasswords
	}
	passwords, err := passgen.GeneratePasswords(numPasswords, values, wordList)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err.Error()))
		return
	}
	passwordData = passwords
	jsonBytes, err := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(passwordData)

	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error: %s", err.Error()))
		return
	}

	c.Data(http.StatusOK, "application/json", jsonBytes)
}
