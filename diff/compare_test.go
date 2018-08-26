package diff

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yudai/gojsondiff"
)

func jd(a, b string) []gojsondiff.Delta {
	JSONDiffer := gojsondiff.New()
	jsonDiff, err := JSONDiffer.Compare([]byte(a), []byte(b))
	if err != nil {
		panic(err)
	}
	return jsonDiff.Deltas()
}

func TestJSONDiffToDeltas(t *testing.T) {
	cases := []struct {
		desc     string
		A        string
		B        string
		ExpDelta []Delta
	}{
		{"no deltas when JSON is the same",
			`{"keyA": 1}`, `{"keyA": 1}`, []Delta{}},
		{"one delta with value modified",
			`{"keyA": 1}`, `{"keyA": 2}`, []Delta{
				Delta{Item{"keyA", 1.}, Item{"keyA", 2.}}}},
		{"one delta with key added",
			`{"keyA": 1}`, `{"keyA": 1, "keyB": 2}`, []Delta{
				Delta{Item{}, Item{"keyB", 2.}}}},
		{"one delta with key deleted",
			`{"keyA": 1, "keyB": 2}`, `{"keyA": 1}`, []Delta{
				Delta{Item{"keyB", 2.}, Item{}}}},
		{"one delta with key modified",
			`{"keyA": 1}`, `{"keyB": 1}`, []Delta{
				Delta{Item{"keyA", 1.}, Item{}},
				Delta{Item{}, Item{"keyB", 1.}}}},
		{"no deltas when nested JSON is the same",
			`{"keyA": 1, "nested": {"keyB": 2}}`,
			`{"keyA": 1, "nested": {"keyB": 2}}`, []Delta{}},
		{"one deltas when nested JSON value changed",
			`{"keyA": 1, "nested": {"keyB": 2}}`,
			`{"keyA": 1, "nested": {"keyB": 3}}`, []Delta{
				Delta{Item{"nested.keyB", 2.}, Item{"nested.keyB", 3.}},
			}},
	}

	for _, c := range cases {
		actual := []Delta{}
		actual = jsonDiffToDeltas("", actual, jd(c.A, c.B))
		assert.Equal(t, c.ExpDelta, actual, "expected "+c.desc)
	}

}
