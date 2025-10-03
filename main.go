package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	eventCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kubernetes_events_total",
			Help: "Number of Kubernetes events observed, grouped by namespace, reason, and type",
		},
		[]string{"namespace", "reason", "type"},
	)

	lokiURL    string
	httpClient = &http.Client{Timeout: 5 * time.Second}
)

// Loki log entry
type LokiEntry struct {
	Streams []struct {
		Stream map[string]string `json:"stream"`
		Values [][]string        `json:"values"`
	} `json:"streams"`
}

func pushToLoki(event *v1.Event) {
	if lokiURL == "" {
		return
	}

	labels := map[string]string{
		"namespace": event.Namespace,
		"reason":    event.Reason,
		"type":      event.Type,
		"component": event.Source.Component,
	}

	entry := LokiEntry{
		Streams: []struct {
			Stream map[string]string `json:"stream"`
			Values [][]string        `json:"values"`
		}{
			{
				Stream: labels,
				Values: [][]string{
					{
						fmt.Sprintf("%d000000000", time.Now().Unix()), // ns precision
						fmt.Sprintf("[%s] %s: %s", event.Type, event.Reason, event.Message),
					},
				},
			},
		},
	}

	payload, _ := json.Marshal(entry)

	req, err := http.NewRequest("POST", lokiURL+"/loki/api/v1/push", bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("failed to create loki request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("failed to push to loki: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		log.Printf("loki returned non-2xx status: %d", resp.StatusCode)
	}
}

func main() {
	listenAddr := flag.String("listen", ":8080", "HTTP listen address")
	flag.StringVar(&lokiURL, "loki-url", os.Getenv("LOKI_URL"), "Loki push API base URL")
	flag.Parse()

	prometheus.MustRegister(eventCounter)

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Error building in-cluster config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}

	// Watch events
	watchlist := cache.NewListWatchFromClient(
		clientset.CoreV1().RESTClient(),
		"events",
		metav1.NamespaceAll,
		fields.Everything(),
	)

	_, controller := cache.NewInformer(
		watchlist,
		&v1.Event{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				event := obj.(*v1.Event)
				eventCounter.WithLabelValues(event.Namespace, event.Reason, event.Type).Inc()
				pushToLoki(event)
			},
		},
	)

	stopCh := make(chan struct{})
	go controller.Run(stopCh)

	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Starting server on %s", *listenAddr)
	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
