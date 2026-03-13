package channel

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	common "github.com/yeying-community/router/common"
	"github.com/yeying-community/router/common/client"
	"github.com/yeying-community/router/common/helper"
	"github.com/yeying-community/router/common/logger"
	"github.com/yeying-community/router/internal/admin/model"
)

type testArtifactRecord struct {
	Path        string
	Name        string
	ContentType string
	Size        int64
}

var artifactFilenameSanitizer = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func channelTestArtifactRootDir() string {
	logDir := strings.TrimSpace(*common.LogDir)
	if logDir == "" {
		logDir = "./logs"
	}
	return filepath.Join(logDir, "artifacts", "channel-tests")
}

func sanitizeArtifactFilenamePart(value string) string {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return "artifact"
	}
	normalized = strings.ReplaceAll(normalized, "/", "-")
	normalized = strings.ReplaceAll(normalized, "\\", "-")
	normalized = artifactFilenameSanitizer.ReplaceAllString(normalized, "-")
	normalized = strings.Trim(normalized, "-._")
	if normalized == "" {
		return "artifact"
	}
	return normalized
}

func contentTypeMatchesExpected(contentType string, expectedKind string) bool {
	normalizedType := strings.ToLower(strings.TrimSpace(contentType))
	switch expectedKind {
	case model.ProviderModelTypeImage:
		return strings.HasPrefix(normalizedType, "image/")
	case model.ProviderModelTypeAudio:
		return strings.HasPrefix(normalizedType, "audio/")
	default:
		return false
	}
}

func detectArtifactContentType(data []byte, fallback string) string {
	if normalized := strings.TrimSpace(fallback); normalized != "" {
		return normalized
	}
	if len(data) == 0 {
		return "application/octet-stream"
	}
	sniffSize := len(data)
	if sniffSize > 512 {
		sniffSize = 512
	}
	return http.DetectContentType(data[:sniffSize])
}

func fileExtensionForArtifact(contentType string, fallback string) string {
	normalized := strings.TrimSpace(contentType)
	if normalized != "" {
		if extensions, err := mime.ExtensionsByType(normalized); err == nil && len(extensions) > 0 {
			return extensions[0]
		}
	}
	if fallback != "" {
		return fallback
	}
	return ".bin"
}

func decodeBase64Payload(raw string) ([]byte, string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, "", fmt.Errorf("empty base64 payload")
	}
	contentType := ""
	if strings.HasPrefix(trimmed, "data:") {
		parts := strings.SplitN(trimmed, ",", 2)
		if len(parts) != 2 {
			return nil, "", fmt.Errorf("invalid data url")
		}
		meta := parts[0]
		trimmed = parts[1]
		if strings.HasPrefix(meta, "data:") {
			contentType = strings.TrimPrefix(meta, "data:")
			contentType = strings.TrimSuffix(contentType, ";base64")
		}
	}
	decoded, err := base64.StdEncoding.DecodeString(trimmed)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(trimmed)
		if err != nil {
			return nil, "", err
		}
	}
	return decoded, contentType, nil
}

func downloadArtifactFromURL(ctx context.Context, rawURL string, expectedKind string) ([]byte, string, error) {
	trimmedURL := strings.TrimSpace(rawURL)
	if trimmedURL == "" {
		return nil, "", fmt.Errorf("empty artifact url")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, trimmedURL, nil)
	if err != nil {
		return nil, "", err
	}
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, "", fmt.Errorf("download artifact failed: http %d %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	contentType := detectArtifactContentType(body, resp.Header.Get("Content-Type"))
	if !contentTypeMatchesExpected(contentType, expectedKind) {
		return nil, "", fmt.Errorf("artifact content-type mismatch: %s", contentType)
	}
	return body, contentType, nil
}

func tryDecodeArtifactString(value string, expectedKind string) ([]byte, string, bool) {
	decoded, contentType, err := decodeBase64Payload(value)
	if err != nil || len(decoded) == 0 {
		return nil, "", false
	}
	detectedType := detectArtifactContentType(decoded, contentType)
	if !contentTypeMatchesExpected(detectedType, expectedKind) {
		return nil, "", false
	}
	return decoded, detectedType, true
}

func extractArtifactFromValue(ctx context.Context, value any, expectedKind string) ([]byte, string, error) {
	switch typed := value.(type) {
	case map[string]any:
		for _, key := range []string{"b64_json", "base64", "image_base64", "audio_base64", "audio", "data"} {
			raw, ok := typed[key].(string)
			if !ok {
				continue
			}
			if data, contentType, matched := tryDecodeArtifactString(raw, expectedKind); matched {
				return data, contentType, nil
			}
		}
		for _, key := range []string{"url", "image_url", "audio_url"} {
			raw, ok := typed[key].(string)
			if !ok {
				continue
			}
			data, contentType, err := downloadArtifactFromURL(ctx, raw, expectedKind)
			if err == nil {
				return data, contentType, nil
			}
		}
		for _, child := range typed {
			if data, contentType, err := extractArtifactFromValue(ctx, child, expectedKind); err == nil {
				return data, contentType, nil
			}
		}
	case []any:
		for _, child := range typed {
			if data, contentType, err := extractArtifactFromValue(ctx, child, expectedKind); err == nil {
				return data, contentType, nil
			}
		}
	case string:
		if data, contentType, matched := tryDecodeArtifactString(typed, expectedKind); matched {
			return data, contentType, nil
		}
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(typed)), "http://") || strings.HasPrefix(strings.ToLower(strings.TrimSpace(typed)), "https://") {
			return downloadArtifactFromURL(ctx, typed, expectedKind)
		}
	}
	return nil, "", fmt.Errorf("artifact not found")
}

