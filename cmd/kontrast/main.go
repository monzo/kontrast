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
	//removeFromGit := flag.Bool("remove-from-git", false, "Remove files for resources which are not present in kubernetes")
	//removeResources := flag.Bool("remove-resources", false, "Remove resources from kubernetes where a deployment is not present")
	silent := flag.Bool("silent", false, "Hides output")

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

	if deltas := scanForChanges(args[0], config, *onlyShowDeltas, *silent); deltas > 0 {
		os.Exit(2)
	}
}

func scanForChanges(filename string, config *rest.Config, onlyShowDeltas bool, silent bool) int {

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

		var notPresentList []diff.Diff
		var changesPresentList []diff.Diff
		var unchangedList []diff.Diff

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
				notPresentList = append(notPresentList, d)
			case diff.ChangesPresentDiff:
				changesPresent = true
				status = fmt.Sprintf("%d changes", len(d.Deltas()))
				changesPresentList = append(changesPresentList, d)
			case diff.UnchangedDiff:
				changesPresent = false
				unchangedList = append(unchangedList, d)
			}

			// If we want everything OR there are changes
			if !onlyShowDeltas || changesPresent {
				if !silent {
					kind := r.Object.GetObjectKind().GroupVersionKind().Kind
					ref := fmt.Sprintf("%s/%s", r.Namespace, r.Name)
					fmt.Printf("%-50s %-25s %-50s: %s\n\n", ref, kind, fp, status)
					fmt.Println(d.Pretty(colorEnabled))
					totalDeltas++
				}
			}

		}

		makeChanges(fp, notPresentList, changesPresentList, unchangedList)

		return nil
	})

	return totalDeltas
}

func makeChanges(fp string, notPresentList []diff.Diff, changesPresentList []diff.Diff, unchangedList []diff.Diff) {

	if len(notPresentList) > 0 && len(changesPresentList) == 0 {
		// Simple case; All resources are not present
		if len(unchangedList) == 0 {
			// Just remove it from Git
			fmt.Printf("%s\n", fp)
			fmt.Printf("git rm %s\n", fp)

			// There is at least one resource that is not present, and the rest are unchanged
		} else if len(unchangedList) > 0 {
			// Safety check
			// There must not be a deployment or cronjob in `unchangedList`
			// There must be a deployment or cronjob in `notPresentList` to give us a strong signal this removal is intentional
			if !containsCriticalResource(unchangedList) && containsCriticalResource(notPresentList) {
				// Delete unchanged resources
				fmt.Printf("%s\n", fp)
				for _, value := range unchangedList {
					printDelete(value.DiffMeta().Resource)
				}
				// guess we'll need a git rm too, but lets wait for the next run
			}
		}
	}
}

func containsCriticalResource(diffs []diff.Diff) bool {
	for _, value := range diffs {
		if criticalResource(value.DiffMeta().Resource) {
			return true
		}
	}
	return false
}

func criticalResource(resource *k8s.Resource) bool {
	switch resource.Object.GetObjectKind().GroupVersionKind().Kind {
	case "Deployment":
		return true
	case "CronJob":
		return true
	}
	return false
}

// I guess this intended as a dry run output, as k8s.Resource has a delete function
func printDelete(resource *k8s.Resource) {
	namespace := resource.Namespace
	kind := resource.Object.GetObjectKind().GroupVersionKind().Kind
	name := resource.Name
	switch kind {
	case "ConfigMap":
		fmt.Printf("kubectl delete configmap -n %s %s\n", namespace, name)
	case "CronJob":
		fmt.Printf("kubectl delete cronjob -n %s %s\n", namespace, name)
	case "Deployment":
		fmt.Printf("kubectl delete deployment -n %s %s\n", namespace, name)
	case "NetworkPolicy":
		fmt.Printf("kubectl delete networkpolicy -n %s %s\n", namespace, name)
	case "PodDisruptionBudget":
		fmt.Printf("kubectl delete poddisruptionbudget -n %s %s\n", namespace, name)
	case "Service":
		fmt.Printf("kubectl delete service -n %s %s\n", namespace, name)
	case "ServiceAccount":
		fmt.Printf("kubectl delete serviceaccount -n %s %s\n", namespace, name)
	case "VerticalPodAutoscaler":
		fmt.Printf("kubectl delete verticalpodautoscaler -n %s %s\n", namespace, name)
	default:
		fmt.Printf("Unable to delete %s of type %s\n", name, kind)
	}
}

func fatal(msg string, args ...interface{}) {
	fmt.Printf(msg+"\n", args...)
	os.Exit(1)
}
