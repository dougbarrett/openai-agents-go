// Copyright 2025 The NLP Odyssey Authors
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

package agents

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/nlpodyssey/openai-agents-go/runcontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSingleHandoffSetup(t *testing.T) {
	agent1 := &Agent{Name: "test_1"}
	agent2 := &Agent{
		Name:          "test_2",
		AgentHandoffs: []*Agent{agent1},
	}

	handoffs, err := Runner().getHandoffs(agent1)
	require.NoError(t, err)
	assert.Len(t, handoffs, 0)

	handoffs, err = Runner().getHandoffs(agent2)
	require.NoError(t, err)
	require.Len(t, handoffs, 1)

	obj := handoffs[0]
	assert.Equal(t, DefaultHandoffToolName(agent1), obj.ToolName)
	assert.Equal(t, DefaultHandoffToolDescription(agent1), obj.ToolDescription)
	assert.Equal(t, "test_1", obj.AgentName)
}

func TestMultipleHandoffsSetup(t *testing.T) {
	agent1 := &Agent{Name: "test_1"}
	agent2 := &Agent{Name: "test_2"}
	agent3 := &Agent{
		Name:          "test_3",
		AgentHandoffs: []*Agent{agent1, agent2},
	}

	handoffs, err := Runner().getHandoffs(agent3)
	require.NoError(t, err)
	require.Len(t, handoffs, 2)

	assert.Equal(t, DefaultHandoffToolName(agent1), handoffs[0].ToolName)
	assert.Equal(t, DefaultHandoffToolName(agent2), handoffs[1].ToolName)

	assert.Equal(t, DefaultHandoffToolDescription(agent1), handoffs[0].ToolDescription)
	assert.Equal(t, DefaultHandoffToolDescription(agent2), handoffs[1].ToolDescription)

	assert.Equal(t, "test_1", handoffs[0].AgentName)
	assert.Equal(t, "test_2", handoffs[1].AgentName)
}

func TestCustomHandoffSetup(t *testing.T) {
	agent1 := &Agent{Name: "test_1"}
	agent2 := &Agent{Name: "test_2"}
	agent3 := &Agent{
		Name: "test_3",
		AgentHandoffs: []*Agent{
			agent1,
		},
		Handoffs: []Handoff{
			UnsafeHandoffFromAgent(HandoffFromAgentParams{
				Agent:                   agent2,
				ToolNameOverride:        "custom_tool_name",
				ToolDescriptionOverride: "custom tool description",
			}),
		},
	}

	handoffs, err := Runner().getHandoffs(agent3)
	require.NoError(t, err)
	require.Len(t, handoffs, 2)

	assert.Equal(t, "custom_tool_name", handoffs[0].ToolName)
	assert.Equal(t, DefaultHandoffToolName(agent1), handoffs[1].ToolName)

	assert.Equal(t, "custom tool description", handoffs[0].ToolDescription)
	assert.Equal(t, DefaultHandoffToolDescription(agent1), handoffs[1].ToolDescription)

	assert.Equal(t, "test_2", handoffs[0].AgentName)
	assert.Equal(t, "test_1", handoffs[1].AgentName)
}

type HandoffToolTestFoo struct {
	Bar string `json:"bar"`
}

type HandoffToolTestFooSchema struct{}

func (HandoffToolTestFooSchema) Name() string             { return "Foo" }
func (HandoffToolTestFooSchema) IsPlainText() bool        { return false }
func (HandoffToolTestFooSchema) IsStrictJSONSchema() bool { return true }
func (HandoffToolTestFooSchema) JSONSchema() map[string]any {
	return map[string]any{
		"title":                "Foo",
		"type":                 "object",
		"required":             []string{"bar"},
		"additionalProperties": false,
		"properties": map[string]any{
			"bar": map[string]any{
				"title": "Bar",
				"type":  "string",
			},
		},
	}
}
func (HandoffToolTestFooSchema) ValidateJSON(jsonStr string) (any, error) {
	r := strings.NewReader(jsonStr)
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	var v HandoffToolTestFoo
	err := dec.Decode(&v)
	return v, err
}

