package action

import (
	"testing"
)

func TestQualityFromDistance(t *testing.T) {
	windowSize := 12 // 12 tick window

	tests := []struct {
		name     string
		distance int
		want     Quality
	}{
		{"exact sweet spot", 0, QualityExcellent},
		{"very close", 1, QualityGreat},
		{"medium distance", 3, QualityGood},
		{"far but in window", 5, QualityNice},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := qualityFromDistance(tt.distance, windowSize)
			if got != tt.want {
				t.Errorf("qualityFromDistance(%d, %d) = %v, want %v", tt.distance, windowSize, got, tt.want)
			}
		})
	}
}

func TestResultFromQuality(t *testing.T) {
	tests := []struct {
		quality  Quality
		wantMult float64
		wantOK   bool
	}{
		{QualityMiss, 1.0, false},
		{QualityNice, 1.25, true},
		{QualityGood, 1.5, true},
		{QualityGreat, 1.75, true},
		{QualityExcellent, 2.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.quality.String(), func(t *testing.T) {
			r := ResultFromQuality(tt.quality)
			if r.BonusMult != tt.wantMult {
				t.Errorf("BonusMult = %f, want %f", r.BonusMult, tt.wantMult)
			}
			if r.Success != tt.wantOK {
				t.Errorf("Success = %v, want %v", r.Success, tt.wantOK)
			}
		})
	}
}
