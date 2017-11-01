// Copyright © 2017 Dennis Walters
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

// Command is an interface that describes the sorts of commands that keylargo
// is able to test. Basically, we have to be able to set up the ARGS for the
// command, and we have to actually be able to execute it.
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
