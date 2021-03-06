// Copyright 2019 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// {{/*
// +build execgen_template
//
// This file is the execgen template for selection_ops.eg.go. It's formatted in
// a special way, so it's both valid Go and a valid text/template input. This
// permits editing this file with editor support.
//
// */}}

package colexec

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/col/typeconv"
	"github.com/cockroachdb/cockroach/pkg/sql/colconv"
	"github.com/cockroachdb/cockroach/pkg/sql/colexec/execgen"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecbase"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecbase/colexecerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
)

// {{/*
// Declarations to make the template compile properly.

// _LEFT_CANONICAL_TYPE_FAMILY is the template variable.
const _LEFT_CANONICAL_TYPE_FAMILY = types.UnknownFamily

// _LEFT_TYPE_WIDTH is the template variable.
const _LEFT_TYPE_WIDTH = 0

// _RIGHT_CANONICAL_TYPE_FAMILY is the template variable.
const _RIGHT_CANONICAL_TYPE_FAMILY = types.UnknownFamily

// _RIGHT_TYPE_WIDTH is the template variable.
const _RIGHT_TYPE_WIDTH = 0

// _ASSIGN_CMP is the template function for assigning the result of comparing
// the second input to the third input into the first input.
func _ASSIGN_CMP(_, _, _, _, _, _ interface{}) int {
	colexecerror.InternalError(errors.AssertionFailedf(""))
}

// */}}

// {{/*
func _SEL_CONST_LOOP(_HAS_NULLS bool) { // */}}
	// {{define "selConstLoop" -}}
	// {{$hasNulls := $.HasNulls}}
	// {{with $.Overload}}
	if sel := batch.Selection(); sel != nil {
		sel = sel[:n]
		for _, i := range sel {
			var cmp bool
			arg := col.Get(i)
			_ASSIGN_CMP(cmp, arg, p.constArg, _, col, _)
			// {{if _HAS_NULLS}}
			isNull = nulls.NullAt(i)
			// {{else}}
			isNull = false
			// {{end}}
			if cmp && !isNull {
				sel[idx] = i
				idx++
			}
		}
	} else {
		batch.SetSelection(true)
		sel := batch.Selection()
		_ = col.Get(n - 1)
		for i := 0; i < n; i++ {
			var cmp bool
			arg := col.Get(i)
			_ASSIGN_CMP(cmp, arg, p.constArg, _, col, _)
			// {{if _HAS_NULLS}}
			isNull = nulls.NullAt(i)
			// {{else}}
			isNull = false
			// {{end}}
			if cmp && !isNull {
				sel[idx] = i
				idx++
			}
		}
	}
	// {{end}}
	// {{end}}
	// {{/*
} // */}}

// {{/*
func _SEL_LOOP(_HAS_NULLS bool) { // */}}
	// {{define "selLoop" -}}
	// {{$hasNulls := $.HasNulls}}
	// {{with $.Overload}}
	if sel := batch.Selection(); sel != nil {
		sel = sel[:n]
		for _, i := range sel {
			var cmp bool
			arg1 := col1.Get(i)
			arg2 := col2.Get(i)
			_ASSIGN_CMP(cmp, arg1, arg2, _, col1, col2)
			// {{if _HAS_NULLS}}
			isNull = nulls.NullAt(i)
			// {{else}}
			isNull = false
			// {{end}}
			if cmp && !isNull {
				sel[idx] = i
				idx++
			}
		}
	} else {
		batch.SetSelection(true)
		sel := batch.Selection()
		_ = col1.Get(n - 1)
		_ = col2.Get(n - 1)
		for i := 0; i < n; i++ {
			var cmp bool
			arg1 := col1.Get(i)
			arg2 := col2.Get(i)
			_ASSIGN_CMP(cmp, arg1, arg2, _, col1, col2)
			// {{if _HAS_NULLS}}
			isNull = nulls.NullAt(i)
			// {{else}}
			isNull = false
			// {{end}}
			if cmp && !isNull {
				sel[idx] = i
				idx++
			}
		}
	}
	// {{end}}
	// {{end}}
	// {{/*
} // */}}

