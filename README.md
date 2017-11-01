Keylargo is a set [godog](https://github.com/DATA-DOG/godog) helpers for in-process CLI application testing. In essence, it is a minimal [Aruba](https://github.com/cucumber/aruba)-like experience.

## Why? ##

Because I work on a lot of CLI apps, and I drank the Cucumber/Aruba special punch to such a degree that it's my preferred way to work.

Also, since it is an in-process tester, it's fairly trivial to mock external REST APIs and the like (for better or for worse).

## Installation ##

I suggest the use of [dep](https://github.com/golang/dep) for managing your Go deps, but you should be able to install it directly without issues:

```
go get github.com/ess/keylargo
```

As mentioned, though, it's probably better to use a package manager that supports SemVer.

## Usage ###

Let's talk about the `echo` command for a second. By default, it just prints its arguments to STDOUT, followed by a newline. If you provide the `-n` flag, it still does that, but it does not print the newline.

### The Command ###

In order for keylargo to run a command, it must implement the `keylargo.Command` interface:

```go
type Command interface {
	SetArgs([]string)
	Execute() error
}
```

You can use any CLI framework you like (or just roll your own implementation), but a great example of an object that already implements this interface is `*cobra.Command` from [cobra](https://github.com/spf13/cobra).

Also, in order for keylargo to actually run your command, you'll need to set it as the root command via `keylargo.SetRootCmd(yourCommand)`, most likely in your godog suite setup (see `main_test.go` below).

### features/echo.feature ###

```gherkin
Feature: echo
  In order to print stuff to the terminal
  As a user
  I want to be able to echo some text
  And I want for this stanza to be less awkward, but that's what examples are like.

  Scenario: default behavior
    When I run "echo ohai tharrrrrrr"
    Then the output contains "ohai tharrrrrrr"
    And the output ends with a newline
    And the command succeeds

  Scenario: suppress newline
    When I run "echo -n sup duder"
    Then the output contains "sup duder"
    But the output has no newlines
    And the command succeeds
```

### main_test.go ###

```go
package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/godog"
	"github.com/ess/keylargo"
)

func theOutputContains(expected string) error {
	if !strings.Contains(keylargo.LastCommandOutput(), expected) {
		return fmt.Errorf("Expected output to contain '%s'", expected)
	}

	return nil
}

func terminatingNewline() error {
	if !strings.HasSuffix(keylargo.LastCommandOutput(), "\n") {
		return fmt.Errorf("Expected output to have a newline terminator")
	}

	return nil
}

func noNewline() error {
	if strings.Contains(keylargo.LastCommandOutput(), "\n") {
		return fmt.Errorf("Expected output to contain no newlines")
	}

	return nil
}

func TestMain(m *testing.M) {
	// Set up the command that is actually run via keylargo. The object passed in
	// must implement the following interface:
	//
	//   type Command interface{
	//     SetArgs([]string)
	//     Execute() error
	//   }

	keylargo.SetRootCmd(echoCmd)

	status := godog.RunWithOptions("godog",
		func(s *godog.Suite) {
			// Add the keylargo steps to the suite
			keylargo.StepUp(s)

			// Add your own custom steps

			s.Step(`^the output contains "([^"]*)"$`, theOutputContains)
			s.Step(`the output ends with a newline`, terminatingNewline)
			s.Step(`the output has no newlines`, noNewline)

			s.BeforeScenario(func(interface{}) {
				skipNewline = false
			})

		},

		godog.Options{
			Format:    "pretty",
			Paths:     []string{"features"},
			Randomize: time.Now().UTC().UnixNano(),
		},
	)

	if st := m.Run(); st > status {
		status = st
	}

	os.Exit(status)
}
```

### main.go ###

```go
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var skipNewline bool

var echoCmd = &cobra.Command{
	Use:   "echo",
	Short: "what did i just say?",
	RunE: func(cmd *cobra.Command, args []string) error {
		output := strings.Join(args, " ")

		fmt.Printf("%s"+terminator(), output)

		return nil
	},
}

func init() {
	echoCmd.Flags().BoolVarP(&skipNewline, "no-newline", "n", false,
		"Suppress the terminating newline")
}

func main() {
	err := echoCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func terminator() string {
	if skipNewline {
		return ""
	}

	return "\n"
}
```

## History ##
