package models

import (
	"encoding/json"
	"testing"
)

func TestIntakeMoneyExactReaisAndCents(t *testing.T) {
	for raw, cents := range map[string]int64{"79.9": 7990, "10": 1000, "0": 0, "0.01": 1, "999999999999.99": 99999999999999} {
		var m IntakeMoney
		if err := json.Unmarshal([]byte(raw), &m); err != nil {
			t.Fatal(err)
		}
		if int64(m) != cents {
			t.Fatalf("%s became %d cents", raw, m)
		}
		encoded, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		var roundTrip IntakeMoney
		if err = json.Unmarshal(encoded, &roundTrip); err != nil || roundTrip != m {
			t.Fatalf("round trip %s: %v", encoded, err)
		}
	}
	for _, raw := range []string{`-1`, `79.999`, `"79.90"`, `1e3`, `1000000000000`, `true`} {
		var m IntakeMoney
		if json.Unmarshal([]byte(raw), &m) == nil {
			t.Errorf("accepted invalid money %s", raw)
		}
	}
}
