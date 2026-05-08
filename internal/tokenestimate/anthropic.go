package tokenestimate

import (
	"encoding/json"
	"fmt"
	"strings"
)

type anthropicEstimateRequest struct {
	System     any                        `json:"system,omitempty"`
	Messages   []anthropicEstimateItem    `json:"messages,omitempty"`
	Tools      []anthropicEstimateTool    `json:"tools,omitempty"`
	MaxTokens  int                        `json:"max_tokens,omitempty"`
	Metadata   map[string]json.RawMessage `json:"metadata,omitempty"`
	ToolChoice any                        `json:"tool_choice,omitempty"`
}

type anthropicEstimateItem struct {
	Role    string `json:"role,omitempty"`
	Content any    `json:"content,omitempty"`
}

type anthropicEstimateTool struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	InputSchema any    `json:"input_schema,omitempty"`
}

func estimateAnthropicFromRaw(req EstimateRequest, model string) (EstimateResult, error) {
	meta, err := extractAnthropicMeta(req.RawBody)
	if err != nil {
		return EstimateResult{}, err
	}
	return EstimateResult{
		PromptTokens: estimateTextsHeuristic(meta.Texts, familyAnthropic),
		Source:       "local_anthropic_heuristic",
		Precision:    PrecisionHeuristic,
		Estimator:    "anthropic_heuristic",
	}, nil
}

func extractAnthropicMeta(raw []byte) (EstimateMeta, error) {
	meta := EstimateMeta{}
	if len(raw) == 0 {
		return meta, fmt.Errorf("anthropic raw request body is empty")
	}
	var req anthropicEstimateRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return meta, fmt.Errorf("unmarshal anthropic request: %w", err)
	}
	meta.MaxTokens = req.MaxTokens
	appendAnyText(&meta, req.System)
	for _, message := range req.Messages {
		meta.MessagesCount++
		appendText(&meta, message.Role)
		extractAnthropicContent(&meta, message.Content)
	}
	for _, tool := range req.Tools {
		meta.ToolsCount++
		appendToolText(&meta, tool.Name)
		appendToolText(&meta, tool.Description)
		appendToolAnyText(&meta, tool.InputSchema)
	}
	appendExtraAnyText(&meta, req.ToolChoice)
	appendExtraAnyText(&meta, req.Metadata)
	return meta, nil
}

func extractAnthropicContent(meta *EstimateMeta, content any) {
	switch v := content.(type) {
	case string:
		appendText(meta, v)
	case []any:
		for _, item := range v {
			block, ok := item.(map[string]any)
			if !ok {
				appendAnyText(meta, item)
				continue
			}
			extractAnthropicContentBlock(meta, block)
		}
	case map[string]any:
		extractAnthropicContentBlock(meta, v)
	default:
		appendAnyText(meta, content)
	}
}

func extractAnthropicContentBlock(meta *EstimateMeta, block map[string]any) {
	blockType := strings.ToLower(strings.TrimSpace(fmt.Sprint(block["type"])))
	appendText(meta, blockType)
	switch blockType {
	case "text", "input_text", "output_text":
		appendMapValueText(meta, block, "text")
	case "thinking", "redacted_thinking":
		appendMapValueText(meta, block, "thinking")
		appendMapValueText(meta, block, "text")
	case "tool_use":
		appendMapValueText(meta, block, "id")
		appendMapValueText(meta, block, "name")
		if input, ok := block["input"]; ok {
			appendAnyText(meta, input)
		}
	case "tool_result":
		appendMapValueText(meta, block, "tool_use_id")
		if resultContent, ok := block["content"]; ok {
			extractAnthropicContent(meta, resultContent)
		}
	default:
		appendMapValueText(meta, block, "text")
		if nested, ok := block["content"]; ok {
			extractAnthropicContent(meta, nested)
		}
		if input, ok := block["input"]; ok {
			appendAnyText(meta, input)
		}
	}
}
