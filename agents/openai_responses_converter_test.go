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

package agents_test

import (
	"context"
	"errors"
	"testing"

	"github.com/nlpodyssey/openai-agents-go/agents"
	"github.com/nlpodyssey/openai-agents-go/computer"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/responses"
	"github.com/openai/openai-go/shared/constant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertToolChoiceStandardValues(t *testing.T) {
	// Make sure that the standard ToolChoice values map to themselves or
	// to "auto"/"required"/"none" as appropriate, and that special string
	// values map to the appropriate items.

	type R = responses.ResponseNewParamsToolChoiceUnion

	testCases := []struct {
		toolChoice string
		want       R
	}{
		{"", R{}},
		{"auto", R{OfToolChoiceMode: param.NewOpt(responses.ToolChoiceOptionsAuto)}},
		{"required", R{OfToolChoiceMode: param.NewOpt(responses.ToolChoiceOptionsRequired)}},
		{"none", R{OfToolChoiceMode: param.NewOpt(responses.ToolChoiceOptionsNone)}},
		{"file_search", R{OfHostedTool: &responses.ToolChoiceTypesParam{Type: responses.ToolChoiceTypesTypeFileSearch}}},
		{"web_search_preview", R{OfHostedTool: &responses.ToolChoiceTypesParam{Type: responses.ToolChoiceTypesTypeWebSearchPreview}}},
		{"web_search_preview_2025_03_11", R{OfHostedTool: &responses.ToolChoiceTypesParam{Type: responses.ToolChoiceTypesTypeWebSearchPreview2025_03_11}}},
		{"computer_use_preview", R{OfHostedTool: &responses.ToolChoiceTypesParam{Type: responses.ToolChoiceTypesTypeComputerUsePreview}}},
		{"image_generation", R{OfHostedTool: &responses.ToolChoiceTypesParam{Type: responses.ToolChoiceTypesTypeImageGeneration}}},
		{"code_interpreter", R{OfHostedTool: &responses.ToolChoiceTypesParam{Type: responses.ToolChoiceTypesTypeCodeInterpreter}}},
		{"mcp", R{OfHostedTool: &responses.ToolChoiceTypesParam{Type: responses.ToolChoiceTypesTypeMcp}}},
		{"my_function", R{ // Arbitrary string should be interpreted as a function name.
			OfFunctionTool: &responses.ToolChoiceFunctionParam{
				Name: "my_function",
				Type: constant.ValueOf[constant.Function](),
			},
		}},
	}

	for _, tc := range testCases {
		t.Run("toolChoice = "+tc.toolChoice, func(t *testing.T) {
			v := agents.ResponsesConverter().ConvertToolChoice(tc.toolChoice)
			assert.Equal(t, tc.want, v)
		})
	}
}

type PlainTextSchema struct{}

func (p PlainTextSchema) IsPlainText() bool                { return true }
func (p PlainTextSchema) Name() string                     { return "PlainText" }
func (p PlainTextSchema) JSONSchema() map[string]any       { return nil }
func (p PlainTextSchema) IsStrictJSONSchema() bool         { return false }
func (p PlainTextSchema) ValidateJSON(string) (any, error) { return nil, errors.New("not implemented") }

type FakeSchema struct{}

func (p FakeSchema) IsPlainText() bool { return false }
func (p FakeSchema) Name() string      { return "Fake" }
func (p FakeSchema) JSONSchema() map[string]any {
	return map[string]any{"title": "Fake"}
}
func (p FakeSchema) IsStrictJSONSchema() bool         { return true }
func (p FakeSchema) ValidateJSON(string) (any, error) { return nil, errors.New("not implemented") }

func TestGetResponseFormatPlainTextAndJsonSchema(t *testing.T) {
	// For plain text output, the converter should return a zero-value,
	// indicating no special response format constraint.
	// If an output schema is provided for a structured type, the converter
	// should return a ResponseTextConfigParam with the schema and strictness.

	// Default output (None) should be considered plain text.
	v := agents.ResponsesConverter().GetResponseFormat(nil)
	assert.Zero(t, v)

	// An explicit plain-text schema should also yield zero-value.
	v = agents.ResponsesConverter().GetResponseFormat(PlainTextSchema{})
	assert.Zero(t, v)

	// A model-based schema should produce a format object.
	v = agents.ResponsesConverter().GetResponseFormat(FakeSchema{})
	assert.Equal(t, responses.ResponseTextConfigParam{
		Format: responses.ResponseFormatTextConfigUnionParam{
			OfJSONSchema: &responses.ResponseFormatTextJSONSchemaConfigParam{
				Name:   "final_output",
				Schema: FakeSchema{}.JSONSchema(),
				Strict: param.NewOpt(true),
				Type:   constant.ValueOf[constant.JSONSchema](),
			},
		},
	}, v)
}

// DummyComputer tool implements a computer.Computer with minimal methods.
type DummyComputer struct{}