// selConstOpBase contains all of the fields for binary selections with a
// constant, except for the constant itself.
type selConstOpBase struct {
	OneInputNode
	colIdx         int
	overloadHelper execgen.OverloadHelper
}

// selOpBase contains all of the fields for non-constant binary selections.
type selOpBase struct {
	OneInputNode
	col1Idx        int
	col2Idx        int
	overloadHelper execgen.OverloadHelper
}

// {{define "selConstOp"}}
type _OP_CONST_NAME struct {
	selConstOpBase
	constArg _R_GO_TYPE
}

func (p *_OP_CONST_NAME) Next(ctx context.Context) coldata.Batch {
	// In order to inline the templated code of overloads, we need to have a
	// `_overloadHelper` local variable of type `execgen.OverloadHelper`.
	_overloadHelper := p.overloadHelper
	// However, the scratch is not used in all of the selection operators, so
	// we add this to go around "unused" error.
	_ = _overloadHelper
	var isNull bool
	for {
		batch := p.input.Next(ctx)
		if batch.Length() == 0 {
			return batch
		}

		vec := batch.ColVec(p.colIdx)
		col := vec._L_TYP()
		var idx int
		n := batch.Length()
		if vec.MaybeHasNulls() {
			nulls := vec.Nulls()
			_SEL_CONST_LOOP(true)
		} else {
			_SEL_CONST_LOOP(false)
		}
		if idx > 0 {
			batch.SetLength(idx)
			return batch
		}
	}
}

func (p *_OP_CONST_NAME) Init() {
	p.input.Init()
}

// {{end}}

// {{define "selOp"}}
type _OP_NAME struct {
	selOpBase
}

func (p *_OP_NAME) Next(ctx context.Context) coldata.Batch {
	// In order to inline the templated code of overloads, we need to have a
	// `_overloadHelper` local variable of type `execgen.OverloadHelper`.
	_overloadHelper := p.overloadHelper
	// However, the scratch is not used in all of the selection operators, so
	// we add this to go around "unused" error.
	_ = _overloadHelper
	var isNull bool
	for {
		batch := p.input.Next(ctx)
		if batch.Length() == 0 {
			return batch
		}

		vec1 := batch.ColVec(p.col1Idx)
		vec2 := batch.ColVec(p.col2Idx)
		col1 := vec1._L_TYP()
		col2 := vec2._R_TYP()
		n := batch.Length()

		var idx int
		if vec1.MaybeHasNulls() || vec2.MaybeHasNulls() {
			nulls := vec1.Nulls().Or(vec2.Nulls())
			_SEL_LOOP(true)
		} else {
			_SEL_LOOP(false)
		}
		if idx > 0 {
			batch.SetLength(idx)
			return batch
		}
	}
}

func (p *_OP_NAME) Init() {
	p.input.Init()
}

// {{end}}

// {{range .CmpOps}}
// {{range .LeftFamilies}}
// {{range .LeftWidths}}
// {{range .RightFamilies}}
// {{range .RightWidths}}

// {{template "selConstOp" .}}
// {{template "selOp" .}}

// {{end}}
// {{end}}
// {{end}}
// {{end}}
// {{end}}

