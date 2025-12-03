package kit_test

import (
	"testing"

	"nova/internal/kit"

	"github.com/stretchr/testify/assert"
)

type KeyValue struct {
	Name  string
	Value string
}

func TestUniqBy(t *testing.T) {
	tests := []struct {
		name  string
		slice []*KeyValue
		want  []*KeyValue
	}{
		{
			name: "test uniq by",
			slice: []*KeyValue{
				{Name: "a", Value: "1"},
				{Name: "b", Value: "2"},
				{Name: "a", Value: "1"},
				{Name: "b", Value: "2"},
				{Name: "a", Value: "2"},
			},
			want: []*KeyValue{
				{Name: "a", Value: "1"},
				{Name: "b", Value: "2"},
				{Name: "a", Value: "2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := kit.UniqBy(tt.slice, func(i *KeyValue) string { return i.Name + ":" + i.Value })
			assert.Equal(t, tt.want, got)
		})
	}
}
