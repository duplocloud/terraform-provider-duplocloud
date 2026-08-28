package duplosdk

import "testing"

func TestSortLbConfigurationsByLbIndex(t *testing.T) {
	lbs := []DuploLbConfiguration{
		{LbIndex: 3, Protocol: "http", Port: "8080"},
		{LbIndex: 1, Protocol: "http", Port: "80"},
		{LbIndex: 2, Protocol: "tcp", Port: "443"},
	}

	sortLbConfigurations(lbs)

	for i, want := range []int{1, 2, 3} {
		if lbs[i].LbIndex != want {
			t.Errorf("position %d: got LbIndex %d, want %d", i, lbs[i].LbIndex, want)
		}
	}
}

func TestSortLbConfigurationsWithoutLbIndex(t *testing.T) {
	// Portals without the UseLbIndex feature leave LbIndex zero on every config; the
	// fallback keys must still give a deterministic order.
	lbs := []DuploLbConfiguration{
		{LbType: 1, Protocol: "tcp", Port: "443"},
		{LbType: 1, Protocol: "http", Port: "8080"},
		{LbType: 0, Protocol: "http", Port: "80"},
		{LbType: 1, Protocol: "http", Port: "3000"},
	}

	sortLbConfigurations(lbs)

	got := make([]string, 0, len(lbs))
	for _, lb := range lbs {
		got = append(got, lb.Port)
	}
	want := []string{"80", "3000", "8080", "443"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got order %v, want %v", got, want)
		}
	}
}
