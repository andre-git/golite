package lib

import (
	"fmt"
	"golite/internal/vdbe"
	"strings"
)

func Abs(args []vdbe.Value) (vdbe.Value, error) {
	if len(args) != 1 { return nil, fmt.Errorf("abs() requires 1 argument") }
	v := args[0]
	switch v.Type() {
	case 1: 
		val := v.Int64()
		if val < 0 {
			if val == -9223372036854775808 {
				return nil, fmt.Errorf("integer overflow")
			}
			val = -val
		}
		return &intValue{val}, nil
	case 5: 
		return nil, nil
	default:
		return v, nil
	}
}

func Lower(args []vdbe.Value) (vdbe.Value, error) {
	if len(args) != 1 { return nil, fmt.Errorf("lower() requires 1 argument") }
	return &stringValue{strings.ToLower(args[0].String())}, nil
}

func Upper(args []vdbe.Value) (vdbe.Value, error) {
	if len(args) != 1 { return nil, fmt.Errorf("upper() requires 1 argument") }
	return &stringValue{strings.ToUpper(args[0].String())}, nil
}

type intValue struct{ val int64 }
func (v *intValue) Int64() int64 { return v.val }
func (v *intValue) String() string { return fmt.Sprintf("%d", v.val) }
func (v *intValue) Type() byte { return 1 }

type stringValue struct{ val string }
func (v *stringValue) Int64() int64 { return 0 }
func (v *stringValue) String() string { return v.val }
func (v *stringValue) Type() byte { return 3 }
