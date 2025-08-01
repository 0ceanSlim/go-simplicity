// examples/simple_multisig.go
//go:build ignore
// +build ignore

package main

// MultiSigValidation simulates a 2-of-3 multisig
func MultiSigValidation(sig1Valid bool, sig2Valid bool, sig3Valid bool) bool {
	validSigs := 0

	if sig1Valid {
		validSigs = validSigs + 1
	}

	if sig2Valid {
		validSigs = validSigs + 1
	}

	if sig3Valid {
		validSigs = validSigs + 1
	}

	// Need at least 2 valid signatures
	return validSigs >= 2
}

// CheckAmount validates transaction amount
func CheckAmount(amount uint64) bool {
	const maxAmount uint64 = 1000000 // 1M sats
	const minAmount uint64 = 1000    // 1K sats

	return amount >= minAmount && amount <= maxAmount
}

// MultiSigPayment validates a multisig payment
func MultiSigPayment(
	amount uint64,
	sig1Valid bool,
	sig2Valid bool,
	sig3Valid bool,
) bool {
	// Validate amount first
	if !CheckAmount(amount) {
		return false
	}

	// Check multisig condition
	return MultiSigValidation(sig1Valid, sig2Valid, sig3Valid)
}

func main() {
	amount := uint64(50000)
	sig1 := true  // Alice signed
	sig2 := true  // Bob signed
	sig3 := false // Charlie didn't sign

	result := MultiSigPayment(amount, sig1, sig2, sig3)

	if result {
		// Payment authorized
		return
	} else {
		// Payment rejected
		return
	}
}
