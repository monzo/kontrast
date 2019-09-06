package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/monzo/kontrast/pkg/diff"
	"github.com/monzo/kontrast/pkg/k8s"
	"k8s.io/client-go/rest"
)

var (
	kubeconfig   *string
	colorEnabled bool
)

func main() {
	defaultKubeConfig := os.Getenv("KUBECONFIG")
	if defaultKubeConfig == "" {
		if home, err := os.UserHomeDir(); err == nil {
			defaultKubeConfig = filepath.Join(home, ".kube", "config")
		}
	}

	kubeconfig = flag.String("kubeconfig", defaultKubeConfig, "(optional) absolute path to the kubeconfig file")
	colorDisabled := flag.Bool("no-color", false, "Disables ANSI colour output")
	onlyShowDeltas := flag.Bool("deltas-only", true, "Only show files with changes")

	flag.Parse()
	args := flag.Args()

	colorEnabled = !*colorDisabled

	if len(args) != 1 {
		flag.Usage()
		fatal("Error: requires positional argument for directory/file to check")
	}

	config, err := k8s.LoadConfig(*kubeconfig)
	if err != nil {
		fatal("error: %f", err)
	}

	fmt.Println()

	if deltas := scanForChanges(args[0], config, *onlyShowDeltas); deltas > 0 {
		os.Exit(2)
	}
}

func scanForChanges(filename string, config *rest.Config, onlyShowDeltas bool) int {

	helper, err := k8s.NewResourceHelperWithDefaults(config)
	if err != nil {
		fatal("error: %f", err)
	}

	log.SetOutput(ioutil.Discard)

	totalDeltas := 0
	filepath.Walk(filename, func(fp string, fi os.FileInfo, err error) error {

		if err != nil {
			fmt.Println(err) // can't walk here,
			return nil
		}

		if fi.IsDir() {
			return nil // not a file. ignore.
		}

		if !strings.HasSuffix(fi.Name(), ".yaml") {
			return nil
		}

		resources, err := helper.NewResourcesFromFilename(fp)
		if err != nil {
			fmt.Printf("Error getting resource: %v\n", err)
			return nil
		}

		for _, r := range resources {
			changesPresent := false

			d, err := diff.GetDiffsForResource(r, helper)
			if err != nil {
				fmt.Printf("Error getting resource: %v\n", err)
				return nil
			}

			var status string
			switch d.(type) {
			case diff.NotPresentOnServerDiff:
				status = "not found on server"
				changesPresent = true
			case diff.ChangesPresentDiff:
				changesPresent = len(d.Deltas()) > 0
				status = fmt.Sprintf("%d changes", len(d.Deltas()))
			}

			// If we want everything OR there are changes
			if !onlyShowDeltas || changesPresent {
				kind := r.Object.GetObjectKind().GroupVersionKind().Kind
				ref := fmt.Sprintf("%s/%s", r.Namespace, r.Name)
				fmt.Printf("%-50s %-25s %-50s: %s\n\n", ref, kind, fp, status)
				fmt.Println(d.Pretty(colorEnabled))
				totalDeltas++
			}
		}
		return nil
	})

	return totalDeltas
}

func fatal(msg string, args ...interface{}) {
	fmt.Printf(msg+"\n", args...)
	os.Exit(1)
}
