//go:build windows
// +build windows

package quicnet

const (
	defaultInterpreter    = "powershell"
	defaultInterpreterArg = "-Command"
	defaultScriptSuffix   = ".ps1"
)
