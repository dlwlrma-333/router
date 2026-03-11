package channel

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	openaiadaptor "github.com/yeying-community/router/internal/relay/adaptor/openai"
)

type responsesEnvelope struct {
	Output []struct {
		Content []struct {
			Type       string `json:"type"`
			Text       string `json:"text"`
			OutputText string `json:"output_text"`
			Result     string `json:"result"`
		} `json:"content"`
	} `json:"output"`
}

func parseChatModelTestResponse(resp string) (*openaiadaptor.TextResponse, string, error) {
	var response openaiadaptor.TextResponse
	err := json.Unmarshal([]byte(resp), &response)
	if err != nil {
		return nil, "", err
	}
	if len(response.Choices) == 0 {
		return nil, "", errors.New("response has no choices")
	}
	stringContent, ok := response.Choices[0].Content.(string)
	if !ok {
		return nil, "", errors.New("response content is not string")
	}
	return &response, stringContent, nil
}

func parseResponsesModelTestResponse(resp string) (string, error) {
	var env responsesEnvelope
	if err := json.Unmarshal([]byte(resp), &env); err != nil {
		return "", err
	}
	contentTypes := make([]string, 0)
	for _, output := range env.Output {
		for _, content := range output.Content {
			if content.Type != "" {
				contentTypes = append(contentTypes, content.Type)
			} else {
				contentTypes = append(contentTypes, "<empty>")
			}
			if content.Text != "" {
				return content.Text, nil
			}
			if content.OutputText != "" {
				return content.OutputText, nil
			}
		}
	}
	return "", errors.New("response has no output text, content types: " + strings.Join(contentTypes, ","))
}

func parseResponsesImageTestResponse(resp string) (string, error) {
	var env responsesEnvelope
	if err := json.Unmarshal([]byte(resp), &env); err != nil {
		return "", err
	}
	contentTypes := make([]string, 0)
	imageCount := 0
	for _, output := range env.Output {
		for _, content := range output.Content {
			contentType := strings.TrimSpace(content.Type)
			if contentType == "" {
				contentType = "<empty>"
			}
			contentTypes = append(contentTypes, contentType)
			if content.Text != "" || content.OutputText != "" {
				return "responses 接口返回成功", nil
			}
			switch contentType {
			case "image_generation_call", "output_image", "image":
				imageCount++
			}
		}
	}
	if imageCount > 0 {
		return fmt.Sprintf("responses 接口返回 %d 个图片结果", imageCount), nil
	}
	return "", errors.New("response has no image output, content types: " + strings.Join(contentTypes, ","))
}

func parseTextModelTestResponse(resp string) (string, error) {
	_, chatText, chatErr := parseChatModelTestResponse(resp)
	if chatErr == nil {
		return chatText, nil
	}
	responsesText, responsesErr := parseResponsesModelTestResponse(resp)
	if responsesErr == nil {
		return responsesText, nil
	}
	return "", fmt.Errorf("parse as chat failed: %v; parse as responses failed: %v", chatErr, responsesErr)
}
