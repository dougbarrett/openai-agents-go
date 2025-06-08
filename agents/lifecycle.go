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

	"github.com/nlpodyssey/openai-agents-go/runcontext"
)

// RunHooks is implemented by an object that receives callbacks on various
// lifecycle events in an agent run.
type RunHooks interface {
	// OnAgentStart is called before the agent is invoked. Called each time the current agent changes.
	OnAgentStart(ctx context.Context, contextWrapper *runcontext.Wrapper, agent *Agent) error

	// OnAgentEnd is called when the agent produces a final output.
	OnAgentEnd(ctx context.Context, contextWrapper *runcontext.Wrapper, agent *Agent, output any) error

	// OnHandoff is called when a handoff occurs.
	OnHandoff(ctx context.Context, contextWrapper *runcontext.Wrapper, fromAgent, toAgent *Agent) error

	// OnToolStart is called before a tool is invoked.
	OnToolStart(ctx context.Context, contextWrapper *runcontext.Wrapper, agent *Agent, tool Tool) error

	// OnToolEnd is called after a tool is invoked.
	OnToolEnd(ctx context.Context, contextWrapper *runcontext.Wrapper, agent *Agent, tool Tool, result any) error
}

type NoOpRunHooks struct{}

func (NoOpRunHooks) OnAgentStart(context.Context, *runcontext.Wrapper, *Agent) error {
	return nil
}
func (NoOpRunHooks) OnAgentEnd(context.Context, *runcontext.Wrapper, *Agent, any) error {
	return nil
}
func (NoOpRunHooks) OnHandoff(context.Context, *runcontext.Wrapper, *Agent, *Agent) error {
	return nil
}
func (NoOpRunHooks) OnToolStart(context.Context, *runcontext.Wrapper, *Agent, Tool) error {
	return nil
}
func (NoOpRunHooks) OnToolEnd(context.Context, *runcontext.Wrapper, *Agent, Tool, any) error {
	return nil
}

// AgentHooks is implemented by an object that receives callbacks on various
// lifecycle events for a specific agent.
// You can set this on `agent.Hooks` to receive events for that specific agent.
type AgentHooks interface {
	// OnStart is called before the agent is invoked. Called each time the running agent is changed to this agent.
	OnStart(ctx context.Context, contextWrapper *runcontext.Wrapper, agent *Agent) error

	// OnEnd is called when the agent produces a final output.
	OnEnd(ctx context.Context, contextWrapper *runcontext.Wrapper, agent *Agent, output any) error

	// OnHandoff is called when the agent is being handed off to.
	// The `source` is the agent that is handing off to this agent.
	OnHandoff(ctx context.Context, contextWrapper *runcontext.Wrapper, agent, source *Agent) error

	// OnToolStart is called before a tool is invoked.
	OnToolStart(ctx context.Context, contextWrapper *runcontext.Wrapper, agent *Agent, tool Tool) error

	// OnToolEnd is called after a tool is invoked.
	OnToolEnd(ctx context.Context, contextWrapper *runcontext.Wrapper, agent *Agent, tool Tool, result any) error
}
