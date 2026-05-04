package controller

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/yeying-community/router/common/client"
	"github.com/yeying-community/router/common/config"
	"github.com/yeying-community/router/common/ctxkey"
	"github.com/yeying-community/router/common/logger"
	adminmodel "github.com/yeying-community/router/internal/admin/model"
	relaymeta "github.com/yeying-community/router/internal/relay/meta"
)

const (
	endpointPolicyDefaultMediaMaxBytes = 5 * 1024 * 1024
	endpointPolicyDefaultTimeoutMs     = 10 * 1000
)

var validateEndpointPolicyFetchHost = validatePolicyFetchHost

type endpointPolicyError struct {
	code   string
	status int
	err    error
}

func (e *endpointPolicyError) Error() string {
	if e == nil || e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (e *endpointPolicyError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

func (e *endpointPolicyError) StatusCode() int {
	if e == nil || e.status == 0 {
		return http.StatusBadRequest
	}
	return e.status
}

func (e *endpointPolicyError) ErrorCode() string {
	if e == nil || strings.TrimSpace(e.code) == "" {
		return "endpoint_policy_failed"
	}
	return e.code
}

func newEndpointPolicyError(code string, status int, format string, args ...any) error {
	return &endpointPolicyError{
		code:   strings.TrimSpace(code),
		status: status,
		err:    fmt.Errorf(format, args...),
	}
}

type endpointPolicyReport struct {
	policyID      string
	actions       []string
	changedFields []string
	reasons       []string
}

func (r *endpointPolicyReport) addAction(actionType string, reason string) {
	actionType = strings.TrimSpace(actionType)
	if actionType != "" {
		r.actions = append(r.actions, actionType)
	}
	reason = strings.TrimSpace(reason)
	if reason != "" {
		r.reasons = append(r.reasons, reason)
	}
}

func (r *endpointPolicyReport) addChangedField(field string) {
	field = strings.TrimSpace(field)
	if field != "" {
		r.changedFields = append(r.changedFields, field)
	}
}

type fetchedPolicyMedia struct {
	MIMEType string
	Base64   string
	DataURL  string
}

func applyEndpointRequestPolicy(c *gin.Context, meta *relaymeta.Meta, raw []byte) ([]byte, error) {
	if len(raw) == 0 || meta == nil || meta.EndpointPolicy == nil || !meta.EndpointPolicy.Enabled {
		return raw, nil
	}
	requestPolicy, err := meta.EndpointPolicy.ParseRequestPolicy()
	if err != nil {
		return nil, newEndpointPolicyError("invalid_policy", http.StatusInternalServerError, "parse endpoint policy failed: %v", err)
	}
	if len(requestPolicy.Actions) == 0 {
		return raw, nil
	}
	payload := map[string]any{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, newEndpointPolicyError("invalid_policy", http.StatusInternalServerError, "parse request payload failed: %v", err)
	}
	report := &endpointPolicyReport{policyID: strings.TrimSpace(meta.EndpointPolicy.ID)}
	mediaCache := make(map[string]fetchedPolicyMedia)
	changed := false
	for _, action := range requestPolicy.Actions {
		actionChanged, actionErr := applyEndpointPolicyAction(c, meta, payload, action, report, mediaCache)
		if actionErr != nil {
			return nil, actionErr
		}
		if actionChanged {
			changed = true
		}
	}
	if !changed {
		return raw, nil
	}
	updatedRaw, err := json.Marshal(payload)
	if err != nil {
		return nil, newEndpointPolicyError("endpoint_policy_failed", http.StatusInternalServerError, "marshal endpoint-adjusted request failed: %v", err)
	}
	c.Set(ctxkey.KeyRequestBody, updatedRaw)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(updatedRaw))
	logEndpointPolicyReport(c, meta, report)
	return updatedRaw, nil
}

func applyEndpointPolicyAction(
	c *gin.Context,
	meta *relaymeta.Meta,
	payload map[string]any,
	action adminmodel.ChannelModelEndpointPolicyAction,
	report *endpointPolicyReport,
	mediaCache map[string]fetchedPolicyMedia,
) (bool, error) {
	actionType := strings.TrimSpace(action.Type)
	switch actionType {
	case "":
		return false, nil
	case adminmodel.ChannelEndpointPolicyActionDropFields:
		return applyEndpointPolicyDropFields(payload, action, report), nil
	case adminmodel.ChannelEndpointPolicyActionRejectUnsupportedInput:
		return false, applyEndpointPolicyRejectUnsupportedInput(payload, action, report)
	case adminmodel.ChannelEndpointPolicyActionImageURLToBase64:
		return applyEndpointPolicyImageURLToBase64(c, meta, payload, action, report, mediaCache)
	default:
		return false, newEndpointPolicyError("invalid_policy", http.StatusInternalServerError, "unsupported endpoint policy action %q", actionType)
	}
}

func applyEndpointPolicyDropFields(payload map[string]any, action adminmodel.ChannelModelEndpointPolicyAction, report *endpointPolicyReport) bool {
	if payload == nil || len(action.Fields) == 0 {
		return false
	}
	changed := false
	for _, field := range action.Fields {
		if _, ok := payload[field]; !ok {
			continue
		}
		delete(payload, field)
		report.addChangedField(field)
		changed = true
	}
	if changed {
		report.addAction(action.Type, action.Reason)
	}
	return changed
}

func applyEndpointPolicyRejectUnsupportedInput(payload map[string]any, action adminmodel.ChannelModelEndpointPolicyAction, report *endpointPolicyReport) error {
	inputTypes := normalizePolicyInputTypes(action.InputTypes)
	if len(inputTypes) == 0 {
		return nil
	}
	for _, detected := range collectPolicyInputKinds(payload) {
		if _, ok := inputTypes[detected]; ok {
			report.addAction(action.Type, action.Reason)
			return newEndpointPolicyError(
				"unsupported_input",
				http.StatusBadRequest,
				"endpoint policy rejected unsupported input %q: %s",
				detected,
				strings.TrimSpace(action.Reason),
			)
		}
	}
	return nil
}

func applyEndpointPolicyImageURLToBase64(
	c *gin.Context,
	meta *relaymeta.Meta,
	payload map[string]any,
	action adminmodel.ChannelModelEndpointPolicyAction,
	report *endpointPolicyReport,
	mediaCache map[string]fetchedPolicyMedia,
) (bool, error) {
	inputTypes := normalizePolicyInputTypes(action.InputTypes)
	changed := false
	err := walkPolicyPayload(payload, func(path string, node map[string]any) error {
		kind, rawURL, setter := detectImageURLPolicyNode(node)
		if setter == nil || strings.TrimSpace(rawURL) == "" {
			return nil
		}
		if len(inputTypes) > 0 {
			if _, ok := inputTypes[kind]; !ok {
				return nil
			}
		}
		if cached, ok := mediaCache[rawURL]; ok {
			setter(cached)
			report.addChangedField(path)
			changed = true
			return nil
		}
		fetched, fetchErr := fetchPolicyMedia(c.Request.Context(), rawURL, action.Limits)
		if fetchErr != nil {
			return fetchErr
		}
		mediaCache[rawURL] = fetched
		setter(fetched)
		report.addChangedField(path)
		changed = true
		return nil
	})
	if err != nil {
		return false, err
	}
	if changed {
		report.addAction(action.Type, action.Reason)
		logger.Debugf(
			c.Request.Context(),
			"[endpoint_policy_image_url_to_base64] channel_id=%s model=%s endpoint=%s changed=%d",
			strings.TrimSpace(meta.ChannelId),
			strings.TrimSpace(meta.ActualModelName),
			strings.TrimSpace(meta.UpstreamRequestPath),
			len(report.changedFields),
		)
	}
	return changed, nil
}

func walkPolicyPayload(value any, visit func(path string, node map[string]any) error) error {
	return walkPolicyPayloadWithPath("", value, visit)
}

func walkPolicyPayloadWithPath(path string, value any, visit func(path string, node map[string]any) error) error {
	switch typed := value.(type) {
	case map[string]any:
		if err := visit(path, typed); err != nil {
			return err
		}
		for key, child := range typed {
			childPath := key
			if path != "" {
				childPath = path + "." + key
			}
			if err := walkPolicyPayloadWithPath(childPath, child, visit); err != nil {
				return err
			}
		}
	case []any:
		for idx, child := range typed {
			childPath := fmt.Sprintf("[%d]", idx)
			if path != "" {
				childPath = fmt.Sprintf("%s[%d]", path, idx)
			}
			if err := walkPolicyPayloadWithPath(childPath, child, visit); err != nil {
				return err
			}
		}
	}
	return nil
}

func normalizePolicyInputTypes(values []string) map[string]struct{} {
	normalized := make(map[string]struct{})
	for _, value := range values {
		key := strings.TrimSpace(strings.ToLower(value))
		if key == "" {
			continue
		}
		normalized[key] = struct{}{}
	}
	return normalized
}

func collectPolicyInputKinds(payload map[string]any) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0)
	_ = walkPolicyPayload(payload, func(_ string, node map[string]any) error {
		kind, _, _ := detectImageURLPolicyNode(node)
		if kind != "" {
			if _, ok := seen[kind]; !ok {
				seen[kind] = struct{}{}
				result = append(result, kind)
			}
		}
		return nil
	})
	return result
}

