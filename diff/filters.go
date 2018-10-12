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
}

func shouldKeepMetadata(d Delta) bool {
	switch d.SourceItem.Key {
	case "apiVersion", "kind", "metadata.namespace":
		return false
	}
	for _, re := range sourceRes {
		if re.MatchString(d.SourceItem.Key) {
			return false
		}
	}
	switch d.ServerItem.Key {
	case "metadata.generation",
		"metadata.selfLink",
		"spec.template.spec.serviceAccount":
		return false
	case "metadata.annotations":
		anns, ok := d.ServerItem.Value.(map[string]interface{})
		if ok && anns["kubectl.kubernetes.io/last-applied-configuration"] != struct{}{} {
			return false
		}
	}
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
