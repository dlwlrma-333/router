package billing

import (
	"github.com/yeying-community/router/internal/admin/model"
	tokenrepo "github.com/yeying-community/router/internal/admin/repository/token"
	userrepo "github.com/yeying-community/router/internal/admin/repository/user"
)

type OpenAISubscriptionResponse struct {
	Object             string  `json:"object"`
	HasPaymentMethod   bool    `json:"has_payment_method"`
	SoftLimitUSD       float64 `json:"soft_limit_usd"`
	HardLimitUSD       float64 `json:"hard_limit_usd"`
	SystemHardLimitUSD float64 `json:"system_hard_limit_usd"`
	AccessUntil        int64   `json:"access_until"`
}

type OpenAIUsageResponse struct {
	Object string `json:"object"`
	//DailyCosts []OpenAIUsageDailyCost `json:"daily_costs"`
	TotalUsage float64 `json:"total_usage"` // unit: 0.01 dollar
}

func GetTokenByID(tokenId string) (*model.Token, error) {
	return tokenrepo.GetByID(tokenId)
}

func GetUserQuota(userId string) (int64, error) {
	return userrepo.GetQuota(userId)
}

func GetUserUsedQuota(userId string) (int64, error) {
	return userrepo.GetUsedQuota(userId)
}
