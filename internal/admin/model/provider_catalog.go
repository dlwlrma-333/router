package model

import (
	"sort"
	"strings"

	commonutils "github.com/yeying-community/router/common/utils"
	relaychannel "github.com/yeying-community/router/internal/relay/channel"
)

const (
	ProviderModelTypeText  = "text"
	ProviderModelTypeImage = "image"
	ProviderModelTypeAudio = "audio"
	ProviderModelTypeVideo = "video"

	ProviderPriceUnitPer1KTokens = "per_1k_tokens"
	ProviderPriceUnitPer1KChars  = "per_1k_chars"
	ProviderPriceUnitPerImage    = "per_image"
	ProviderPriceUnitPerVideo    = "per_video"

	ProviderPriceCurrencyUSD = "USD"
)

type ProviderModelDetail struct {
	Model       string  `json:"model"`
	Type        string  `json:"type,omitempty"`
	InputPrice  float64 `json:"input_price,omitempty"`
	OutputPrice float64 `json:"output_price,omitempty"`
	PriceUnit   string  `json:"price_unit,omitempty"`
	Currency    string  `json:"currency,omitempty"`
	Source      string  `json:"source,omitempty"`
	UpdatedAt   int64   `json:"updated_at,omitempty"`
}

type ProviderCatalogSeed struct {
	Provider     string
	Name         string
	BaseURL      string
	SortOrder    int
	ModelDetails []ProviderModelDetail
}

func NormalizeProviderModelDetails(details []ProviderModelDetail) []ProviderModelDetail {
	index := make(map[string]int, len(details))
	normalized := make([]ProviderModelDetail, 0, len(details))
	for _, detail := range details {
		modelName := strings.TrimSpace(detail.Model)
		if modelName == "" {
			continue
		}
		t := normalizeModelType(detail.Type, modelName)
		priceUnit := strings.TrimSpace(strings.ToLower(detail.PriceUnit))
		if priceUnit == "" {
			priceUnit = defaultPriceUnitByType(t, modelName)
		}
		currency := strings.TrimSpace(strings.ToUpper(detail.Currency))
		if currency == "" {
			currency = ProviderPriceCurrencyUSD
		}
		source := strings.TrimSpace(strings.ToLower(detail.Source))
		if source == "" {
			source = "manual"
		}
		inputPrice := detail.InputPrice
		if inputPrice < 0 {
			inputPrice = 0
		}
		outputPrice := detail.OutputPrice
		if outputPrice < 0 {
			outputPrice = 0
		}
		entry := ProviderModelDetail{
			Model:       modelName,
			Type:        t,
			InputPrice:  inputPrice,
			OutputPrice: outputPrice,
			PriceUnit:   priceUnit,
			Currency:    currency,
			Source:      source,
			UpdatedAt:   detail.UpdatedAt,
		}
		if idx, ok := index[modelName]; ok {
			existing := normalized[idx]
			if existing.Type == "" {
				existing.Type = entry.Type
			}
			if existing.PriceUnit == "" {
				existing.PriceUnit = entry.PriceUnit
			}
			if existing.Currency == "" {
				existing.Currency = entry.Currency
			}
			if existing.InputPrice <= 0 && entry.InputPrice > 0 {
				existing.InputPrice = entry.InputPrice
			}
			if existing.OutputPrice <= 0 && entry.OutputPrice > 0 {
				existing.OutputPrice = entry.OutputPrice
			}
			if entry.Source != "default" {
				existing.Source = entry.Source
			}
			if entry.UpdatedAt > existing.UpdatedAt {
				existing.UpdatedAt = entry.UpdatedAt
			}
			normalized[idx] = existing
			continue
		}
		index[modelName] = len(normalized)
		normalized = append(normalized, entry)
	}
	sort.Slice(normalized, func(i, j int) bool {
		return normalized[i].Model < normalized[j].Model
	})
	return normalized
}

