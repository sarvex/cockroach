// Copyright 2021 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package util

// CombineUniqueString combines two ordered string slices and returns
// the result without duplicates.
func CombineUniqueString(a []string, b []string) []string {
	res := make([]string, 0, len(a))
	aIter, bIter := 0, 0
	for aIter < len(a) && bIter < len(b) {
		if a[aIter] == b[bIter] {
			res = append(res, a[aIter])
			aIter++
			bIter++
		} else if a[aIter] < b[bIter] {
			res = append(res, a[aIter])
			aIter++
		} else {
			res = append(res, b[bIter])
			bIter++
		}
	}
	if aIter < len(a) {
		res = append(res, a[aIter:]...)
	}
	if bIter < len(b) {
		res = append(res, b[bIter:]...)
	}
	return res
}
