package errs

import (
	"errors"
	"fmt"
)

var (
	// 只支持一级指针
	ErrPointerOnly = errors.New("Pointer Only")
	ErrNoRows      = errors.New("No rows in result set")
)

func NewErrUnsupportedExpression(expr any) error {
	return fmt.Errorf("Unsupported expression: %v", expr)
}

func NewErrUnkonwnField(name string) error {
	return fmt.Errorf("Unkonwn field name: %v", name)
}

func NewErrUnkonwnColumn(name string) error {
	return fmt.Errorf("Unkonwn column name: %v", name)
}

func NewErrInvalidTagContent(pair string) error {
	return fmt.Errorf("Invalid tag content: %v", pair)
}
