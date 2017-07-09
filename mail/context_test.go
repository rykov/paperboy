package mail

import (
	"reflect"
	"testing"
)

func TestContextFlattenMap(t *testing.T) {
	input := map[string]interface{}{
		"Level1a": "1",
		"Level1b": map[string]interface{}{
			"Level2a": "2",
			"Level2b": map[string]interface{}{
				"Level3": "3",
			},
		},
	}

	expected := map[string]interface{}{
		"Level1a":                "1",
		"Level1b.Level2a":        "2",
		"Level1b.Level2b.Level3": "3",
	}

	if out := flattenMap(input); !reflect.DeepEqual(out, expected) {
		t.Errorf("Output mismatch:\nExpected:%v\nActual:%v", expected, out)
	}
}
