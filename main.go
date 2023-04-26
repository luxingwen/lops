package main

import (
	"fmt"
	"time"

	"lops/quicnet"
)

func main() {
	cmdRunner := quicnet.NewCmdRunner()

	scriptContent := `
echo "Hello, $1!"
echo "This is a test script."
`

	interpreter := "sh"
	params := []string{"-c", scriptContent, "World"}
	timeout := 5 * time.Second
	stdin := ""

	output, err := cmdRunner.RunScript(scriptContent, interpreter, params, timeout, stdin)

	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Output:", output)
	}
}
