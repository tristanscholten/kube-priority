package annotation

import "testing"

func TestParse(t *testing.T) {
	bounds := Bounds{Minimum: -2147483648, MaximumExclusive: 1000000000}
	tests := []struct {
		name, raw string
		want      int32
		ok        bool
	}{
		{"positive", "100", 100, true}, {"zero", "0", 0, true}, {"negative", "-10", -10, true}, {"upper-valid", "999999999", 999999999, true},
		{"upper-invalid", "1000000000", 0, false}, {"above", "1000000001", 0, false}, {"int32-overflow", "2147483648", 0, false}, {"int32-underflow", "-2147483649", 0, false},
		{"empty", "", 0, false}, {"leading-space", " 1", 0, false}, {"trailing-space", "1 ", 0, false}, {"inside-space", "1 2", 0, false}, {"plus", "+1", 0, false}, {"decimal", "1.0", 0, false}, {"hex", "0x10", 0, false}, {"bad", "abc", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.raw, bounds)
			if tt.ok && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tt.ok && err == nil {
				t.Fatalf("expected error")
			}
			if tt.ok && got.Value != tt.want {
				t.Fatalf("got %d want %d", got.Value, tt.want)
			}
		})
	}
}

func TestParseMinimum(t *testing.T) {
	_, err := Parse("-1", Bounds{Minimum: 0, MaximumExclusive: 1000000000})
	if err == nil {
		t.Fatal("expected stricter minimum rejection")
	}
}
