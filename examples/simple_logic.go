//go:build ignore
// +build ignore

package main

func CheckConditions(a bool, b bool) bool {
	return a && b
}

func ProcessAmount(amount uint64) bool {
	const minAmount uint64 = 1000
	return amount >= minAmount
}

func main() {
	result1 := CheckConditions(true, false)
	result2 := ProcessAmount(5000)

	if result1 || result2 {
		return
	}
}
