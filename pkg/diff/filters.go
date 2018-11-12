package diff

import (
	"regexp"
)

var filters = []*regexp.Regexp{
	regexp.MustCompile(`apiVersion`),
	regexp.MustCompile(`kind`),
	regexp.MustCompile(`metadata\.(creationTimestamp|generation|selfLink|resourceVersion|uid)`),
	regexp.MustCompile(`spec\.jobTemplate\.spec\.backoffLimit`),
	regexp.MustCompile(`spec\.template\.spec\.volumes\.[0-9]+\.hostPath\.type`),
	regexp.MustCompile(`spec\.template\.spec\.volumes\.[0-9]+\.emptyDir\.sizeLimit`),
	regexp.MustCompile(`spec\.template\.spec\.serviceAccount`),
	regexp.MustCompile(`spec\.ports\.[0-9]+\.nodePort`),
	regexp.MustCompile(`spec\.(clusterIP|volumeName)`),
	regexp.MustCompile(`secrets`),
	regexp.MustCompile(`status.*`),
}

var serverRes = []*regexp.Regexp{}

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

	// Ignore anything in the filter list
	for _, re := range filters {
		if re.MatchString(d.SourceItem.Key) || re.MatchString(d.ServerItem.Key) {
			return false
		}
	}

	// Special cases for things that harder to filter with a regex :-)
	switch d.ServerItem.Key {
	case "metadata.annotations":
		anns, ok := d.ServerItem.Value.(map[string]interface{})
		if ok {
			if anns["kubectl.kubernetes.io/last-applied-configuration"] != struct{}{} || anns["pod.alpha.kubernetes.io/init-containers"] != struct{}{} {
				return false
			}
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
