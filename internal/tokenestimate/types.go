package tokenestimate

import relaymodel "github.com/yeying-community/router/internal/relay/model"

type Precision string

const (
	PrecisionExact     Precision = "exact"
	PrecisionHeuristic Precision = "heuristic"
	PrecisionCoarse    Precision = "coarse"
)

type EstimateRequest struct {
	RelayMode int
	Model     string
	RawBody   []byte
	Request   *relaymodel.GeneralOpenAIRequest
}

type EstimateMeta struct {
	Texts         []string
	ToolTexts     []string
	ExtraTexts    []string
	ToolsCount    int
	NamesCount    int
	MessagesCount int
	MaxTokens     int
	Messages      []relaymodel.Message
	Input         any
}

type EstimateResult struct {
	PromptTokens int
	Source       string
	Precision    Precision
	Estimator    string
}
