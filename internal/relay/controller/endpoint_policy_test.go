package controller

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/yeying-community/router/common/client"
	"github.com/yeying-community/router/internal/admin/model"
	relaymeta "github.com/yeying-community/router/internal/relay/meta"
	relaymodel "github.com/yeying-community/router/internal/relay/model"
	"github.com/yeying-community/router/internal/relay/relaymode"
)

func TestGetRequestBodyMessagesPassThroughAppliesDropFields(t *testing.T) {
	t.Helper()
	c := newPolicyTestContext(t, `{
		"model":"claude-opus-4-6",
		"messages":[{"role":"user","content":"hello"}],
		"presence_penalty":0.5,
		"frequency_penalty":0.2
	}`)
	meta := &relaymeta.Meta{
		Mode:                relaymode.Messages,
		UpstreamMode:        relaymode.Messages,
		ActualModelName:     "claude-opus-4-6",
		UpstreamRequestPath: model.ChannelModelEndpointMessages,
		EndpointPolicy: &model.ChannelModelEndpointPolicy{
			ID:            "policy-1",
			Enabled:       true,
			Endpoint:      model.ChannelModelEndpointMessages,
			RequestPolicy: `{"actions":[{"type":"drop_fields","fields":["presence_penalty","frequency_penalty"],"reason":"drop legacy penalties"}]}`,
		},
	}
	reader, err := getRequestBody(c, meta, &relaymodel.GeneralOpenAIRequest{}, nil)
	if err != nil {
		t.Fatalf("getRequestBody returned error: %v", err)
	}
	raw, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("io.ReadAll returned error: %v", err)
	}
	payload := map[string]any{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	if _, ok := payload["presence_penalty"]; ok {
		t.Fatalf("presence_penalty still exists: %#v", payload["presence_penalty"])
	}
	if _, ok := payload["frequency_penalty"]; ok {
		t.Fatalf("frequency_penalty still exists: %#v", payload["frequency_penalty"])
	}
}

func TestGetRequestBodyMessagesPassThroughRejectsUnsupportedImageURL(t *testing.T) {
	t.Helper()
	c := newPolicyTestContext(t, `{
		"model":"claude-opus-4-6",
		"messages":[{"role":"user","content":[{"type":"image","source":{"type":"url","url":"https://example.com/a.png"}}]}]
	}`)
	meta := &relaymeta.Meta{
		Mode:                relaymode.Messages,
		UpstreamMode:        relaymode.Messages,
		ActualModelName:     "claude-opus-4-6",
		UpstreamRequestPath: model.ChannelModelEndpointMessages,
		EndpointPolicy: &model.ChannelModelEndpointPolicy{
			ID:            "policy-2",
			Enabled:       true,
			Endpoint:      model.ChannelModelEndpointMessages,
			RequestPolicy: `{"actions":[{"type":"reject_unsupported_input","input_types":["anthropic.image_url"],"reason":"image url unsupported"}]}`,
		},
	}
	_, err := getRequestBody(c, meta, &relaymodel.GeneralOpenAIRequest{}, nil)
	if err == nil {
		t.Fatalf("getRequestBody returned nil error, want endpoint policy error")
	}
	policyErr, ok := err.(*endpointPolicyError)
	if !ok {
		t.Fatalf("error type = %T, want *endpointPolicyError", err)
	}
	if policyErr.ErrorCode() != "unsupported_input" {
		t.Fatalf("policyErr.ErrorCode() = %q, want unsupported_input", policyErr.ErrorCode())
	}
}

func TestGetRequestBodyMessagesPassThroughConvertsAnthropicImageURLToBase64(t *testing.T) {
	t.Helper()
	mediaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte{
			0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
			0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
			0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
			0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
			0x89, 0x00, 0x00, 0x00, 0x0d, 0x49, 0x44, 0x41,
			0x54, 0x78, 0x9c, 0x63, 0xf8, 0xcf, 0xc0, 0x00,
			0x00, 0x03, 0x01, 0x01, 0x00, 0xc9, 0xfe, 0x92,
			0xef, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e,
			0x44, 0xae, 0x42, 0x60, 0x82,
		})
	}))
	defer mediaServer.Close()

	originalClient := client.UserContentRequestHTTPClient
	originalValidateHost := validateEndpointPolicyFetchHost
	client.UserContentRequestHTTPClient = mediaServer.Client()
	client.UserContentRequestHTTPClient.Timeout = 2 * time.Second
	defer func() {
		client.UserContentRequestHTTPClient = originalClient
		validateEndpointPolicyFetchHost = originalValidateHost
	}()
	validateEndpointPolicyFetchHost = func(_ context.Context, _ string) error {
		return nil
	}

	body := `{
		"model":"claude-opus-4-6",
		"messages":[{"role":"user","content":[{"type":"image","source":{"type":"url","url":"` + mediaServer.URL + `/img.png"}}]}]
	}`
	c := newPolicyTestContext(t, body)
	meta := &relaymeta.Meta{
		Mode:                relaymode.Messages,
		UpstreamMode:        relaymode.Messages,
		ActualModelName:     "claude-opus-4-6",
		ChannelId:           "channel-1",
		UpstreamRequestPath: model.ChannelModelEndpointMessages,
		EndpointPolicy: &model.ChannelModelEndpointPolicy{
			ID:            "policy-3",
			Enabled:       true,
			Endpoint:      model.ChannelModelEndpointMessages,
			RequestPolicy: `{"actions":[{"type":"image_url_to_base64","reason":"convert image url","limits":{"max_bytes":10240,"timeout_ms":2000,"allowed_content_types":["image/png"]}}]}`,
		},
	}
	reader, err := getRequestBody(c, meta, &relaymodel.GeneralOpenAIRequest{}, nil)
	if err != nil {
		t.Fatalf("getRequestBody returned error: %v", err)
	}
	raw, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("io.ReadAll returned error: %v", err)
	}
	payload := map[string]any{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	messages, ok := payload["messages"].([]any)
	if !ok || len(messages) != 1 {
		t.Fatalf("messages = %#v, want one message", payload["messages"])
	}
	messageItem, ok := messages[0].(map[string]any)
	if !ok {
		t.Fatalf("messages[0] = %#v, want map", messages[0])
	}
	contentList, ok := messageItem["content"].([]any)
	if !ok || len(contentList) != 1 {
		t.Fatalf("content = %#v, want one content block", messageItem["content"])
	}
	contentItem, ok := contentList[0].(map[string]any)
	if !ok {
		t.Fatalf("content[0] = %#v, want map", contentList[0])
	}
	source, ok := contentItem["source"].(map[string]any)
	if !ok {
		t.Fatalf("source = %#v, want map", contentItem["source"])
	}
	if got := strings.TrimSpace(source["type"].(string)); got != "base64" {
		t.Fatalf("source.type = %q, want base64", got)
	}
	if got := strings.TrimSpace(source["media_type"].(string)); got != "image/png" {
		t.Fatalf("source.media_type = %q, want image/png", got)
	}
	data := strings.TrimSpace(source["data"].(string))
	if data == "" {
		t.Fatalf("source.data is empty")
	}
	if strings.Contains(data, mediaServer.URL) {
		t.Fatalf("source.data still contains original url: %q", data)
	}
}

func newPolicyTestContext(t *testing.T, body string) *gin.Context {
	t.Helper()
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	return c
}
