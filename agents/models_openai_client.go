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
	"slices"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
)

type OpenaiClient struct {
	openai.Client
	BaseURL param.Opt[string]
}

func NewOpenaiClient(baseURL param.Opt[string], opts ...option.RequestOption) OpenaiClient {
	if baseURL.Valid() {
		opts = append(slices.Clone(opts), option.WithBaseURL(baseURL.Value))
	}
	return OpenaiClient{
		Client:  openai.NewClient(opts...),
		BaseURL: baseURL,
	}
}