func detectImageURLPolicyNode(node map[string]any) (string, string, func(media fetchedPolicyMedia)) {
	nodeType := strings.TrimSpace(strings.ToLower(fmt.Sprint(node["type"])))
	switch nodeType {
	case "image":
		source, ok := node["source"].(map[string]any)
		if !ok {
			return "", "", nil
		}
		if strings.TrimSpace(strings.ToLower(fmt.Sprint(source["type"]))) != "url" {
			return "", "", nil
		}
		rawURL := strings.TrimSpace(fmt.Sprint(source["url"]))
		if rawURL == "" {
			return "", "", nil
		}
		return "anthropic.image_url", rawURL, func(media fetchedPolicyMedia) {
			node["source"] = map[string]any{
				"type":       "base64",
				"media_type": media.MIMEType,
				"data":       media.Base64,
			}
		}
	case "image_url":
		switch imageURL := node["image_url"].(type) {
		case string:
			rawURL := strings.TrimSpace(imageURL)
			if rawURL == "" || strings.HasPrefix(strings.ToLower(rawURL), "data:") {
				return "", "", nil
			}
			return "openai.image_url", rawURL, func(media fetchedPolicyMedia) {
				node["image_url"] = media.DataURL
			}
		case map[string]any:
			rawURL := strings.TrimSpace(fmt.Sprint(imageURL["url"]))
			if rawURL == "" || strings.HasPrefix(strings.ToLower(rawURL), "data:") {
				return "", "", nil
			}
			return "openai.image_url", rawURL, func(media fetchedPolicyMedia) {
				imageURL["url"] = media.DataURL
			}
		}
	case "input_image":
		rawURL := strings.TrimSpace(fmt.Sprint(node["image_url"]))
		if rawURL == "" || strings.HasPrefix(strings.ToLower(rawURL), "data:") {
			return "", "", nil
		}
		return "responses.input_image_url", rawURL, func(media fetchedPolicyMedia) {
			node["image_url"] = media.DataURL
		}
	}
	return "", "", nil
}

