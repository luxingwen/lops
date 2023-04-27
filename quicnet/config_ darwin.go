//go:build darwin
// +build darwin

package quicnet

const (
	defaultInterpreter    = "sh"
	defaultInterpreterArg = "-c"
	defaultScriptSuffix   = ".sh"
)
