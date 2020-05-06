// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package freeroam

import "math"

// Vector represents a two-dimensional vector
type Vector struct {
	X float64
	Y float64
}

// Distance returns Euclidean distance between two Vectors
func Distance(a, b Vector) float64 {
	xd := a.X - b.X
	yd := a.Y - b.Y
	return math.Sqrt(xd*xd + yd*yd)
}
