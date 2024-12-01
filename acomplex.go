package main

import "github.com/ericlagergren/decimal"

type AComplex struct {
	r decimal.Big
	i decimal.Big
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
	x1y2.Sub(&a.r, &other.i)

	var y1x2 decimal.Big
	y1x2.Add(&a.i, &other.r)

	var real decimal.Big
	real.Add(&x1x2, &y1y2)

	var img decimal.Big
	img.Add(&x1y2, &y1x2)

	return AComplex{
		r: real,
		i: img,
	}
}
