package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"net/http"
	"passgen/passgen"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
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
	//c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; frame-ancestors 'none'; form-action 'self';")
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
	wordList, err = loadWordsFromFile(values.WORDLIST_PATH)
	if err != nil {
		log.Fatalf("Error loading wordlist: %v", err)
	}

	// Create the main listener.
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Create a cmux instance
	m := cmux.New(lis)

	// Match connections in order: first gRPC, then HTTP.
	grpcL := m.Match(cmux.HTTP2HeaderField("content-type", "application/grpc"))
	httpL := m.Match(cmux.Any())

	// Create a gRPC server
	grpcServer := grpc.NewServer()
	passgen.RegisterPasswordGeneratorServer(grpcServer, &server{})

	// Wrap the gRPC server with gRPC-Web wrapper
	wrappedGrpc := grpcweb.WrapServer(grpcServer)

	r := gin.New()
	r.Use(PrometheusMiddleware)
	r.Use(gin.Recovery())
	r.Use(cors.Default())
	r.Use(setSecurityHeaders)
	r.Static("/static", "./static")
	r.StaticFile("/robots.txt", "./robots.txt")
	r.NoRoute(func(c *gin.Context) {
		c.File("./index.html")
	})
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	server := &http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if wrappedGrpc.IsGrpcWebRequest(req) {
				// This is a gRPC-Web request, handle it using the gRPC-Web wrapper
				wrappedGrpc.ServeHTTP(w, req)
			} else {
				// This is a regular HTTP request, handle it using Gin
				r.ServeHTTP(w, req)
			}
		}),
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}

	// Use goroutines to serve different protocols concurrently
	go grpcServer.Serve(grpcL)
	go server.Serve(httpL)

	go func() {
		metricsRouter := gin.New()
		metricsRouter.GET("/metrics", gin.WrapH(promhttp.Handler()))

		fmt.Printf("Serving metrics on :9090/metrics\n")
		metricsRouter.Run(":9090")
	}()

	// Start cmux serving
	if err := m.Serve(); err != nil {
		log.Fatalf("cmux serve failed: %v", err)
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

type server struct {
	passgen.UnimplementedPasswordGeneratorServer
}

func (s *server) GetPasswords(ctx context.Context, req *passgen.GenerateRequest) (*passgen.GenerateResponse, error) {
	numPasswords := req.GetLen()
	if numPasswords < 1 || numPasswords > maxNumPasswords {
		numPasswords = defaultNumPasswords
	}

	passwords, err := passgen.GeneratePasswords(int(numPasswords), values, wordList)
	if err != nil {
		return nil, err
	}

	var grpcPasswords []*passgen.Message
	for _, pwd := range passwords {
		grpcPasswords = append(grpcPasswords, &passgen.Message{
			Xkcd:     pwd.Xkcd,
			Original: pwd.Original,
			Length:   int32(pwd.Length),
		})
	}

	return &passgen.GenerateResponse{Passwords: grpcPasswords}, nil
}