func saveChannelTestArtifact(ctx context.Context, taskID string, result model.ChannelTest, expectedKind string, body []byte, fallbackContentType string) (*testArtifactRecord, error) {
	if len(body) == 0 {
		return nil, fmt.Errorf("empty artifact payload")
	}
	contentType := strings.TrimSpace(fallbackContentType)
	data := body
	if !contentTypeMatchesExpected(contentType, expectedKind) {
		var parsed any
		if err := json.Unmarshal(body, &parsed); err != nil {
			return nil, fmt.Errorf("artifact payload is not %s data: %w", expectedKind, err)
		}
		extractedData, extractedType, err := extractArtifactFromValue(ctx, parsed, expectedKind)
		if err != nil {
			return nil, err
		}
		data = extractedData
		contentType = extractedType
	}
	contentType = detectArtifactContentType(data, contentType)
	if !contentTypeMatchesExpected(contentType, expectedKind) {
		return nil, fmt.Errorf("artifact content-type mismatch: %s", contentType)
	}

	rootDir := channelTestArtifactRootDir()
	channelDir := filepath.Join(rootDir, sanitizeArtifactFilenamePart(result.ChannelId))
	if err := os.MkdirAll(channelDir, 0o755); err != nil {
		return nil, err
	}
	extension := fileExtensionForArtifact(contentType, "")
	fileName := fmt.Sprintf(
		"%d_%s_%s_%s%s",
		helper.GetTimestamp(),
		sanitizeArtifactFilenamePart(taskID),
		sanitizeArtifactFilenamePart(result.Model),
		sanitizeArtifactFilenamePart(result.Endpoint),
		extension,
	)
	absPath := filepath.Join(channelDir, fileName)
	if err := os.WriteFile(absPath, data, 0o644); err != nil {
		return nil, err
	}
	relativePath, err := filepath.Rel(rootDir, absPath)
	if err != nil {
		return nil, err
	}
	return &testArtifactRecord{
		Path:        filepath.ToSlash(relativePath),
		Name:        fileName,
		ContentType: contentType,
		Size:        int64(len(data)),
	}, nil
}

func isChannelTestArtifactPathSafe(relativePath string) (string, bool) {
	normalized := filepath.Clean(strings.TrimSpace(relativePath))
	if normalized == "." || normalized == "" || strings.HasPrefix(normalized, "..") || filepath.IsAbs(normalized) {
		return "", false
	}
	rootDir := channelTestArtifactRootDir()
	absPath := filepath.Join(rootDir, normalized)
	rel, err := filepath.Rel(rootDir, absPath)
	if err != nil || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return "", false
	}
	return absPath, true
}

func buildArtifactResponsePayloadForLog(statusCode int, header http.Header, artifact *testArtifactRecord) string {
	payload := map[string]any{
		"status_code":  statusCode,
		"headers":      sanitizeHTTPHeadersForLog(header),
		"body_omitted": true,
	}
	if artifact != nil {
		payload["artifact"] = map[string]any{
			"name":         artifact.Name,
			"content_type": artifact.ContentType,
			"size":         artifact.Size,
		}
	}
	return marshalJSONForLog(payload)
}

func attachArtifactMetadata(result *model.ChannelTest, artifact *testArtifactRecord) {
	if result == nil || artifact == nil {
		return
	}
	result.ArtifactPath = strings.TrimSpace(artifact.Path)
	result.ArtifactName = strings.TrimSpace(artifact.Name)
	result.ArtifactContentType = strings.TrimSpace(artifact.ContentType)
	result.ArtifactSize = artifact.Size
}

func tryPersistTestArtifact(ctx context.Context, taskID string, result *model.ChannelTest, expectedKind string, statusCode int, header http.Header, body []byte) string {
	if result == nil || len(body) == 0 {
		return ""
	}
	artifact, err := saveChannelTestArtifact(ctx, taskID, *result, expectedKind, body, header.Get("Content-Type"))
	if err != nil {
		logger.Warn(context.Background(), fmt.Sprintf("[channel-test-artifact] channel=%s model=%s endpoint=%s reason=%s", result.ChannelId, result.Model, result.Endpoint, err.Error()))
		return buildHTTPResponsePayloadForLog(statusCode, header, body)
	}
	attachArtifactMetadata(result, artifact)
	return buildArtifactResponsePayloadForLog(statusCode, header, artifact)
}

func persistChannelTestArtifactForExecution(ctx context.Context, taskID string, result *model.ChannelTest, execution *channelModelTestExecution) {
	if result == nil || execution == nil || execution.Err != nil || len(execution.ResponseBody) == 0 {
		return
	}
	var expectedKind string
	switch result.Type {
	case model.ProviderModelTypeImage:
		expectedKind = model.ProviderModelTypeImage
	case model.ProviderModelTypeAudio:
		expectedKind = model.ProviderModelTypeAudio
	default:
		return
	}
	execution.OutputPayload = tryPersistTestArtifact(
		ctx,
		taskID,
		result,
		expectedKind,
		execution.ResponseStatusCode,
		execution.ResponseHeader,
		execution.ResponseBody,
	)
}

func buildArtifactDownloadFilename(row model.ChannelTest) string {
	if normalized := strings.TrimSpace(row.ArtifactName); normalized != "" {
		return normalized
	}
	contentType := strings.TrimSpace(row.ArtifactContentType)
	return fmt.Sprintf(
		"%s_%s%s",
		sanitizeArtifactFilenamePart(row.Model),
		sanitizeArtifactFilenamePart(row.Endpoint),
		fileExtensionForArtifact(contentType, ".bin"),
	)
}

func detectContentTypeFromFile(data []byte, fallback string) string {
	if strings.TrimSpace(fallback) != "" {
		return fallback
	}
	return detectArtifactContentType(data, fallback)
}
