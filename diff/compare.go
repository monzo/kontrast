package diff

import (
	"bytes"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	k8sjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/yudai/gojsondiff"
)

func jsonDiffToDeltas(keyPrefix string, deltas []Delta, jsonDeltas []gojsondiff.Delta) []Delta {
	for _, d := range jsonDeltas {
		switch d.(type) {
		case *gojsondiff.Added:
			ad := d.(*gojsondiff.Added)
			server := Item{
				Key:   keyPrefix + ad.PostPosition().String(),
				Value: ad.Value,
			}
			deltas = append(deltas, Delta{Item{}, server})
		case *gojsondiff.Deleted:
			dd := d.(*gojsondiff.Deleted)
			source := Item{
				Key:   keyPrefix + dd.Position.String(),
				Value: dd.Value,
			}
			deltas = append(deltas, Delta{source, Item{}})
		case *gojsondiff.Moved:
			md := d.(*gojsondiff.Moved)
			source := Item{
				Key:   keyPrefix + md.PrePosition().String(),
				Value: md.Value,
			}
			server := Item{
				Key:   keyPrefix + md.PostPosition().String(),
				Value: md.Value,
			}
			deltas = append(deltas, Delta{source, server})
		case *gojsondiff.Modified:
			md := d.(*gojsondiff.Modified)
			source := Item{
				Key:   keyPrefix + md.Position.String(),
				Value: md.OldValue,
			}
			server := Item{
				Key:   keyPrefix + md.Position.String(),
				Value: md.NewValue,
			}
			deltas = append(deltas, Delta{source, server})
		case *gojsondiff.Object:
			obj := d.(*gojsondiff.Object)
			deltas = jsonDiffToDeltas(obj.Position.String()+".", deltas, obj.Deltas)
		case *gojsondiff.Array:
			arr := d.(*gojsondiff.Array)
			deltas = jsonDiffToDeltas(arr.Position.String()+".", deltas, arr.Deltas)
		default:
			fmt.Printf("Unknown type %T: %+v\n", d, d)
		}
	}
	return deltas
}

func calculateDiff(objA, objB runtime.Object) ([]Delta, error) {
	JSONDiffer := gojsondiff.New()
	jsonDiff, err := JSONDiffer.Compare(objToJSON(objA), objToJSON(objB))
	if err != nil {
		return []Delta{}, err
	}
	var deltas []Delta
	return jsonDiffToDeltas("", deltas, jsonDiff.Deltas()), nil
}

func objToJSON(obj runtime.Object) []byte {
	s := k8sjson.NewSerializer(k8sjson.DefaultMetaFactory, scheme.Scheme, scheme.Scheme, false)
	dto := &bytes.Buffer{}
	s.Encode(obj, dto)
	return dto.Bytes()
}
