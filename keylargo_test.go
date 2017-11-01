package keylargo

import (
	"testing"
)

type sweet struct {
	Steps                  map[string]interface{}
	BeforeScenarioHandlers []func(interface{})
}

func (s *sweet) Step(expr interface{}, stepFunc interface{}) {
	var key string

	switch t := expr.(type) {
	case string:
		key = t
	}

	s.Steps[key] = stepFunc
}

func (s *sweet) BeforeScenario(fn func(interface{})) {
	s.BeforeScenarioHandlers = append(s.BeforeScenarioHandlers, fn)
}

func TestStepUp(t *testing.T) {
	s := &sweet{
		Steps: make(map[string]interface{}),
	}

	StepUp(s)

	if s.Steps[`^I run "([^"]*)"$`] == nil {
		t.Fatalf("The 'I run' step was not added")
	}

	if s.Steps[`^the command succeeds$`] == nil {
		t.Fatalf("The 'the command succeeded' step was not added")
	}

	if s.Steps[`^the command fails$`] == nil {
		t.Fatalf("The 'the command fails' step was not added")
	}

	if len(s.BeforeScenarioHandlers) == 0 {
		t.Fatalf("The cleanup handler was not added")
	}
}
