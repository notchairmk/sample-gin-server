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

func runQuery(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	log.Printf("http response: %d\n", resp.StatusCode)
	return nil
}

func runAllQueries(urls []string) error {
	eg := &errgroup.Group{}
	for _, url := range urls {
		totalRequests.Inc()
		egUrl := url
		eg.Go(func() error {
			return runQuery(egUrl)
		})
	}

	return eg.Wait()
}

func NewServer() (*Server, error) {
	prometheus.Register(totalRequests)

	router := gin.Default()
	router.Use(static.Serve("/", static.LocalFile("./client/build", true)))

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	urls := []string{
		"https://google.com",
		"https://apple.com",
		"https://microsoft.com",
		"https://amazon.com",
		"https://example.com",
	}

	router.GET("/ping", func(c *gin.Context) {
		if err := runAllQueries(urls); err != nil {
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
