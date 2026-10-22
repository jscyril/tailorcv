package resumeimport

import "testing"

func TestParseRange(t *testing.T) {
	for _, test := range []struct {
		value, start, end string
		current, ok       bool
	}{
		{"June 2020 – Present", "2020-06", "", true, true},
		{"Aug. 2018 – May 2021", "2018-08", "2021-05", false, true},
		{"Sept 2019 to Dec 2019", "2019-09", "2019-12", false, true},
		{"Summer 2021", "2021-06", "2021-06", false, true},
		{"01/2020 - 03/2021", "2020-01", "2021-03", false, true},
		{"2019 – 2021", "2019-01", "2021-12", false, true},
		{"Expected May 2025", "2025-05", "2025-05", false, true},
		{"Spring", "", "", false, false},
		{"13/2020", "", "", false, false},
	} {
		start, end, current, ok := parseRange(test.value)
		if start != test.start || end != test.end || current != test.current || ok != test.ok {
			t.Errorf("parseRange(%q) = %q, %q, %v, %v; want %q, %q, %v, %v", test.value, start, end, current, ok, test.start, test.end, test.current, test.ok)
		}
	}
}
