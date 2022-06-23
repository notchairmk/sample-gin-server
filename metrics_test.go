package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

type mockRegisterer struct {
	Registry *prometheus.Registry
}

func (r *mockRegisterer) Register(c prometheus.Collector) error {
	r.Registry = prometheus.NewRegistry()
	return r.Registry.Register(c)
}

func (r *mockRegisterer) MustRegister(c ...prometheus.Collector) {
	r.Registry = prometheus.NewRegistry()
	r.Registry.MustRegister(c...)
}

func (r *mockRegisterer) Unregister(c prometheus.Collector) bool {
	r.Registry = prometheus.NewRegistry()
	return r.Registry.Unregister(c)
}

func responseHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("some stuff"))
}

func TestMetricsCount_IntegrationTest(t *testing.T) {
	mockHttpServer := httptest.NewServer(http.HandlerFunc(responseHandler))
	registerer := mockRegisterer{}

	config := ServerConfig{
		promRegisterer: &registerer,

		queryUrls: &[]string{
			mockHttpServer.URL,
		},
	}
	server, err := NewServer(config)
	if err != nil {
		t.Fatalf("unable to initialize server")
	}

	go func() {
		if err := server.Serve(); err != nil {
			t.Logf("received error: %v", err)
		}
	}()

	_, err = http.Get("http://localhost:8080/ping")
	if err != nil {
		t.Fatal(err)
	}

	metrics, err := registerer.Registry.Gather()
	if err != nil {
		t.Fatal(err)
	}

	got := *metrics[0].Metric[0].Counter.Value
	t.Logf("got metric: %f", got)
	expected := float64(1)
	if got != expected {
		t.Fatalf("expected %f, got %f", expected, got)
	}

	server.Shutdown(context.TODO())
}
