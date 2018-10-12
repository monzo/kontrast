package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/milesbxf/kryp/diff"
	"github.com/milesbxf/kryp/pkg/k8s"
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
	changesPresent := false

	filepath.Walk(filename, func(fp string, fi os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err) // can't walk here,
			return nil       // but continue walking elsewhere
		}
		if fi.IsDir() {
			return nil // not a file.  ignore.
		}

		if !strings.HasSuffix(fi.Name(), ".yaml") {
			fmt.Printf("Ignoring %s as it doesn't end in .yaml\n", fp)
			return nil
		}

		fmt.Printf("Checking manifests in %s...\n", fp)
		diffs, err := diff.GetFileDiff(fp, helper)
		if err != nil {
			fatal("error: %f", err)
		}
		for _, d := range diffs {

			switch d.(type) {
			case diff.NotPresentOnServerDiff:
				fmt.Println("Not found")
				changesPresent = true
			case diff.EmptyDiff:
				fmt.Println("No changes")
			case diff.ChangesPresentDiff:
				fmt.Printf("%d changes found:\n", len(d.Deltas()))
				fmt.Println(d.Pretty())
				changesPresent = true
			}
		}
		return nil
	})

	if changesPresent {
		os.Exit(2)
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
