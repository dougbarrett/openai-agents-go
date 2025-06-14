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
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"

	"github.com/nlpodyssey/openai-agents-go/agents"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/responses"
)

//go:embed image_bison.jpg
var imageBison []byte

func main() {
	b64Image := base64.StdEncoding.EncodeToString(imageBison)

	agent := agents.New("Assistant").
		WithInstructions("You are a helpful assistant.").
		WithModel("gpt-4.1-nano")

	result, err := agents.RunResponseInputs(context.Background(), agent, []agents.TResponseInputItem{
		{OfMessage: &responses.EasyInputMessageParam{
			Content: responses.EasyInputMessageContentUnionParam{
				OfInputItemContentList: responses.ResponseInputMessageContentListParam{{
					OfInputImage: &responses.ResponseInputImageParam{
						Detail:   responses.ResponseInputImageDetailAuto,
						ImageURL: param.NewOpt("data:image/jpeg;base64," + b64Image),
					},
				}},
			},
			Role: responses.EasyInputMessageRoleUser,
			Type: responses.EasyInputMessageTypeMessage,
		}},
		{OfMessage: &responses.EasyInputMessageParam{
			Content: responses.EasyInputMessageContentUnionParam{
				OfString: param.NewOpt("What do you see in this image?"),
			},
			Role: responses.EasyInputMessageRoleUser,
			Type: responses.EasyInputMessageTypeMessage,
		}},
	})

	if err != nil {
		panic(err)
	}
	fmt.Println(result.FinalOutput)
}
