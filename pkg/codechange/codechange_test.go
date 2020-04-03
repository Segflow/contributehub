package codechange

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChangeFile(t *testing.T) {
	tt := map[string]struct {
		Before  string
		After   string
		Changes []CodeChange
		Fail    bool
	}{
		"sample1": {
			Before: "testdata/sample1.go.before",
			After:  "testdata/sample1.go.after",
			Changes: []CodeChange{
				{
					Line:   7,
					Column: 15,
					Offset: 47,
					Add:    []byte("<-"),
				},
				{
					Line:   11,
					Column: 12,
					Offset: 78,
					Add:    []byte("<-"),
				},
			},
			Fail: false,
		}}

	for _, tc := range tt {
		out, err := FileApplyChanges(tc.Before, tc.Changes)
		if tc.Fail {
			assert.Error(t, err)
		}

		content, _ := ioutil.ReadFile(tc.After)
		assert.Equal(t, content, out)
	}
}
