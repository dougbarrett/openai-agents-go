package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/nlpodyssey/openai-agents-go/agents"
	"github.com/nlpodyssey/openai-agents-go/tools"
	"github.com/openai/openai-go/responses"
	"github.com/openai/openai-go/shared/constant"
)

func main() {
	agent := agents.New("Image generator").
		WithInstructions("You are a helpful agent.").
		WithTools(tools.ImageGeneration{
			ToolConfig: responses.ToolImageGenerationParam{
				Quality: "low",
				Type:    constant.ValueOf[constant.ImageGeneration](),
			},
		}).
		WithModel("gpt-4o")

	fmt.Println("Generating image, this may take a while...")

	input := "Create an image of a frog eating a pizza, comic book style."
	result, err := agents.Run(context.Background(), agent, input)
	if err != nil {
		panic(err)
	}

	fmt.Println(result.FinalOutput)

	for _, item := range result.NewItems {
		toolCallItem, ok := item.(agents.ToolCallItem)
		if !ok {
			continue
		}
		imageGenerationCall, ok := toolCallItem.RawItem.(agents.ResponseOutputItemImageGenerationCall)
		if !ok {
			continue
		}
		imgResult := imageGenerationCall.Result

		fileName, err := createTempImageFile(imgResult)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Temporary file created: %s\n", fileName)

		if err = openFile(fileName); err != nil {
			panic(err)
		}
	}
}

func openFile(path string) error {
	switch runtime.GOOS {
	case "linux":
		return exec.Command("xdg-open", path).Start()
	case "darwin":
		return exec.Command("open", path).Start()
	case "windows":
		return exec.Command("cmd", "/c", "start", "", path).Start()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func createTempImageFile(b64Content string) (fileName string, err error) {
	b, err := base64.StdEncoding.DecodeString(b64Content)
	if err != nil {
		return "", fmt.Errorf("could not decode base64 image content: %w", err)
	}

	f, err := os.CreateTemp("", "*.png")
	if err != nil {
		return "", err
	}
	defer func() {
		if e := f.Close(); e != nil {
			err = errors.Join(err, e)
		}
	}()

	if _, err = f.Write(b); err != nil {
		panic(err)
	}

	return f.Name(), nil
}
