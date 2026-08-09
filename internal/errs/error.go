package errs

import (
	"errors"
	"fmt"
)

var (
	// 只支持一级指针
	ErrPointerOnly = errors.New("Pointer Only")
)

func NewErrUnsupportedExpression(expr any) error {
	return fmt.Errorf("Unsupported expression: %v", expr)
}

func NewErrUnkonwnField(name string) error {
	return fmt.Errorf("Unkonwn field name: %v", name)
}
