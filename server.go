package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/errgroup"
)

type ServerConfig struct {
	promRegisterer prometheus.Registerer

	queryUrls *[]string
}

// Config fills in default server config if unset
func (s *ServerConfig) Config() {
	if s.promRegisterer == nil {
		s.promRegisterer = prometheus.DefaultRegisterer
	}

	if s.queryUrls == nil {
		s.queryUrls = &[]string{
			"https://google.com",
			"https://apple.com",
			"https://microsoft.com",
			"https://amazon.com",
			"https://example.com",
		}
	}
}

func runQuery(url string) error {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	totalRequests.Inc()
	client := http.DefaultClient
	resp, err := client.Do(request)
	if err != nil {
		return err
	}

	log.Printf("http response: %d\n", resp.StatusCode)
	return nil
}

func runAllQueries(urls []string) error {
	eg := &errgroup.Group{}
	for i := range urls {
		url := urls[i]
		eg.Go(func() error {
			return runQuery(url)
		})
	}

	return eg.Wait()
}

func NewServer(serverConfig ServerConfig) (*Server, error) {
	serverConfig.Config()

	serverConfig.promRegisterer.Register(totalRequests)

	router := gin.Default()
	router.Use(static.Serve("/", static.LocalFile("./client/build", true)))

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router.GET("/ping", func(c *gin.Context) {
		if err := runAllQueries(*serverConfig.queryUrls); err != nil {
			c.JSON(http.StatusNotFound, gin.H{})
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	srv := &Server{
		&http.Server{
			Addr:    ":8080",
			Handler: router,
		},
	}

	return srv, nil
}

type Server struct {
	*http.Server
}

func (srv *Server) Serve() error {
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
