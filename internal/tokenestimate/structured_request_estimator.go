package tokenestimate

import "fmt"

// Structured-request estimators are used by request-backed providers and
// request-backed fallback paths. Anthropic raw messages estimation is excluded.
func estimateOpenAIFromRequest(req EstimateRequest, model string) (EstimateResult, error) {
	if req.Request == nil {
		return EstimateResult{}, fmt.Errorf("structured estimate request is nil")
	}
	meta := extractStructuredMeta(req)
	return EstimateResult{
		PromptTokens: estimateOpenAIExact(meta, model, req.RelayMode),
		Source:       "local_openai_tokenizer",
		Precision:    PrecisionExact,
		Estimator:    "openai_exact",
	}, nil
}

func estimateHeuristicFromRequest(req EstimateRequest, family providerFamily, source string, estimator string) (EstimateResult, error) {
	if req.Request == nil {
		return EstimateResult{}, fmt.Errorf("structured estimate request is nil")
	}
	meta := extractStructuredMeta(req)
	return EstimateResult{
		PromptTokens: estimateTextsHeuristic(meta.Texts, family),
		Source:       source,
		Precision:    PrecisionHeuristic,
		Estimator:    estimator,
	}, nil
}
