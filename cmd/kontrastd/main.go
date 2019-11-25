package main

import (
	"flag"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/monzo/kontrast/pkg/k8s"
	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	kubeconfig *string
	addr       = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
	interval   = flag.String("interval", "1m", "How often to refresh diffs")
)

func main() {
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.Parse()
	args := flag.Args()

	if len(args) != 1 {
		flag.Usage()
		log.Fatalf("Error: requires positional argument for directory/file to check")
	}

	filename := args[0]

	intervalDuration, err := time.ParseDuration(*interval)
	if err != nil {
		log.Fatalf("Could not parse --interval: %s", err.Error())
	}

	config, err := k8s.LoadConfig(*kubeconfig)
	if err != nil {
		log.Info("config load error")
		log.Fatalf("error: %f", err)
	}

	dm, err := NewDiffManager(config)
	if err != nil {
		log.Fatalf("error: %f", err)
	}

	// Set up the Prometheus collector
	collector := NewKontrastCollector(dm)
	prometheus.MustRegister(collector)

	go dm.DiffRun(filename)
	updateTicker := time.NewTicker(intervalDuration)
	go func() {
		for range updateTicker.C {
			dm.DiffRun(filename)
		}
	}()
	defer updateTicker.Stop()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./assets/static"))))
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", handleDiffDisplay(dm, filename))

	log.Infof("Listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
