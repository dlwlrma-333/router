package tokenestimate

import (
	"fmt"
	"strings"

	openaiadaptor "github.com/yeying-community/router/internal/relay/adaptor/openai"
	"github.com/yeying-community/router/internal/relay/relaymode"
)

// OpenAI request counting applies tokenizer-based accounting to structured
// request data that has already been extracted from OpenAI-compatible requests.
func estimateOpenAIExact(meta EstimateMeta, model string, relayMode int) int {
	total := 0
	switch relayMode {
	case relaymode.ChatCompletions, relaymode.Messages:
		total += openaiadaptor.CountTokenMessages(meta.Messages, model)
	case relaymode.Responses:
		if len(meta.Messages) > 0 {
			total += openaiadaptor.CountTokenMessages(meta.Messages, model)
		} else {
			total += countResponsesInput(meta.Input, model)
		}
	default:
		total += openaiadaptor.CountTokenInput(meta.Input, model)
	}
	total += countTexts(meta.ToolTexts, model)
	total += countTexts(meta.ExtraTexts, model)
	if meta.ToolsCount > 0 {
		total += meta.ToolsCount * 8
	}
	return total
}

func countResponsesInput(input any, model string) int {
	switch v := input.(type) {
	case string:
		return openaiadaptor.CountTokenText(v, model)
	case []any:
		total := 0
		for _, item := range v {
			switch typed := item.(type) {
			case string:
				total += openaiadaptor.CountTokenText(typed, model)
			case map[string]any:
				total += countMapValuesText(typed, model)
			default:
				total += openaiadaptor.CountTokenText(strings.TrimSpace(relaymodelString(item)), model)
			}
		}
		return total
	default:
		return openaiadaptor.CountTokenInput(input, model)
	}
}

func countMapValuesText(values map[string]any, model string) int {
	total := 0
	for _, key := range []string{"role", "text", "content", "instructions", "summary", "name"} {
		if value, ok := values[key]; ok {
			total += openaiadaptor.CountTokenText(strings.TrimSpace(relaymodelString(value)), model)
		}
	}
	return total
}

func countTexts(texts []string, model string) int {
	total := 0
	for _, text := range texts {
		total += openaiadaptor.CountTokenText(text, model)
	}
	return total
}

func relaymodelString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	default:
		return fmt.Sprintf("%v", value)
	}
}
