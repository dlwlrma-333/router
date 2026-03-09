package openai

import (
	"fmt"
	"strings"

	relaychannel "github.com/yeying-community/router/internal/relay/channel"
	"github.com/yeying-community/router/internal/relay/model"
)

func ResponseText2Usage(responseText string, modelName string, promptTokens int) *model.Usage {
	usage := &model.Usage{}
	usage.PromptTokens = promptTokens
	usage.CompletionTokens = CountTokenText(responseText, modelName)
	usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	return usage
}

func shouldTrimOpenAIV1Path(baseURL string) bool {
	normalized := strings.ToLower(strings.TrimRight(strings.TrimSpace(baseURL), "/"))
	return strings.HasSuffix(normalized, "/v1") ||
		strings.HasSuffix(normalized, "/openai") ||
		strings.HasSuffix(normalized, "/v1beta/openai")
}

func GetFullRequestURL(baseURL string, requestURL string, channelProtocol int) string {
	if channelProtocol == relaychannel.OpenAI && shouldTrimOpenAIV1Path(baseURL) {
		return fmt.Sprintf("%s%s", strings.TrimSuffix(baseURL, "/"), strings.TrimPrefix(requestURL, "/v1"))
	}
	fullRequestURL := fmt.Sprintf("%s%s", baseURL, requestURL)

	if strings.HasPrefix(baseURL, "https://gateway.ai.cloudflare.com") {
		switch channelProtocol {
		case relaychannel.OpenAI:
			fullRequestURL = fmt.Sprintf("%s%s", baseURL, strings.TrimPrefix(requestURL, "/v1"))
		case relaychannel.Azure:
			fullRequestURL = fmt.Sprintf("%s%s", baseURL, strings.TrimPrefix(requestURL, "/openai/deployments"))
		}
	}
	return fullRequestURL
}
