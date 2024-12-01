package main

import "github.com/ericlagergren/decimal"

func AddDecimal(a decimal.Big, b decimal.Big) decimal.Big {
	var result decimal.Big
	result.Add(&a, &b)
	return result
}

func SubDecimal(a decimal.Big, b decimal.Big) decimal.Big {
	var result decimal.Big
	result.Sub(&a, &b)
	return result
}

func MulDecimal(a decimal.Big, b decimal.Big) decimal.Big {
	var result decimal.Big
	result.Mul(&a, &b)
	return result
}

func CompareDecimal(a decimal.Big, b decimal.Big) int {
	return a.Cmp(&b)
}

func Dec(a float64) decimal.Big {
	var result decimal.Big
	result.SetFloat64(a)
	return result
}