// GetSelectionConstOperator returns the appropriate constant selection operator
// for the given left and right column types and comparison.
func GetSelectionConstOperator(
	cmpOp tree.ComparisonOperator,
	input colexecbase.Operator,
	inputTypes []*types.T,
	colIdx int,
	constArg tree.Datum,
	evalCtx *tree.EvalContext,
	cmpExpr *tree.ComparisonExpr,
) (colexecbase.Operator, error) {
	leftType, constType := inputTypes[colIdx], constArg.ResolvedType()
	c := colconv.GetDatumToPhysicalFn(constType)(constArg)
	selConstOpBase := selConstOpBase{
		OneInputNode: NewOneInputNode(input),
		colIdx:       colIdx,
	}
	if leftType.Family() != types.TupleFamily && constType.Family() != types.TupleFamily {
		// Tuple comparison has special null-handling semantics, so we will
		// fallback to the default comparison operator if either of the
		// input vectors is of a tuple type.
		switch cmpOp {
		// {{range .CmpOps}}
		case tree._NAME:
			switch typeconv.TypeFamilyToCanonicalTypeFamily(leftType.Family()) {
			// {{range .LeftFamilies}}
			case _LEFT_CANONICAL_TYPE_FAMILY:
				switch leftType.Width() {
				// {{range .LeftWidths}}
				case _LEFT_TYPE_WIDTH:
					switch typeconv.TypeFamilyToCanonicalTypeFamily(constType.Family()) {
					// {{range .RightFamilies}}
					case _RIGHT_CANONICAL_TYPE_FAMILY:
						switch constType.Width() {
						// {{range .RightWidths}}
						case _RIGHT_TYPE_WIDTH:
							return &_OP_CONST_NAME{selConstOpBase: selConstOpBase, constArg: c.(_R_GO_TYPE)}, nil
							// {{end}}
						}
						// {{end}}
					}
					// {{end}}
				}
				// {{end}}
			}
			// {{end}}
		}
	}
	return &defaultCmpConstSelOp{
		selConstOpBase:   selConstOpBase,
		adapter:          newComparisonExprAdapter(cmpExpr, evalCtx),
		constArg:         constArg,
		toDatumConverter: colconv.NewVecToDatumConverter(len(inputTypes), []int{colIdx}),
	}, nil
}

// GetSelectionOperator returns the appropriate two column selection operator
// for the given left and right column types and comparison.
func GetSelectionOperator(
	cmpOp tree.ComparisonOperator,
	input colexecbase.Operator,
	inputTypes []*types.T,
	col1Idx int,
	col2Idx int,
	evalCtx *tree.EvalContext,
	cmpExpr *tree.ComparisonExpr,
) (colexecbase.Operator, error) {
	leftType, rightType := inputTypes[col1Idx], inputTypes[col2Idx]
	selOpBase := selOpBase{
		OneInputNode: NewOneInputNode(input),
		col1Idx:      col1Idx,
		col2Idx:      col2Idx,
	}
	if leftType.Family() != types.TupleFamily && rightType.Family() != types.TupleFamily {
		// Tuple comparison has special null-handling semantics, so we will
		// fallback to the default comparison operator if either of the
		// input vectors is of a tuple type.
		switch cmpOp {
		// {{range .CmpOps}}
		case tree._NAME:
			switch typeconv.TypeFamilyToCanonicalTypeFamily(leftType.Family()) {
			// {{range .LeftFamilies}}
			case _LEFT_CANONICAL_TYPE_FAMILY:
				switch leftType.Width() {
				// {{range .LeftWidths}}
				case _LEFT_TYPE_WIDTH:
					switch typeconv.TypeFamilyToCanonicalTypeFamily(rightType.Family()) {
					// {{range .RightFamilies}}
					case _RIGHT_CANONICAL_TYPE_FAMILY:
						switch rightType.Width() {
						// {{range .RightWidths}}
						case _RIGHT_TYPE_WIDTH:
							return &_OP_NAME{selOpBase: selOpBase}, nil
							// {{end}}
						}
						// {{end}}
					}
					// {{end}}
				}
				// {{end}}
			}
			// {{end}}
		}
	}
	return &defaultCmpSelOp{
		selOpBase:        selOpBase,
		adapter:          newComparisonExprAdapter(cmpExpr, evalCtx),
		toDatumConverter: colconv.NewVecToDatumConverter(len(inputTypes), []int{col1Idx, col2Idx}),
	}, nil
}
