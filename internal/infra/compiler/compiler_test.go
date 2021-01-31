package compiler

import "testing"

func TestMake(t *testing.T) {
	Make(MakeRequest{
		PKey:        "4324242",
		FingerPrint: "42424",
		Port:        12,
	})
}
