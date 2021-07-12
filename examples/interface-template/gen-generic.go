// Code generated with https://github.com/kelindar/genny DO NOT EDIT.
// Any changes will be lost if this file is regenerated.

package main

import (
	"fmt"
)

// PrinterStringInterface parameter type printer interface
type PrinterStringInterface interface {
	Print(value string) string
}

// PrinterString parameter type printer
type PrinterString struct {
}

// Print generates string from an object
func (p *PrinterString) Print(value string) string {
	return fmt.Sprintf("%v", value)
}
