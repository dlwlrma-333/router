package tokenestimate

import (
	"encoding/json"
	"fmt"
	"strings"

	relaymodel "github.com/yeying-community/router/internal/relay/model"
	"github.com/yeying-community/router/internal/relay/relaymode"
)

// extractStructuredMeta only applies to Request-backed estimators such as
// OpenAI chat/responses and the current generic structured-request fallbacks.
// Anthropic raw messages estimation does not go through this path.
func extractStructuredMeta(req EstimateRequest) EstimateMeta {
	meta := EstimateMeta{}
	if req.Request == nil {
		return meta
	}
	r := req.Request
	meta.Messages = r.Messages
	meta.Input = r.Input
	meta.MaxTokens = resolveMaxTokens(r)

	switch req.RelayMode {
	case relaymode.ChatCompletions, relaymode.Messages:
		extractMessages(&meta, r.Messages)
		extractTools(&meta, r.Tools)
	case relaymode.Responses:
		if len(r.Messages) > 0 {
			extractMessages(&meta, r.Messages)
		} else {
			extractResponsesInput(&meta, r.Input)
		}
		appendExtraAnyText(&meta, r.Prompt)
		appendExtraAnyText(&meta, r.Metadata)
		appendExtraAnyText(&meta, r.ToolChoice)
		extractTools(&meta, r.Tools)
	default:
		extractMessages(&meta, r.Messages)
		appendExtraAnyText(&meta, r.Input)
		extractTools(&meta, r.Tools)
	}
	return meta
}

func resolveMaxTokens(r *relaymodel.GeneralOpenAIRequest) int {
	maxTokens := r.MaxTokens
	if r.MaxCompletionTokens != nil && *r.MaxCompletionTokens > maxTokens {
		maxTokens = *r.MaxCompletionTokens
	}
	if r.MaxOutputTokens != nil && *r.MaxOutputTokens > maxTokens {
		maxTokens = *r.MaxOutputTokens
	}
	return maxTokens
}

func extractMessages(meta *EstimateMeta, messages []relaymodel.Message) {
	for _, message := range messages {
		meta.MessagesCount++
		appendText(meta, message.Role)
		if message.Name != nil && strings.TrimSpace(*message.Name) != "" {
			meta.NamesCount++
			appendText(meta, *message.Name)
		}
		appendText(meta, message.StringContent())
		if message.ReasoningContent != nil {
			appendAnyText(meta, message.ReasoningContent)
		}
		for _, toolCall := range message.ToolCalls {
			appendText(meta, toolCall.Type)
			appendText(meta, toolCall.Function.Name)
			appendAnyText(meta, toolCall.Function.Arguments)
			appendAnyText(meta, toolCall.Function.Parameters)
		}
		if strings.TrimSpace(message.ToolCallId) != "" {
			appendText(meta, message.ToolCallId)
		}
	}
}

func extractTools(meta *EstimateMeta, tools []relaymodel.Tool) {
	for _, tool := range tools {
		meta.ToolsCount++
		appendToolText(meta, tool.Type)
		appendToolText(meta, tool.Function.Name)
		appendToolText(meta, tool.Function.Description)
		appendToolAnyText(meta, tool.Function.Parameters)
	}
}

func extractResponsesInput(meta *EstimateMeta, input any) {
	switch v := input.(type) {
	case string:
		appendText(meta, v)
	case []any:
		for _, item := range v {
			extractResponsesInput(meta, item)
		}
	case map[string]any:
		appendMapValueText(meta, v, "role")
		appendMapValueText(meta, v, "type")
		for _, key := range []string{"text", "instructions", "summary", "name", "id", "call_id"} {
			appendMapValueText(meta, v, key)
		}
		if content, ok := v["content"]; ok {
			extractResponsesInput(meta, content)
		}
		if inputText, ok := v["input_text"]; ok {
			extractResponsesInput(meta, inputText)
		}
		if outputText, ok := v["output_text"]; ok {
			extractResponsesInput(meta, outputText)
		}
		if args, ok := v["arguments"]; ok {
			appendAnyText(meta, args)
		}
		if params, ok := v["parameters"]; ok {
			appendAnyText(meta, params)
		}
	default:
		appendAnyText(meta, input)
	}
}

func appendMapValueText(meta *EstimateMeta, values map[string]any, key string) {
	value, ok := values[key]
	if !ok || value == nil {
		return
	}
	appendAnyText(meta, value)
}

func appendText(meta *EstimateMeta, value string) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return
	}
	meta.Texts = append(meta.Texts, trimmed)
}

func appendAnyText(meta *EstimateMeta, value any) {
	switch v := value.(type) {
	case nil:
		return
	case string:
		appendText(meta, v)
	case []string:
		for _, item := range v {
			appendText(meta, item)
		}
	case []any:
		for _, item := range v {
			appendAnyText(meta, item)
		}
	case map[string]any:
		b, err := json.Marshal(v)
		if err == nil {
			appendText(meta, string(b))
		}
	default:
		appendText(meta, fmt.Sprintf("%v", value))
	}
}

func appendToolText(meta *EstimateMeta, value string) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return
	}
	meta.ToolTexts = append(meta.ToolTexts, trimmed)
	meta.Texts = append(meta.Texts, trimmed)
}

func appendToolAnyText(meta *EstimateMeta, value any) {
	before := len(meta.Texts)
	appendAnyText(meta, value)
	if len(meta.Texts) > before {
		meta.ToolTexts = append(meta.ToolTexts, meta.Texts[before:]...)
	}
}

func appendExtraText(meta *EstimateMeta, value string) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return
	}
	meta.ExtraTexts = append(meta.ExtraTexts, trimmed)
	meta.Texts = append(meta.Texts, trimmed)
}

func appendExtraAnyText(meta *EstimateMeta, value any) {
	before := len(meta.Texts)
	appendAnyText(meta, value)
	if len(meta.Texts) > before {
		meta.ExtraTexts = append(meta.ExtraTexts, meta.Texts[before:]...)
	}
}
