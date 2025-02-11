// Copyright 2016 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package execinfrapb

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/eval"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/leaktest"
)

type testVarContainer struct {
	used []bool
}

var _ tree.IndexedVarContainer = &testVarContainer{}

func (d *testVarContainer) IndexedVarResolvedType(idx int) *types.T {
	// Track usage of the IndexedVar.
	for idx > len(d.used)-1 {
		d.used = append(d.used, false)
	}
	d.used[idx] = true
	return types.Int
}

func (d *testVarContainer) IndexedVarNodeFormatter(idx int) tree.NodeFormatter {
	n := tree.Name(fmt.Sprintf("var%d", idx))
	return &n
}

func (d *testVarContainer) wasUsed(idx int) bool {
	if idx < len(d.used) {
		return d.used[idx]
	}
	return false
}

func TestProcessExpression(t *testing.T) {
	defer leaktest.AfterTest(t)()

	e := Expression{Expr: "@1 * (@2 + @3) + @1"}

	var c testVarContainer
	h := tree.MakeIndexedVarHelper(&c, 4)
	st := cluster.MakeTestingClusterSettings()
	evalCtx := eval.MakeTestingEvalContext(st)
	semaCtx := tree.MakeSemaContext()
	expr, err := processExpression(context.Background(), e, &evalCtx, &semaCtx, &h)
	if err != nil {
		t.Fatal(err)
	}

	if !c.wasUsed(0) || !c.wasUsed(1) || !c.wasUsed(2) || c.wasUsed(3) {
		t.Errorf("invalid IndexedVarUsed results %t %t %t %t (expected false false false true)",
			c.wasUsed(0), c.wasUsed(1), c.wasUsed(2), c.wasUsed(3))
	}

	str := expr.String()
	expectedStr := "(var0 * (var1 + var2)) + var0"
	if str != expectedStr {
		t.Errorf("invalid expression string '%s', expected '%s'", str, expectedStr)
	}

	// We can process a new expression with the same tree.IndexedVarHelper.
	e = Expression{Expr: "@4 - @1"}
	expr, err = processExpression(context.Background(), e, &evalCtx, &semaCtx, &h)
	if err != nil {
		t.Fatal(err)
	}

	str = expr.String()
	expectedStr = "var3 - var0"
	if str != expectedStr {
		t.Errorf("invalid expression string '%s', expected '%s'", str, expectedStr)
	}
}

// Test that processExpression evaluates constant exprs into datums.
func TestProcessExpressionConstantEval(t *testing.T) {
	defer leaktest.AfterTest(t)()

	e := Expression{Expr: "ARRAY[1:::INT,2:::INT]"}

	h := tree.MakeIndexedVarHelper(nil, 0)
	st := cluster.MakeTestingClusterSettings()
	evalCtx := eval.MakeTestingEvalContext(st)
	semaCtx := tree.MakeSemaContext()
	expr, err := processExpression(context.Background(), e, &evalCtx, &semaCtx, &h)
	if err != nil {
		t.Fatal(err)
	}

	expected := &tree.DArray{
		ParamTyp:    types.Int,
		Array:       tree.Datums{tree.NewDInt(1), tree.NewDInt(2)},
		HasNonNulls: true,
	}
	if !reflect.DeepEqual(expr, expected) {
		t.Errorf("invalid expr '%v', expected '%v'", expr, expected)
	}
}
