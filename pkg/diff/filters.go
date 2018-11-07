package diff

import (
	"regexp"
)

type DeltaFilter func([]Delta) []Delta

var sourceRes = []*regexp.Regexp{
	regexp.MustCompile(`spec\.template\.spec\.volumes\.[0-9]+\.hostPath\.type`),
}
var serverRes = []*regexp.Regexp{
	regexp.MustCompile(`spec\.template\.spec\.volumes\.[0-9]+\.emptyDir\.sizeLimit`),
	regexp.MustCompile(`spec\.ports\.[0-9]+\.nodePort`),
	regexp.MustCompile(`status.*`),
}

func shouldKeepMetadata(d Delta) bool {

	// Items to ignore in the source control files
	switch d.SourceItem.Key {
	case "apiVersion",
		"kind",
		"status.phase",
		"metadata.creationTimestamp",
		"metadata.namespace",
		"spec.jobTemplate.spec.backoffLimit":
		return false
	}

	// As above but looking at the source regex's
	for _, re := range sourceRes {
		if re.MatchString(d.SourceItem.Key) {
			return false
		}
	}

	// Items to ignore from the kubernetes API server
	switch d.ServerItem.Key {
	case "metadata.generation",
		"metadata.selfLink",
		"metadata.resourceVersion",
		"metadata.uid",
		"spec.clusterIP",
		"secrets",
		"spec.volumeName",
		"spec.template.spec.serviceAccount":
		return false
	case "metadata.annotations":
		anns, ok := d.ServerItem.Value.(map[string]interface{})
		if ok && anns["kubectl.kubernetes.io/last-applied-configuration"] != struct{}{} {
			return false
		}
	}

	// As above but looking at the server regex's
	for _, re := range serverRes {
		if re.MatchString(d.ServerItem.Key) {
			return false
		}
	}
	return true
}

func MetadataFilter(deltas []Delta) []Delta {
	var filtered []Delta
	for _, d := range deltas {
		if shouldKeepMetadata(d) {
			filtered = append(filtered, d)
		}
	}
	return filtered
}