func TestHandoffInputType(t *testing.T) {
	onHandoff := func(context.Context, *runcontext.Wrapper, any) error {
		return nil
	}

	agent := &Agent{Name: "test"}
	obj, err := HandoffFromAgent(HandoffFromAgentParams{
		Agent:           agent,
		OnHandoff:       OnHandoffWithInput(onHandoff),
		InputJSONSchema: HandoffToolTestFooSchema{}.JSONSchema(),
	})
	require.NoError(t, err)

	cw := runcontext.NewWrapper(nil)

	// Invalid JSON should raise an error
	_, err = obj.OnInvokeHandoff(t.Context(), cw, "not json")
	require.Error(t, err)

	// Empty JSON should raise an error
	_, err = obj.OnInvokeHandoff(t.Context(), cw, "")
	require.Error(t, err)

	// Valid JSON should call the OnHandoff function
	invoked, err := obj.OnInvokeHandoff(t.Context(), cw, `{"bar": "baz"}`)
	require.NoError(t, err)
	assert.Same(t, agent, invoked)
}

func TestOnHandoffCalled(t *testing.T) {
	wasCalled := false

	onHandoff := func(context.Context, *runcontext.Wrapper, any) error {
		wasCalled = true
		return nil
	}

	agent := &Agent{Name: "test"}
	obj, err := HandoffFromAgent(HandoffFromAgentParams{
		Agent:           agent,
		OnHandoff:       OnHandoffWithInput(onHandoff),
		InputJSONSchema: HandoffToolTestFooSchema{}.JSONSchema(),
	})
	require.NoError(t, err)

	cw := runcontext.NewWrapper(nil)

	// Valid JSON should call the OnHandoff function
	invoked, err := obj.OnInvokeHandoff(t.Context(), cw, `{"bar": "baz"}`)
	require.NoError(t, err)
	assert.Same(t, agent, invoked)
	assert.True(t, wasCalled)
}

func TestOnHandoffError(t *testing.T) {
	handoffErr := errors.New("error")

	onHandoff := func(context.Context, *runcontext.Wrapper, any) error {
		return handoffErr
	}

	agent := &Agent{Name: "test"}
	obj, err := HandoffFromAgent(HandoffFromAgentParams{
		Agent:           agent,
		OnHandoff:       OnHandoffWithInput(onHandoff),
		InputJSONSchema: HandoffToolTestFooSchema{}.JSONSchema(),
	})
	require.NoError(t, err)

	cw := runcontext.NewWrapper(nil)

	// Valid JSON should call the OnHandoff function
	_, err = obj.OnInvokeHandoff(t.Context(), cw, `{"bar": "baz"}`)
	assert.ErrorIs(t, err, handoffErr)
}

func TestOnHandoffWithoutInputCalled(t *testing.T) {
	wasCalled := false

	onHandoff := func(context.Context, *runcontext.Wrapper) error {
		wasCalled = true
		return nil
	}

	agent := &Agent{Name: "test"}
	obj, err := HandoffFromAgent(HandoffFromAgentParams{
		Agent:     agent,
		OnHandoff: OnHandoffWithoutInput(onHandoff),
	})
	require.NoError(t, err)

	cw := runcontext.NewWrapper(nil)

	// Valid JSON should call the OnHandoff function
	invoked, err := obj.OnInvokeHandoff(t.Context(), cw, "")
	require.NoError(t, err)
	assert.Same(t, agent, invoked)
	assert.True(t, wasCalled)
}

func TestOnHandoffWithoutInputError(t *testing.T) {
	handoffErr := errors.New("error")

	onHandoff := func(context.Context, *runcontext.Wrapper) error {
		return handoffErr
	}

	agent := &Agent{Name: "test"}
	obj, err := HandoffFromAgent(HandoffFromAgentParams{
		Agent:     agent,
		OnHandoff: OnHandoffWithoutInput(onHandoff),
	})
	require.NoError(t, err)

	cw := runcontext.NewWrapper(nil)

	// Valid JSON should call the OnHandoff function
	_, err = obj.OnInvokeHandoff(t.Context(), cw, `{"bar": "baz"}`)
	assert.ErrorIs(t, err, handoffErr)
}
