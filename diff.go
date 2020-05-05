// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package freeroam

type slotInfo struct {
	Client         *Client
	LastUpdateID   uint8
	UpdateACKed    bool
	HasSentFull    bool
	ACKMissedCount uint8
	PacketSentSeq  uint16

	LastCPTime uint16
}

type ArrayDiffResult struct {
	Kept    []*Client
	Added   []*Client
	Removed []*Client
}

func ArrayDiff(old []*Client, new []*Client) ArrayDiffResult {
	res := ArrayDiffResult{
		Kept:    make([]*Client, 0),
		Added:   make([]*Client, 0),
		Removed: make([]*Client, 0),
	}
	for _, av := range new {
		if has(old, av) {
			res.Kept = append(res.Kept, av)
		} else {
			res.Added = append(res.Added, av)
		}
	}
	for _, av := range old {
		if !has(new, av) {
			res.Removed = append(res.Removed, av)
		}
	}
	return res
}

func has(a []*Client, v *Client) bool {
	for _, av := range a {
		if av == v {
			return true
		}
	}
	return false
}
