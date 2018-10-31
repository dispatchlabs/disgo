package utils

type Point struct {
	X float64
	Y float64
}

func LinearRegression(points *[]Point) (a float64, b float64) {
	n := float64(len(*points))

	sumX := 0.0
	sumY := 0.0
	sumXY := 0.0
	sumXX := 0.0

	for _, p := range *points {
		sumX += p.X
		sumY += p.Y
		sumXY += p.X * p.Y
		sumXX += p.X * p.X
	}

	base := (n*sumXX - sumX*sumX)
	a = (n*sumXY - sumX*sumY) / base
	b = (sumXX*sumY - sumXY*sumX) / base

	return a, b
}


