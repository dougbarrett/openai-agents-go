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
	"strings"

	"github.com/nlpodyssey/openai-agents-go/modelsettings"
	"github.com/nlpodyssey/openai-agents-go/types/optional"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/param"
)

type chatCmplHelpers struct {
}

func ChatCmplHelpers() chatCmplHelpers { return chatCmplHelpers{} }

func (chatCmplHelpers) IsOpenAI(client OpenaiClient) bool {
	return strings.HasPrefix(client.BaseURL.Or(""), "https://api.openai.com")
}

func (h chatCmplHelpers) GetStoreParam(
	client OpenaiClient,
	modelSettings modelsettings.ModelSettings,
) optional.Optional[bool] {
	// Match the behavior of Responses where store is True when not given
	if modelSettings.Store.Present {
		return modelSettings.Store
	}

	if h.IsOpenAI(client) {
		return optional.Value(true)
	}
	return optional.None[bool]()
}

func (h chatCmplHelpers) GetStreamOptionsParam(
	client OpenaiClient,
	modelSettings modelsettings.ModelSettings,
	stream bool,
) openai.ChatCompletionStreamOptionsParam {
	var options openai.ChatCompletionStreamOptionsParam
	if !stream {
		return options
	}

	if modelSettings.IncludeUsage.Present {
		options.IncludeUsage = param.NewOpt(modelSettings.IncludeUsage.Value)
	} else if h.IsOpenAI(client) {
		options.IncludeUsage = param.NewOpt(true)
	}
	return options
}
