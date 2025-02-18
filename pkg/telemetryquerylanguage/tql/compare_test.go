// Copyright  The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tql

import (
	"fmt"
	"testing"
)

// Our types are bool, int, float, string, Bytes, nil, so we compare all types in both directions.
var (
	ta   = false
	tb   = true
	sa   = "1"
	sb   = "2"
	sn   = ""
	ba   = []byte("1")
	bb   = []byte("2")
	bn   []byte
	i64a = int64(1)
	i64b = int64(2)
	f64a = float64(1)
	f64b = float64(2)
)

type testA struct {
	Value string
}

type testB struct {
	Value string
}

// This is one giant table-driven test that is designed to check almost all of the possibilities for
// comparing two different kinds of items. Each test compares two objects of any type using all
// of the comparison operators, and has a hand-constructed list of expected values.
// The test is not *quite* exhaustive, but it does attempt to check every basic type against
// every other basic type, and includes a pretty good set of tests on the pointers to all the
// basic types as well.
func Test_compare(t *testing.T) {
	tests := []struct {
		name string
		a    any
		b    any
		want []bool // in order of EQ, NE, LT, LTE, GTE, GT.
	}{
		{"identity string", sa, sa, []bool{true, false, false, true, true, false}},
		{"identity int64", i64a, i64a, []bool{true, false, false, true, true, false}},
		{"identity float64", f64a, f64a, []bool{true, false, false, true, true, false}},
		{"identity bytes", ba, ba, []bool{true, false, false, true, true, false}},
		{"identity nil", nil, nil, []bool{true, false, false, true, true, false}},

		{"diff strings", sa, sb, []bool{false, true, true, true, false, false}},
		{"string bytes", sa, ba, []bool{false, true, false, false, false, false}},
		{"string int64", sa, i64a, []bool{false, true, false, false, false, false}},
		{"string float64", sa, f64a, []bool{false, true, false, false, false, false}},
		{"string nil", sa, nil, []bool{false, true, false, false, false, false}},

		{"diff bytes", ba, bb, []bool{false, true, true, true, false, false}},
		{"bytes string", ba, sa, []bool{false, true, false, false, false, false}},
		{"bytes int64", ba, i64a, []bool{false, true, false, false, false, false}},
		{"bytes float64", ba, f64a, []bool{false, true, false, false, false, false}},
		{"bytes nil", ba, nil, []bool{false, true, false, false, false, false}},
		{"bytes nilbytes", ba, bn, []bool{false, true, false, false, false, false}},

		{"false true", ta, tb, []bool{false, true, true, true, false, false}},
		{"true false", tb, ta, []bool{false, true, false, false, true, true}},
		{"true true", ta, ta, []bool{true, false, false, true, true, false}},
		{"false false", ta, ta, []bool{true, false, false, true, true, false}},

		{"bool string", ta, sa, []bool{false, true, false, false, false, false}},
		{"bool int64", ta, i64a, []bool{false, true, false, false, false, false}},
		{"bool float64", ta, f64a, []bool{false, true, false, false, false, false}},
		{"bool and nil bytes", ta, bn, []bool{false, true, false, false, false, false}},
		{"bool and empty string", ta, sn, []bool{false, true, false, false, false, false}},

		{"nil string", nil, sa, []bool{false, true, false, false, false, false}},
		{"nil int64", nil, i64a, []bool{false, true, false, false, false, false}},
		{"nil float64", nil, f64a, []bool{false, true, false, false, false, false}},
		{"nil and nil bytes", nil, bn, []bool{true, false, false, true, true, false}},
		{"nil and empty string", nil, sn, []bool{false, true, false, false, false, false}},

		{"int64 string", i64a, sa, []bool{false, true, false, false, false, false}},
		{"int64 bytes", i64a, ba, []bool{false, true, false, false, false, false}},
		{"int64 nil", i64a, nil, []bool{false, true, false, false, false, false}},
		{"int64 int64", i64a, i64b, []bool{false, true, true, true, false, false}},
		{"int64 float64", i64a, f64b, []bool{false, true, true, true, false, false}},

		{"diff float64s", f64a, f64b, []bool{false, true, true, true, false, false}},
		{"float64 string", f64a, sa, []bool{false, true, false, false, false, false}},
		{"float64 bytes", f64a, ba, []bool{false, true, false, false, false, false}},
		{"float64 nil", f64a, nil, []bool{false, true, false, false, false, false}},
		{"float64 int64", f64a, i64b, []bool{false, true, true, true, false, false}},

		{"non-prim, same type, equal", testA{"hi"}, testA{"hi"}, []bool{true, false, false, false, false, false}},
		{"non-prim, same type, not equal", testA{"hi"}, testA{"byte"}, []bool{false, true, false, false, false, false}},
		{"non-prim, diff type", testA{"hi"}, testB{"hi"}, []bool{false, true, false, false, false, false}},
		{"non-prim, int type", testA{"hi"}, 5, []bool{false, true, false, false, false, false}},
		{"int, non-prim", 5, testA{"hi"}, []bool{false, true, false, false, false, false}},
	}
	ops := []CompareOp{EQ, NE, LT, LTE, GTE, GT}
	for _, tt := range tests {
		for _, op := range ops {
			t.Run(fmt.Sprintf("%s %v", tt.name, op), func(t *testing.T) {
				if got := compare(tt.a, tt.b, op); got != tt.want[op] {
					t.Errorf("compare(%v, %v, %v) = %v, want %v", tt.a, tt.b, op, got, tt.want[op])
				}
			})
		}
	}
}

// Benchmarks -- these benchmarks compare the performance of comparisons of a variety of data types.
// It's not attempting to be exhaustive, but again, it hits most of the major types and combinations.
// The summary is that they're pretty fast; all the calls to compare are 12 ns/op or less on a 2019 intel
// mac pro laptop, and none of them have any allocations.
func BenchmarkCompareEQInt64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		compare(i64a, i64b, EQ)
	}
}

func BenchmarkCompareEQFloat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		compare(f64a, f64b, EQ)
	}
}
func BenchmarkCompareEQString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		compare(sa, sb, EQ)
	}
}
func BenchmarkCompareEQPString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		compare(&sa, &sb, EQ)
	}
}
func BenchmarkCompareEQBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		compare(ba, bb, EQ)
	}
}
func BenchmarkCompareEQNil(b *testing.B) {
	for i := 0; i < b.N; i++ {
		compare(nil, nil, EQ)
	}
}
func BenchmarkCompareNEInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		compare(i64a, i64b, NE)
	}
}

func BenchmarkCompareNEFloat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		compare(f64a, f64b, NE)
	}
}
func BenchmarkCompareNEString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		compare(sa, sb, NE)
	}
}
func BenchmarkCompareLTFloat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		compare(f64a, f64b, LT)
	}
}
func BenchmarkCompareLTString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		compare(sa, sb, LT)
	}
}
func BenchmarkCompareLTNil(b *testing.B) {
	for i := 0; i < b.N; i++ {
		compare(nil, nil, LT)
	}
}

// this is only used for benchmarking, and is a rough equivalent of the original compare function
// before adding LT, LTE, GTE, and GT.
func compareEq(a any, b any, op CompareOp) bool {
	switch op {
	case EQ:
		return a == b
	case NE:
		return a != b
	default:
		return false
	}
}

func BenchmarkCompareEQFunction(b *testing.B) {
	for i := 0; i < b.N; i++ {
		compareEq(sa, sb, EQ)
	}
}
