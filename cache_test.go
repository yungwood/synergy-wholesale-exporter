package main

import "testing"

func TestAPIErrorBackoffSeconds(t *testing.T) {
	originalMin := *apiErrorBackoffMin
	originalMax := *apiErrorBackoffMax
	t.Cleanup(func() {
		*apiErrorBackoffMin = originalMin
		*apiErrorBackoffMax = originalMax
	})

	*apiErrorBackoffMin = 10
	*apiErrorBackoffMax = 60

	tests := []struct {
		name       string
		errorCount int
		want       int64
	}{
		{name: "first failure", errorCount: 0, want: 10},
		{name: "second failure", errorCount: 1, want: 20},
		{name: "third failure", errorCount: 2, want: 40},
		{name: "capped", errorCount: 3, want: 60},
		{name: "remains capped", errorCount: 10, want: 60},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := apiErrorBackoffSeconds(test.errorCount); got != test.want {
				t.Errorf("backoff = %d, want %d", got, test.want)
			}
		})
	}
}

func TestAPIErrorBackoffSecondsHandlesInvalidConfig(t *testing.T) {
	originalMin := *apiErrorBackoffMin
	originalMax := *apiErrorBackoffMax
	t.Cleanup(func() {
		*apiErrorBackoffMin = originalMin
		*apiErrorBackoffMax = originalMax
	})

	*apiErrorBackoffMin = 0
	*apiErrorBackoffMax = 0

	if got := apiErrorBackoffSeconds(0); got != 1 {
		t.Errorf("backoff = %d, want 1", got)
	}
}
