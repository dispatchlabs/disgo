package types

import (
	"testing"
	"math"
	"fmt"
)


func TestHerz(t *testing.T) {
	hertz := Hertz{Db: db}
	hertz.Merge()
}

func TestGrowth(t *testing.T) {
	MaxTTL := 86400  //nbr seconds in a day
	//MinTTL := 1
	UppertTxThreshold := 1000

	// 86.4 = 1 ^ gf
	expGrowthFactor := math.Log2E * float64((MaxTTL / UppertTxThreshold))
	value := fmt.Sprintf("%v\n", expGrowthFactor)
	fmt.Printf(value)
	fmt.Printf("%f\n", 1000 * expGrowthFactor)
	fmt.Printf("%f\n", 500 * expGrowthFactor)
	fmt.Printf("%f\n", 250 * expGrowthFactor)
	fmt.Printf("%f\n", 125 * expGrowthFactor)
	fmt.Printf("%f\n", 75 * expGrowthFactor)
	fmt.Printf("%f\n", 25 * expGrowthFactor)
	fmt.Printf("%f\n", 10 * expGrowthFactor)

}