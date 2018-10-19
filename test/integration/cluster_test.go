package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/monzo/kryp/diff"
	"github.com/monzo/kryp/k8s"
	"github.com/stretchr/testify/assert"
)

func testHelper(t *testing.T) *k8s.ResourceHelper {
	kubeconfigFilename := filepath.Join(os.Getenv("HOME"), ".kube", "config")

	config, err := k8s.LoadConfig(kubeconfigFilename)
	if err != nil {
		t.Fatal(err)
	}

	helper, err := k8s.NewResourceHelperWithDefaults(config)
	if err != nil {
		t.Fatal(err)
	}
	return helper
}

func TestLiveCluster_Diff(t *testing.T) {
	helper := testHelper(t)

	testDeployManifest := filepath.Join("testdata", "nginx.yaml")
	resource, err := helper.NewResourceFromFilename(testDeployManifest)
	if err != nil {
		t.Fatal(err)
	}

	d, err := diff.GetFileDiff(testDeployManifest, helper)
	if err != nil {
		t.Fatal(err)
	}
	assert.IsType(t, diff.NotPresentOnServerDiff{}, d)

	err = resource.Create()
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = resource.Delete()
		if err != nil {
			t.Fatal(err)
		}
	}()

	d, err = diff.GetFileDiff(testDeployManifest, helper)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range d.Deltas() {
		fmt.Printf("%#v\n", d)
	}

	assert.IsType(t, diff.ChangesPresentDiff{}, d)

}

// func TestLiveCluster_ResourceDoesntExist(t *testing.T) {
// 	helper := testHelper(t)

// 	path := filepath.Join("testdata", "nginx.yaml")
// 	diff.GetFileDiff(path)

// }

/* test cases
- resource doesn't exist
- resource exists, but no diffs
- resource exists, diff
*/
