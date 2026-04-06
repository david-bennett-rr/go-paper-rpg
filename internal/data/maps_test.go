package data

import "testing"

func TestNormalizeMapPreservesFractionalPropScale(t *testing.T) {
	mapDef := &MapDef{
		Ground: GroundDef{
			Size: [2]float64{10, 10},
		},
		Props: []PropDef{
			{
				ID:       "tree_1",
				Prefab:   "tree",
				Position: [3]float64{2.2, 0, -1.6},
				Scale:    [3]float64{1.25, 0.8, 1.5},
			},
		},
	}

	NormalizeMap(mapDef)

	got := mapDef.Props[0]
	if got.Position[0] != 2 || got.Position[2] != -2 {
		t.Fatalf("expected prop position to snap to tile center, got %+v", got.Position)
	}
	if got.Scale != [3]float64{1.25, 0.8, 1.5} {
		t.Fatalf("expected fractional scale to be preserved, got %+v", got.Scale)
	}
}

func TestNormalizeMapDefaultsZeroPropScaleAxes(t *testing.T) {
	mapDef := &MapDef{
		Ground: GroundDef{
			Size: [2]float64{10, 10},
		},
		Props: []PropDef{
			{
				ID:       "rock_1",
				Prefab:   "rock",
				Position: [3]float64{0, 0, 0},
				Scale:    [3]float64{1.5, 0, 0.75},
			},
		},
	}

	NormalizeMap(mapDef)

	if got := mapDef.Props[0].Scale; got != [3]float64{1.5, 1, 0.75} {
		t.Fatalf("expected zero scale axes to default to 1, got %+v", got)
	}
}
