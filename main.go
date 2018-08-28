package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/milesbxf/kryp/diff"
	"github.com/milesbxf/kryp/k8s"
)

var (
	kubeconfig *string
)

func main() {
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	if len(os.Args) != 2 {
		fatal("Program takes a single argument")
	}
	log.SetOutput(ioutil.Discard)

	filename := os.Args[1]

	config, err := k8s.LoadConfig(*kubeconfig)
	if err != nil {
		fatal("error: %f", err)
	}

	helper, err := k8s.NewResourceHelperWithDefaults(config)
	if err != nil {
		fatal("error: %f", err)
	}

	d, err := diff.GetFileDiff(filename, helper)
	if err != nil {
		fatal("error: %f", err)
	}

	switch d.(type) {
	case diff.NotPresentOnServerDiff:
		fmt.Println("Not found")
	case diff.EmptyDiff:
		fmt.Println("No changes")
	case diff.ChangesPresentDiff:
		fmt.Printf("%d changes found:\n", len(d.Deltas()))
		fmt.Println(d.Pretty())
	}

}

func fatal(msg string, args ...interface{}) {
	fmt.Printf(msg+"\n", args)
	os.Exit(1)
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
