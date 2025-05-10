package helper

import (
	"github.com/jaam8/web_calculator/common-lib/errors"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestToRPN(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		want    []string
		wantErr error
	}{
		{
			name:    "basic math with precedence",
			expr:    "3+4*2/(1-5)",
			want:    []string{"3", "4", "2", "*", "1", "5", "-", "/", "+"},
			wantErr: nil,
		},
		{
			name:    "multiple parentheses and operations",
			expr:    "5+((3*4)/2)-1",
			want:    []string{"5", "3", "4", "*", "2", "/", "+", "1", "-"},
			wantErr: nil,
		},
		{
			name:    "long addition chain",
			expr:    "1+2+3+4+5",
			want:    []string{"1", "2", "+", "3", "+", "4", "+", "5", "+"},
			wantErr: nil,
		},
		{
			name:    "nested parentheses",
			expr:    "((2+3)*(4+5))",
			want:    []string{"2", "3", "+", "4", "5", "+", "*"},
			wantErr: nil,
		},
		{
			name:    "decimal numbers",
			expr:    "0.5+1.25*2",
			want:    []string{"0.5", "1.25", "2", "*", "+"},
			wantErr: nil,
		},
		{
			name:    "unary operator not supported — passed through validation",
			expr:    "4+(3-2)*2",
			want:    []string{"4", "3", "2", "-", "2", "*", "+"},
			wantErr: nil,
		},
		{
			name:    "complex without spaces",
			expr:    "6/(2*(1+2))",
			want:    []string{"6", "2", "1", "2", "+", "*", "/"},
			wantErr: nil,
		},
		{
			name:    "expression with whitespaces",
			expr:    " 7 + 8 * ( 9 - 3 ) ",
			want:    []string{"7", "8", "9", "3", "-", "*", "+"},
			wantErr: nil,
		},
		{
			name:    "division before addition",
			expr:    "10/2+3",
			want:    []string{"10", "2", "/", "3", "+"},
			wantErr: nil,
		},
		{
			name:    "multiple nested operations",
			expr:    "((1+2)*3)-(4/(2+2))",
			want:    []string{"1", "2", "+", "3", "*", "4", "2", "2", "+", "/", "-"},
			wantErr: nil,
		},
		{
			name:    "too many operators in one operation",
			expr:    "2+-2",
			want:    nil,
			wantErr: errors.ErrInvalidExpression,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToRPN(tt.expr)
			require.Equal(t, tt.want, got)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
