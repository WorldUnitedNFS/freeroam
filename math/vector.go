// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
package math

import "math"

// Vector2D represents a two-dimensional vector
type Vector2D struct {
	X float64
	Y float64
}

// Vector3D represents a three-dimensional vector
type Vector3D struct {
	X float64
	Y float64
	Z float64
}

// Distance returns Euclidean distance between two Vectors
func Distance(a, b Vector2D) float64 {
	xd := a.X - b.X
	yd := a.Y - b.Y
	return math.Sqrt(xd*xd + yd*yd)
}
