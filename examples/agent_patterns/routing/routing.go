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

package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/nlpodyssey/openai-agents-go/agents"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/responses"
)

/*
This example shows the handoffs/routing pattern. The triage agent receives the first message, and
then hands off to the appropriate agent based on the language of the request. Responses are
streamed to the user.
*/

var (
	Model       = agents.NewAgentModelName("gpt-4o-mini")
	FrenchAgent = agents.New("french_agent").
			WithInstructions("You only speak French").
			WithModelOpt(param.NewOpt(Model))
	SpanishAgent = agents.New("spanish_agent").
			WithInstructions("You only speak Spanish").
			WithModelOpt(param.NewOpt(Model))
	EnglishAgent = agents.New("english_agent").
			WithInstructions("You only speak English").
			WithModelOpt(param.NewOpt(Model))
	TriageAgent = agents.New("triage_agent").
			WithInstructions("Handoff to the appropriate agent based on the language of the request.").
			WithAgentHandoffs(FrenchAgent, SpanishAgent, EnglishAgent).
			WithModelOpt(param.NewOpt(Model))
)

func main() {
	fmt.Print("Hi! We speak French, Spanish and English. How can I help? ")
	_ = os.Stdout.Sync()
	line, _, err := bufio.NewReader(os.Stdin).ReadLine()
	if err != nil {
		panic(err)
	}
	msg := string(line)

	agent := TriageAgent

	inputs := []agents.TResponseInputItem{{
		OfMessage: &responses.EasyInputMessageParam{
			Content: responses.EasyInputMessageContentUnionParam{
				OfString: param.NewOpt(msg),
			},
			Role: responses.EasyInputMessageRoleUser,
			Type: responses.EasyInputMessageTypeMessage,
		},
	}}

	for {
		result, err := agents.RunResponseInputsStreamed(
			context.Background(), agent, inputs,
		)
		if err != nil {
			panic(err)
		}
		err = result.StreamEvents(func(event agents.StreamEvent) error {
			e, ok := event.(agents.RawResponsesStreamEvent)
			if !ok {
				return nil
			}
			data := e.Data
			switch data.Type {
			case "response.output_text.delta":
				fmt.Print(data.Delta.OfString)
				_ = os.Stdout.Sync()
			case "response.content_part.done":
				fmt.Println()
			}
			return nil
		})
		if err != nil {
			panic(err)
		}

		inputs = result.ToInputList()
		fmt.Printf("\n\n")

		fmt.Print("Enter a message: ")
		_ = os.Stdout.Sync()
		line, _, err = bufio.NewReader(os.Stdin).ReadLine()
		if err != nil {
			panic(err)
		}
		userMsg := string(line)
		inputs = append(inputs, agents.TResponseInputItem{
			OfMessage: &responses.EasyInputMessageParam{
				Content: responses.EasyInputMessageContentUnionParam{
					OfString: param.NewOpt(userMsg),
				},
				Role: responses.EasyInputMessageRoleUser,
				Type: responses.EasyInputMessageTypeMessage,
			},
		})
		agent = result.CurrentAgent()
	}
}
