package helper

import (
	"github.com/jaam8/web_calculator/common-lib/errors"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestValidateExpression(t *testing.T) {
	tests := []struct {
		name string
		expr string
		want error
	}{
		{
			name: "empty expression",
			expr: "",
			want: errors.ErrInvalidExpression,
		},
		{
			name: "only space",
			expr: " ",
			want: errors.ErrInvalidExpression,
		},
		{
			name: "invalid characters",
			expr: "2+2=4",
			want: errors.ErrInvalidExpression,
		},
		{
			name: "unbalanced brackets",
			expr: "(2+2",
			want: errors.ErrInvalidExpression,
		},
		{
			name: "operator at end",
			expr: "2+2+",
			want: errors.ErrInvalidExpression,
		},
		{
			name: "operator at start",
			expr: "+2+2",
			want: errors.ErrInvalidExpression,
		},
		{
			name: "two operators in operation",
			expr: "2++2",
			want: errors.ErrInvalidExpression,
		},
		{
			name: "division by zero",
			expr: "2/0",
			want: errors.ErrDivideByZero,
		},
		{
			name: "with excess dot",
			expr: "2.0.1+3",
			want: errors.ErrInvalidExpression,
		},
		{
			name: "operation with wrong unary operator",
			expr: "2+-2",
			want: errors.ErrInvalidExpression,
		},
		{
			name: "valid operation with extra brackets",
			expr: "((2+3)*4)",
			want: nil,
		},
		{
			name: "valid divide expression",
			expr: "2/0.1",
			want: nil,
		},
		{
			name: "valid brackets expression",
			expr: "2+(2*2)",
			want: nil,
		},
		{
			name: "valid expression",
			expr: "2+2",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateExpression(tt.expr)
			if tt.want != nil {
				require.ErrorIs(t, err, tt.want)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
