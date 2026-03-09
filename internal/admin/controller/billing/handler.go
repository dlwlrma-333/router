package billing

import (
	"github.com/gin-gonic/gin"
	"github.com/yeying-community/router/common/config"
	"github.com/yeying-community/router/common/ctxkey"
	"github.com/yeying-community/router/internal/admin/model"
	billingsvc "github.com/yeying-community/router/internal/admin/service/billing"
	relaymodel "github.com/yeying-community/router/internal/relay/model"
)

func GetSubscription(c *gin.Context) {
	var remainQuota int64
	var usedQuota int64
	var err error
	var token *model.Token
	var expiredTime int64
	if config.DisplayTokenStatEnabled {
		tokenId := c.GetString(ctxkey.TokenId)
		if tokenId != "" {
			token, err = billingsvc.GetTokenByID(tokenId)
			if err == nil {
				expiredTime = token.ExpiredTime
				remainQuota = token.RemainQuota
				usedQuota = token.UsedQuota
			}
		}
	}
	if token == nil {
		userId := c.GetString(ctxkey.Id)
		remainQuota, err = billingsvc.GetUserQuota(userId)
		if err != nil {
			usedQuota, err = billingsvc.GetUserUsedQuota(userId)
		}
	}
	if expiredTime <= 0 {
		expiredTime = 0
	}
	if err != nil {
		Error := relaymodel.Error{
			Message: err.Error(),
			Type:    "upstream_error",
		}
		c.JSON(200, gin.H{
			"error": Error,
		})
		return
	}
	quota := remainQuota + usedQuota
	amount := float64(quota)
	if config.DisplayInCurrencyEnabled {
		amount /= config.QuotaPerUnit
	}
	if token != nil && token.UnlimitedQuota {
		amount = 100000000
	}
	subscription := billingsvc.OpenAISubscriptionResponse{
		Object:             "billing_subscription",
		HasPaymentMethod:   true,
		SoftLimitUSD:       amount,
		HardLimitUSD:       amount,
		SystemHardLimitUSD: amount,
		AccessUntil:        expiredTime,
	}
	c.JSON(200, subscription)
	return
}

func GetUsage(c *gin.Context) {
	var quota int64
	var err error
	var token *model.Token
	if config.DisplayTokenStatEnabled {
		tokenId := c.GetString(ctxkey.TokenId)
		if tokenId != "" {
			token, err = billingsvc.GetTokenByID(tokenId)
			if err == nil {
				quota = token.UsedQuota
			}
		}
	}
	if token == nil {
		userId := c.GetString(ctxkey.Id)
		quota, err = billingsvc.GetUserUsedQuota(userId)
	}
	if err != nil {
		Error := relaymodel.Error{
			Message: err.Error(),
			Type:    "one_api_error",
		}
		c.JSON(200, gin.H{
			"error": Error,
		})
		return
	}
	amount := float64(quota)
	if config.DisplayInCurrencyEnabled {
		amount /= config.QuotaPerUnit
	}
	usage := billingsvc.OpenAIUsageResponse{
		Object:     "list",
		TotalUsage: amount * 100,
	}
	c.JSON(200, usage)
	return
}
