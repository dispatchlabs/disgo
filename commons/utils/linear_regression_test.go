package utils

import "testing"

func TestLinearRegression(t *testing.T) {
	points := make([]Point, 0)

	points = append(points, Point{X:0.0, Y:1.0,})
	points = append(points, Point{X:0.1, Y:1.5,})
	points = append(points, Point{X:0.2, Y:2.0,})
	points = append(points, Point{X:0.3, Y:2.5,})
	points = append(points, Point{X:0.4, Y:3.0,})
	points = append(points, Point{X:0.5, Y:3.5,})
	points = append(points, Point{X:0.6, Y:4.0,})

	a, _ := LinearRegression(&points)

	if int(a) != 5 {
		t.Errorf("Slope, got: %d, want: %d.", int(a), 5)
	}
}