func (DummyComputer) Environment(context.Context) (computer.Environment, error) {
	return computer.EnvironmentLinux, nil
}
func (DummyComputer) Dimensions(context.Context) (computer.Dimensions, error) {
	return computer.Dimensions{Width: 800, Height: 600}, nil
}
func (DummyComputer) Screenshot(context.Context) (string, error) {
	return "", errors.New("not implemented")
}
func (DummyComputer) Click(context.Context, int64, int64, computer.Button) error {
	return errors.New("not implemented")
}
func (DummyComputer) DoubleClick(context.Context, int64, int64) error {
	return errors.New("not implemented")
}
func (DummyComputer) Scroll(context.Context, int64, int64, int64, int64) error {
	return errors.New("not implemented")
}
func (DummyComputer) Type(context.Context, string) error {
	return errors.New("not implemented")
}
func (DummyComputer) Wait(context.Context) error {
	return errors.New("not implemented")
}
func (DummyComputer) Move(context.Context, int64, int64) error {
	return errors.New("not implemented")
}
func (DummyComputer) Keypress(context.Context, []string) error {
	return errors.New("not implemented")
}
func (DummyComputer) Drag(context.Context, []computer.Position) error {
	return errors.New("not implemented")
}

func TestConvertToolsBasicTypesAndIncludes(t *testing.T) {
	// Construct a variety of tool types and make sure `ConvertTools` returns
	// a matching list of tool params and the expected includes.

	// Simple function tool
	toolFn := agents.FunctionTool{
		Name:             "fn",
		Description:      "...",
		ParamsJSONSchema: map[string]any{"title": "Fn"},
		OnInvokeTool: func(context.Context, string) (any, error) {
			return nil, errors.New("not implemented")
		},
	}

	// File search tool with IncludeSearchResults set
	fileTool := agents.FileSearchTool{
		MaxNumResults:        param.NewOpt[int64](3),
		VectorStoreIDs:       []string{"vs1"},
		IncludeSearchResults: true,
	}

	// Web search tool with custom params
	webTool := agents.WebSearchTool{SearchContextSize: responses.WebSearchToolSearchContextSizeHigh}

	// Wrap our concrete computer in a tools.ComputerTool for conversion.
	compTool := agents.ComputerTool{Computer: DummyComputer{}}
	allTools := []agents.Tool{toolFn, fileTool, webTool, compTool}
	converted, err := agents.ResponsesConverter().ConvertTools(t.Context(), allTools, nil)
	require.NoError(t, err)
	assert.Equal(t, &agents.ConvertedTools{
		Tools: []responses.ToolUnionParam{
			{
				OfFunction: &responses.FunctionToolParam{
					Name:        "fn",
					Parameters:  toolFn.ParamsJSONSchema,
					Strict:      param.NewOpt(true),
					Description: param.NewOpt("..."),
					Type:        constant.ValueOf[constant.Function](),
				},
			},
			{
				OfFileSearch: &responses.FileSearchToolParam{
					VectorStoreIDs: []string{"vs1"},
					MaxNumResults:  param.NewOpt[int64](3),
					Type:           constant.ValueOf[constant.FileSearch](),
				},
			},
			{
				OfWebSearchPreview: &responses.WebSearchToolParam{
					Type:              responses.WebSearchToolTypeWebSearchPreview,
					UserLocation:      responses.WebSearchToolUserLocationParam{},
					SearchContextSize: responses.WebSearchToolSearchContextSizeHigh,
				},
			},
			{
				OfComputerUsePreview: &responses.ComputerToolParam{
					DisplayHeight: 600,
					DisplayWidth:  800,
					Environment:   responses.ComputerToolEnvironmentLinux,
					Type:          constant.ValueOf[constant.ComputerUsePreview](),
				},
			},
		},
		// The Includes list should have exactly the include for file search
		// when IncludeSearchResults is true.
		Includes: []responses.ResponseIncludable{
			responses.ResponseIncludableFileSearchCallResults,
		},
	}, converted)

	t.Run("only one computer tool should be allowed", func(t *testing.T) {
		_, err = agents.ResponsesConverter().ConvertTools(t.Context(), []agents.Tool{compTool, compTool}, nil)
		var target agents.UserError
		assert.ErrorAs(t, err, &target)
	})
}

func TestConvertToolsIncludesHandoffs(t *testing.T) {
	//  When handoff objects are included, `ConvertTools` should append their
	//  tool param items after tools and include appropriate descriptions.

	agent := &agents.Agent{
		Name:               "support",
		HandoffDescription: "Handles support",
	}
	handoff, err := agents.SafeHandoffFromAgent(agents.HandoffFromAgentParams{Agent: agent})
	require.NoError(t, err)
	require.NotNil(t, handoff)

	converted, err := agents.ResponsesConverter().ConvertTools(t.Context(), nil, []agents.Handoff{*handoff})
	require.NoError(t, err)
	assert.Equal(t, &agents.ConvertedTools{
		Tools: []responses.ToolUnionParam{
			{
				OfFunction: &responses.FunctionToolParam{
					Name: agents.DefaultHandoffToolName(agent),
					Parameters: map[string]any{
						"type":                 "object",
						"additionalProperties": false,
						"properties":           map[string]any{},
						"required":             []string{},
					},
					Strict:      param.NewOpt(true),
					Description: param.NewOpt(agents.DefaultHandoffToolDescription(agent)),
					Type:        constant.ValueOf[constant.Function](),
				},
			},
		},
		Includes: nil,
	}, converted)
}
