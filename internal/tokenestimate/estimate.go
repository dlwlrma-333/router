package tokenestimate

import "fmt"

func Estimate(req EstimateRequest) (EstimateResult, error) {
	model := resolveEstimateModel(req)
	if model == "" {
		return EstimateResult{}, fmt.Errorf("estimate model is empty")
	}
	return estimateByProvider(req, model)
}

func estimateTextsHeuristic(texts []string, family providerFamily) int {
	total := 0
	for _, text := range texts {
		total += estimateHeuristicText(text, family)
	}
	return total
}
