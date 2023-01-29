package random

import (
	"fmt"
	"math/rand"
)

// Generates a number between min and max inclusive.
func Range(min, max int) int {
	if min > max {
		panic(fmt