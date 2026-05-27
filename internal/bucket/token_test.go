package bucket

import (
	"testing"
	"time"
)

func TestNewBucket(t *testing.T) {
	cases := []struct {
		name       string
		capacity   float64
		refillRate float64
		wantErr    bool
	}{
		{name: "valid input", capacity: 100, refillRate: 10, wantErr: false},
		{name: "zero capacity", capacity: 0, refillRate: 10, wantErr: true},
		{name: "zero ", capacity: 100, refillRate: 0, wantErr: true},
		{name: "negative capacity", capacity: -5, refillRate: 0, wantErr: true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := NewBucket(testCase.capacity, testCase.refillRate)
			if (err != nil) != testCase.wantErr {
				t.Errorf("NewBucket(%f, %f)errors out with %v, wantErr %t", testCase.capacity, testCase.refillRate, err, testCase.wantErr)
			}
		})
	}
}

func TestAllow(t *testing.T) {
	cases := []struct {
		name       string
		cost       int
		want       bool
		capacity   float64
		refillRate float64
	}{
		{name: "allow", cost: 10, want: true, capacity: 100, refillRate: 10},
		{name: "deny", cost: 1000, want: false, capacity: 100, refillRate: 10},
		{name: "boundary condition", cost: 100, want: true, capacity: 100, refillRate: 10},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := NewBucket(tc.capacity, tc.refillRate)
			if err != nil {
				t.Fatalf("bucket failed")
			}

			got := b.Allow(tc.cost)
			if got != tc.want {
				t.Errorf("Allow(%d) = %t, want %t", tc.cost, got, tc.want)
			}

		})
	}
}

// TestRefillAfterDepletion tests tokens repopulating after using all of the tokens
// Allow(1) will fail after depletion, but aftet waiting succeed.
func TestRefillAfterDepletion(t *testing.T) {
	b, _ := NewBucket(10, 10)
	b.Allow(10)
	b.Allow(1)
	if b.Allow(1) {
		t.Fatal("expected bucket to be empty after draining")
	}
	time.Sleep(200 * time.Millisecond)
	if !b.Allow(1) {
		t.Errorf("expected Allow(1) to succeed after refill got denied")
	}
}