func fetchPolicyMedia(ctx context.Context, rawURL string, limits *adminmodel.ChannelModelEndpointPolicyActionLimit) (fetchedPolicyMedia, error) {
	parsedURL, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return fetchedPolicyMedia{}, newEndpointPolicyError("media_fetch_failed", http.StatusBadRequest, "invalid media url: %v", err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fetchedPolicyMedia{}, newEndpointPolicyError("unsupported_input", http.StatusBadRequest, "unsupported media url scheme %q", parsedURL.Scheme)
	}
	if err := validateEndpointPolicyFetchHost(ctx, parsedURL.Hostname()); err != nil {
		return fetchedPolicyMedia{}, err
	}
	timeoutMs := endpointPolicyDefaultTimeoutMs
	if limits != nil && limits.TimeoutMs > 0 {
		timeoutMs = limits.TimeoutMs
	} else if config.UserContentRequestTimeout > 0 {
		timeoutMs = config.UserContentRequestTimeout * 1000
	}
	maxBytes := int64(endpointPolicyDefaultMediaMaxBytes)
	if limits != nil && limits.MaxBytes > 0 {
		maxBytes = limits.MaxBytes
	}
	allowedTypes := map[string]struct{}{}
	if limits != nil {
		for _, item := range limits.AllowedContentTypes {
			normalized := normalizeMediaType(item)
			if normalized == "" {
				continue
			}
			allowedTypes[normalized] = struct{}{}
		}
	}

	reqCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return fetchedPolicyMedia{}, newEndpointPolicyError("media_fetch_failed", http.StatusBadRequest, "create media request failed: %v", err)
	}
	resp, err := client.UserContentRequestHTTPClient.Do(req)
	if err != nil {
		return fetchedPolicyMedia{}, newEndpointPolicyError("media_fetch_failed", http.StatusBadGateway, "fetch media failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fetchedPolicyMedia{}, newEndpointPolicyError("media_fetch_failed", http.StatusBadGateway, "fetch media failed with status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return fetchedPolicyMedia{}, newEndpointPolicyError("media_fetch_failed", http.StatusBadGateway, "read media failed: %v", err)
	}
	if int64(len(body)) > maxBytes {
		return fetchedPolicyMedia{}, newEndpointPolicyError("media_too_large", http.StatusBadRequest, "media size %d exceeds limit %d", len(body), maxBytes)
	}
	mediaType := normalizeMediaType(resp.Header.Get("Content-Type"))
	if mediaType == "" {
		mediaType = normalizeMediaType(http.DetectContentType(body))
	}
	if !strings.HasPrefix(mediaType, "image/") {
		return fetchedPolicyMedia{}, newEndpointPolicyError("unsupported_media_type", http.StatusBadRequest, "unsupported media type %q", mediaType)
	}
	if len(allowedTypes) > 0 {
		if _, ok := allowedTypes[mediaType]; !ok {
			return fetchedPolicyMedia{}, newEndpointPolicyError("unsupported_media_type", http.StatusBadRequest, "media type %q is not allowed by endpoint policy", mediaType)
		}
	}
	base64Value := base64.StdEncoding.EncodeToString(body)
	return fetchedPolicyMedia{
		MIMEType: mediaType,
		Base64:   base64Value,
		DataURL:  fmt.Sprintf("data:%s;base64,%s", mediaType, base64Value),
	}, nil
}

