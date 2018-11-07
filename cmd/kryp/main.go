package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/monzo/kryp/pkg/diff"
	"github.com/monzo/kryp/pkg/k8s"
	"k8s.io/client-go/rest"
)

var (
	kubeconfig   *string
	colorEnabled bool
)

func main() {

	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	colorDisabled := flag.Bool("no-color", false, "Disables ANSI colour output")

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

	if deltas := scanForChanges(args[0], config); deltas > 0 {
		os.Exit(2)
	}
}

func scanForChanges(filename string, config *rest.Config) int {

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
			//fmt.Printf("Ignoring %s as it doesn't end in .yaml\n", fp)
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
				if deltas := len(d.Deltas()); deltas > 0 {
					changesPresent = true
					status = fmt.Sprintf("%d changes", deltas)
				}
			}

			if changesPresent {
				kind := r.Object.GetObjectKind().GroupVersionKind().Kind
				ref := fmt.Sprintf("%s/%s", r.Namespace, r.Name)
				fmt.Printf("%-50s %-25s %-50s: %s\n", ref, kind, fp, status)
				fmt.Println(d.Pretty(colorEnabled))
				totalDeltas++
			}
		}
		return nil
	})

	return totalDeltas
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
