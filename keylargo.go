// Package keylargo adds minimal Aruba-like functionality to a godog test
// suite.
package keylargo

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/DATA-DOG/godog"
)

type Command interface {
	SetArgs([]string)
	Execute() error
}

var rootCmd Command
var commandOutput string
var lastCommandRanErr error

// StepUp, given a godog suite, adds the keylargo step definitions and
// state cleanup to that suite. Effectively, this should be called in
// your godog suite setup.
func StepUp(s *godog.Suite) {
	s.Step(`^I run "([^"]*)"$`, iRun)
	s.Step(`the command succeeds`, theCommandSucceeds)
	s.Step(`the command fails`, theCommandFails)

	s.BeforeScenario(func(interface{}) {
		commandOutput = ""
		lastCommandRanErr = nil
	})
}

// SetRootCmd sets the root command for the application under test. This is
// used in the keylargo steps to run the app in question. The command passed
// in must implement keylargo's Command interface.
func SetRootCmd(cmd Command) {
	rootCmd = cmd
}

// LastCommandOutput returns the captured output of the last command run via
// the "I run" keylargo step.
func LastCommandOutput() string {
	return commandOutput
}

func iRun(fullCommand string) error {
	if rootCmd == nil {
		return fmt.Errorf("You can't get to Key Largo unless you set the root command.")
	}

	args := strings.Split(fullCommand, " ")[1:]

	rootCmd.SetArgs(args)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	lastCommandRanErr = rootCmd.Execute()
	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	w.Close()
	os.Stdout = old
	commandOutput = <-outC

	return nil
}

func theCommandSucceeds() error {
	if lastCommandRanErr != nil {
		return fmt.Errorf(
			"Expected a good exit status, got '%s'",
			lastCommandRanErr.Error(),
		)
	}

	return nil
}

func theCommandFails() error {
	if lastCommandRanErr == nil {
		return fmt.Errorf(
			"Expected a bad exit status, got nil",
		)
	}

	return nil
}