func validatePolicyFetchHost(ctx context.Context, hostname string) error {
	normalizedHost := strings.TrimSpace(strings.ToLower(hostname))
	if normalizedHost == "" {
		return newEndpointPolicyError("media_fetch_failed", http.StatusBadRequest, "media host is empty")
	}
	if normalizedHost == "localhost" {
		return newEndpointPolicyError("unsupported_input", http.StatusBadRequest, "localhost media url is not allowed")
	}
	if addr, err := netip.ParseAddr(normalizedHost); err == nil {
		if isBlockedPolicyFetchAddr(addr) {
			return newEndpointPolicyError("unsupported_input", http.StatusBadRequest, "private media address %q is not allowed", normalizedHost)
		}
		return nil
	}
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, normalizedHost)
	if err != nil {
		return newEndpointPolicyError("media_fetch_failed", http.StatusBadGateway, "resolve media host failed: %v", err)
	}
	for _, ip := range ips {
		addr, ok := netip.AddrFromSlice(ip.IP)
		if !ok {
			continue
		}
		if isBlockedPolicyFetchAddr(addr) {
			return newEndpointPolicyError("unsupported_input", http.StatusBadRequest, "private media host %q is not allowed", normalizedHost)
		}
	}
	return nil
}

func isBlockedPolicyFetchAddr(addr netip.Addr) bool {
	return addr.IsLoopback() || addr.IsPrivate() || addr.IsMulticast() || addr.IsLinkLocalUnicast() || addr.IsLinkLocalMulticast() || addr.IsUnspecified()
}

func normalizeMediaType(value string) string {
	mediaType, _, err := mime.ParseMediaType(strings.TrimSpace(value))
	if err != nil {
		return strings.TrimSpace(strings.ToLower(value))
	}
	return strings.TrimSpace(strings.ToLower(mediaType))
}

func logEndpointPolicyReport(c *gin.Context, meta *relaymeta.Meta, report *endpointPolicyReport) {
	if c == nil || meta == nil || report == nil {
		return
	}
	logger.Debugf(
		c.Request.Context(),
		"[endpoint_policy] channel_id=%s model=%s endpoint=%s policy_id=%s actions=%s changed_fields=%s reasons=%s",
		strings.TrimSpace(meta.ChannelId),
		strings.TrimSpace(meta.ActualModelName),
		strings.TrimSpace(meta.UpstreamRequestPath),
		strings.TrimSpace(report.policyID),
		strings.Join(report.actions, ","),
		strings.Join(report.changedFields, ","),
		strings.Join(report.reasons, " | "),
	)
}
