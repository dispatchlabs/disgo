package main

import "fmt"

type Point struct {
	x float64
	y float64
}

func LinearRegression(points *[]Point) (a float64, b float64) {
	n := float64(len(*points))

	sumX := 0.0
	sumY := 0.0
	sumXY := 0.0
	sumXX := 0.0

	for _, p := range *points {
		sumX += p.x
		sumY += p.y
		sumXY += p.x * p.y
		sumXX += p.x * p.x
	}

	base := (n*sumXX - sumX*sumX)
	a = (n*sumXY - sumX*sumY) / base
	b = (sumXX*sumY - sumXY*sumX) / base

	return a, b
}

// EXAMPLE USAGE
// points := make([]Point, 0)

// points = append(points, Point{x:0.0, y:1.0,})
// points = append(points, Point{x:0.1, y:1.5,})
// points = append(points, Point{x:0.2, y:2.0,})
// points = append(points, Point{x:0.3, y:2.5,})
// points = append(points, Point{x:0.4, y:3.0,})
// points = append(points, Point{x:0.5, y:3.5,})
// points = append(points, Point{x:0.6, y:4.0,})

// a, b := LeastSquaresMethod(&points)
