// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package freeroam

import (
	"bytes"
	"fmt"
	"strconv"
)

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func cStrLen(n []byte) int {
	for i := 0; i < len(n); i++ {
		if n[i] == 0 {
			return i
		}
	}
	return len(n)
}

func printBinary(b []byte) {
	var o bytes.Buffer
	for _, b := range b {
		fm := strconv.FormatInt(int64(b), 2)
		for i := 0; i < (8 - len(fm)); i++ {
			o.WriteRune('0')
		}
		o.WriteString(fm)
		o.WriteRune(' ')
	}
	fmt.Println(o.String())
}
