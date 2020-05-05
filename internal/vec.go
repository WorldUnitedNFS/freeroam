// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package internal

import "math"

type Vector2D struct {
	X float64
	Y float64
}

func (self Vector2D) Sub(vec Vector2D) Vector2D {
	return Vector2D{
		X: self.X - vec.X,
		Y: self.Y - vec.Y,
	}
}

func (self Vector2D) Abs() Vector2D {
	return Vector2D{
		X: math.Abs(self.X),
		Y: math.Abs(self.Y),
	}
}

func (self Vector2D) Length() float64 {
	tsqrt := rootF64(self.X) + rootF64(self.Y)
	return math.Sqrt(tsqrt)
}

func absI64(i int64) int64 {
	if i < 0 {
		return -i
	}
	return i
}

func rootI64(i int64) int64 {
	return i * i
}

func rootF64(i float64) float64 {
	return i * i
}
