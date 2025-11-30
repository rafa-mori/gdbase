package main

import (
	"github.com/kubex-ecosystem/gdbase/internal/module"
	gl "github.com/kubex-ecosystem/logz"
)

// main initializes the logger and creates a new GDBase instance.
func main() {
	if err := module.RegX().Command().Execute(); err != nil {
		gl.Log("fatal", err.Error())
	}
}
