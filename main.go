package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/milesbxf/petrel/k8s"
)

var (
	kubeconfig *string
	showLogs   *bool
)

func main() {
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	showLogs = flag.Bool("verbose", false, "Show verbose logging")
	flag.Parse()

	if !*showLogs {
		log.SetOutput(ioutil.Discard)
	}

	k8s.LoadConfig(*kubeconfig)

}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