func inferProviderByModel(modelName string, channelProtocol int, hasChannelProtocol bool) string {
	provider := commonutils.NormalizeProvider(commonutils.ResolveProvider(modelName))
	if provider != "" && provider != "unknown" {
		return provider
	}

	if strings.Contains(modelName, "/") {
		parts := strings.SplitN(modelName, "/", 2)
		prefix := commonutils.NormalizeProvider(parts[0])
		if prefix != "" && prefix != "unknown" {
			return prefix
		}
		plainPrefix := strings.TrimSpace(strings.ToLower(parts[0]))
		if plainPrefix != "" {
			return plainPrefix
		}
	}

	if hasChannelProtocol && channelProtocol > 0 && channelProtocol < len(relaychannel.ChannelProtocolNames) {
		rawProvider := strings.TrimSpace(relaychannel.ChannelProtocolNames[channelProtocol])
		normalized := commonutils.NormalizeProvider(rawProvider)
		if normalized != "" && normalized != "unknown" {
			return normalized
		}
		if rawProvider != "" && rawProvider != "unknown" {
			return strings.ToLower(rawProvider)
		}
	}

	lower := strings.ToLower(strings.TrimSpace(modelName))
	switch {
	case strings.HasPrefix(lower, "ernie-"):
		return "baidu"
	case strings.HasPrefix(lower, "spark-"):
		return "xunfei"
	case strings.HasPrefix(lower, "moonshot-") || strings.HasPrefix(lower, "kimi-"):
		return "moonshot"
	case strings.HasPrefix(lower, "baichuan-"):
		return "baichuan"
	case strings.HasPrefix(lower, "yi-"):
		return "lingyiwanwu"
	case strings.HasPrefix(lower, "step-"):
		return "stepfun"
	case strings.HasPrefix(lower, "groq-"):
		return "groq"
	case strings.HasPrefix(lower, "ollama-"):
		return "ollama"
	}
	return "other"
}

func normalizeModelType(raw string, modelName string) string {
	trimmed := strings.TrimSpace(strings.ToLower(raw))
	switch trimmed {
	case ProviderModelTypeText, ProviderModelTypeImage, ProviderModelTypeAudio, ProviderModelTypeVideo:
		return trimmed
	}
	lower := strings.ToLower(strings.TrimSpace(modelName))
	if lower == "" {
		return ProviderModelTypeText
	}
	switch {
	case strings.HasPrefix(lower, "veo"),
		strings.Contains(lower, "text-to-video"),
		strings.Contains(lower, "video-generation"),
		strings.Contains(lower, "video_generation"),
		strings.Contains(lower, "video"):
		return ProviderModelTypeVideo
	}
	if isKnownImageModel(modelName) {
		return ProviderModelTypeImage
	}
	switch {
	case strings.Contains(lower, "whisper"),
		strings.HasPrefix(lower, "tts-"),
		strings.Contains(lower, "audio"):
		return ProviderModelTypeAudio
	case strings.HasPrefix(lower, "dall-e"),
		strings.HasPrefix(lower, "cogview"),
		strings.Contains(lower, "stable-diffusion"),
		strings.HasPrefix(lower, "wanx"),
		strings.HasPrefix(lower, "step-1x"),
		strings.Contains(lower, "flux"):
		return ProviderModelTypeImage
	default:
		return ProviderModelTypeText
	}
}

func isKnownImageModel(modelName string) bool {
	switch strings.TrimSpace(strings.ToLower(modelName)) {
	case "dall-e-2",
		"dall-e-3",
		"ali-stable-diffusion-xl",
		"ali-stable-diffusion-v1.5",
		"wanx-v1",
		"cogview-3",
		"step-1x-medium":
		return true
	default:
		return false
	}
}

func InferModelType(modelName string) string {
	return normalizeModelType("", modelName)
}

func defaultPriceUnitByType(modelType string, modelName string) string {
	t := normalizeModelType(modelType, modelName)
	switch t {
	case ProviderModelTypeImage:
		return ProviderPriceUnitPerImage
	case ProviderModelTypeVideo:
		return ProviderPriceUnitPerVideo
	case ProviderModelTypeAudio:
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(modelName)), "tts-") {
			return ProviderPriceUnitPer1KChars
		}
		return ProviderPriceUnitPer1KTokens
	default:
		return ProviderPriceUnitPer1KTokens
	}
}
