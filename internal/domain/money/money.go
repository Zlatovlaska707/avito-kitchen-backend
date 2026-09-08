package money

import "fmt"

type Minor int64

func LineTotal(unitPrice Minor, quantity int) Minor {
	return unitPrice * Minor(quantity)
}

func Sum(values ...Minor) Minor {
	var total Minor
	for _, v := range values {
		total += v
	}
	return total
}

// ApplyPercent applies percent discount with half-up rounding on the result.
func ApplyPercent(amount Minor, percent int64) Minor {
	if percent <= 0 {
		return amount
	}
	if percent >= 100 {
		return 0
	}
	numerator := int64(amount)*percent + 50
	return Minor(numerator / 100)
}

func (m Minor) String() string {
	return fmt.Sprintf("%d", m)
}
