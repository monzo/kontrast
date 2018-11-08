package diff

import (
	"regexp"
)

type DeltaFilter func([]Delta) []Delta

var sourceRes = []*regexp.Regexp{
	regexp.MustCompile(`apiVersion`),
	regexp.MustCompile(`status.*`),
	regexp.MustCompile(`kind`),
	regexp.MustCompile(`metadata\.creationTimestamp`),
	regexp.MustCompile(`spec\.jobTemplate\.spec\.backoffLimit`),
	regexp.MustCompile(`spec\.template\.spec\.volumes\.[0-9]+\.hostPath\.type`),
}

var serverRes = []*regexp.Regexp{
	regexp.MustCompile(`status.*`),
	regexp.MustCompile(`metadata\.(generation|selfLink|resourceVersion|uid)`),
	regexp.MustCompile(`spec\.template\.spec\.volumes\.[0-9]+\.emptyDir\.sizeLimit`),
	regexp.MustCompile(`spec\.template\.spec\.serviceAccount`),
	regexp.MustCompile(`spec\.ports\.[0-9]+\.nodePort`),
	regexp.MustCompile(`spec\.(clusterIP|volumeName)`),
	regexp.MustCompile(`secrets`),
}

// This function receives a diff between the source and server for a sepcific key
// and returns whether we should care about the delta
//
// Examples:
//   We don't care about this:
//   Source:  {metadata.creationTimestamp <nil>}
//   Server:  {metadata.creationTimestamp 2018-10-15T13:21:32Z}
//
//   We do care about this:
//   Source:  {spec.template.spec.containers.1.image 442690283804.dkr.ecr.eu-west-1.amazonaws.com/monzo/kryp:21069d0}
//   Server:  {spec.template.spec.containers.1.image 442690283804.dkr.ecr.eu-west-1.amazonaws.com/monzo/kryp:1b1d0b3}
//
func shouldKeepMetadata(d Delta) bool {

	//fmt.Printf("Source: %v\nServer: %v\n\n", d.SourceItem, d.ServerItem)

	// Items to ignore in the source control files
	for _, re := range sourceRes {
		if re.MatchString(d.SourceItem.Key) {
			return false
		}
	}

	// Items to ignore from the kubernetes API server
	switch d.ServerItem.Key {
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

func metadataFilter(deltas []Delta) []Delta {
	var filtered []Delta
	for _, d := range deltas {
		if shouldKeepMetadata(d) {
			filtered = append(filtered, d)
		}
	}
	return filtered
}
