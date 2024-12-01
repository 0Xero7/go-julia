package main

import (
	"github.com/ericlagergren/decimal"
)

type AComplex struct {
	r decimal.Big
	i decimal.Big
}

func New(r float64, i float64) *AComplex {
	var real decimal.Big
	real.SetFloat64(r)

	var img decimal.Big
	img.SetFloat64(i)

	return &AComplex{
		r: real,
		i: img,
	}
}

func Add(a AComplex, other AComplex) AComplex {
	var real decimal.Big
	real.Add(&a.r, &other.r)

	var img decimal.Big
	img.Add(&a.i, &other.i)

	return AComplex{
		r: real,
		i: img,
	}
}

func Mul(a AComplex, other AComplex) AComplex {
	var x1x2 decimal.Big
	x1x2.Mul(&a.r, &other.r)

	var y1y2 decimal.Big
	y1y2.Mul(&a.i, &other.i)

	var x1y2 decimal.Big
	x1y2.Mul(&a.r, &other.i)

	var x2y1 decimal.Big
	x2y1.Mul(&a.i, &other.r)

	var real decimal.Big
	real.Sub(&x1x2, &y1y2)

	var img decimal.Big
	img.Add(&x1y2, &x2y1)

	// fmt.Println(x1.String(), y1.String(), x2.String(), y2.String(), x1x2.String(), y1y2.String(), x1y2.String(), x2y1.String())

	return AComplex{
		r: real,
		i: img,
	}
}

func Gt2(a AComplex) bool {
	x := a.r
	y := a.i

	x.Mul(&x, &x)
	y.Mul(&y, &y)

	var four decimal.Big
	four.SetFloat64(4)

	return x.Add(&x, &y).Cmp(&four) > 0
}
