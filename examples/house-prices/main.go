// Command house-prices fits a line through past sales (size against price)
// with least-squares linear regression and uses it to price new listings,
// the simplest model behind valuation tools.
package main

import (
	"fmt"
	"log"

	"github.com/braswelljr/alGOrithm/pkg/ml/regression"
)

func main() {
	sqm := []float64{50, 65, 80, 95, 110, 130, 150}
	priceK := []float64{150, 190, 232, 270, 318, 372, 425}
	slope, intercept, err := regression.Linear(sqm, priceK)
	if err != nil {
		log.Fatal(err)
	}
	preds := make([]float64, len(sqm))
	for i, x := range sqm {
		preds[i] = slope*x + intercept
	}
	r2, _ := regression.R2(priceK, preds)
	fmt.Printf("price = %.2fk per m2 * size + %.1fk (R^2 %.3f)\n", slope, intercept, r2)
	for _, size := range []float64{70, 120} {
		fmt.Printf("  %3.0f m2 listing: about $%.0fk\n", size, slope*size+intercept)
	}
}
