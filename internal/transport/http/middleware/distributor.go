package middleware

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/yeying-community/router/common/ctxkey"
	"github.com/yeying-community/router/common/logger"
	"github.com/yeying-community/router/internal/admin/model"
	relaychannel "github.com/yeying-community/router/internal/relay/channel"
	"github.com/yeying-community/router/internal/relay/relaymode"
)

type ModelRequest struct {
	Model string `json:"model" form:"model"`
}

func resolveRequestCapability(path string) string {
	switch relaymode.GetByPath(path) {
	case relaymode.Responses:
		return model.ChannelCapabilityResponses
	default:
		return ""
	}
}

func filterChannelsByCapabilityProfile(channels []*model.Channel, capability string, clientProfile string) []*model.Channel {
	if capability == "" {
		return channels
	}
	filtered := make([]*model.Channel, 0, len(channels))
	for _, channel := range channels {
		if channel == nil {
			continue
		}
		if channel.SupportsCapabilityClientProfile(capability, clientProfile) {
			filtered = append(filtered, channel)
		}
	}
	return filtered
}

func pickChannelByPriority(channels []*model.Channel, ignoreFirstPriority bool) *model.Channel {
	if len(channels) == 0 {
		return nil
	}
	endIdx := len(channels)
	firstPriority := channels[0].GetPriority()
	if firstPriority > 0 {
		for i := range channels {
			if channels[i].GetPriority() != firstPriority {
				endIdx = i
				break
			}
		}
	}
	targets := channels[:endIdx]
	if ignoreFirstPriority && endIdx < len(channels) {
		targets = channels[endIdx:]
	}
	if len(targets) == 0 {
		return nil
	}
	return targets[rand.Intn(len(targets))]
}

func Distribute() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		userId := c.GetString(ctxkey.Id)
		userGroup, _ := model.CacheGetUserGroup(userId)
		c.Set(ctxkey.Group, userGroup)
		requestCapability := resolveRequestCapability(c.Request.URL.Path)
		inboundUserAgent := strings.TrimSpace(c.Request.UserAgent())
		clientProfiles, err := model.ListEnabledClientProfilesWithDB(model.DB)
		if err != nil {
			abortWithMessage(c, http.StatusInternalServerError, "读取客户端画像失败")
			return
		}
		clientProfile := model.ResolveClientProfileByUserAgent(inboundUserAgent, clientProfiles)
		var requestModel string
		var channel *model.Channel
		channelId, ok := c.Get(ctxkey.SpecificChannelId)
		if ok {
			id := fmt.Sprintf("%v", channelId)
			channel, err = model.GetChannelById(id, true)
			if err != nil {
				abortWithMessage(c, http.StatusBadRequest, "无效的渠道 Id")
				return
			}
			if channel.Status != model.ChannelStatusEnabled {
				abortWithMessage(c, http.StatusForbidden, "该渠道已被禁用")
				return
			}
			if requestCapability != "" && !channel.SupportsCapabilityClientProfile(requestCapability, clientProfile) {
				abortWithMessage(c, http.StatusForbidden, "该渠道未开放当前客户端的 responses 能力")
				return
			}
		} else {
			requestModel = c.GetString(ctxkey.RequestModel)
			candidates, err := model.CacheListSatisfiedChannels(userGroup, requestModel)
			if err != nil {
				message := fmt.Sprintf("当前分组 %s 下对于模型 %s 无可用渠道", userGroup, requestModel)
				if channel != nil {
					logger.SysError(fmt.Sprintf("渠道不存在：%s", channel.Id))
					message = "数据库一致性已被破坏，请联系管理员"
				}
				abortWithMessage(c, http.StatusServiceUnavailable, message)
				return
			}
			filtered := filterChannelsByCapabilityProfile(candidates, requestCapability, clientProfile)
			channel = pickChannelByPriority(filtered, false)
			if channel == nil {
				message := fmt.Sprintf("当前分组 %s 下对于模型 %s 无可用渠道", userGroup, requestModel)
				if requestCapability == model.ChannelCapabilityResponses {
					message = fmt.Sprintf("当前分组 %s 下对于模型 %s 无可用 responses 渠道（client_profile=%s）", userGroup, requestModel, clientProfile)
				}
				abortWithMessage(c, http.StatusServiceUnavailable, message)
				return
			}
		}
		logger.Debugf(ctx, "user id %s, user group: %s, request model: %s, using channel #%s", userId, userGroup, requestModel, channel.Id)
		SetupContextForSelectedChannel(c, channel, requestModel)
		c.Next()
	}
}

func SetupContextForSelectedChannel(c *gin.Context, channel *model.Channel, modelName string) {
	channelProtocol := channel.GetChannelProtocol()
	c.Set(ctxkey.Channel, channelProtocol)
	c.Set(ctxkey.ChannelId, channel.Id)
	c.Set(ctxkey.ChannelName, channel.Name)
	if channel.SystemPrompt != nil && *channel.SystemPrompt != "" {
		c.Set(ctxkey.SystemPrompt, *channel.SystemPrompt)
	}
	if channel.ModelRatio != nil {
		c.Set(ctxkey.ModelRatio, *channel.ModelRatio)
	} else {
		c.Set(ctxkey.ModelRatio, "")
	}
	if channel.CompletionRatio != nil {
		c.Set(ctxkey.CompletionRatio, *channel.CompletionRatio)
	} else {
		c.Set(ctxkey.CompletionRatio, "")
	}
	c.Set(ctxkey.ModelMapping, channel.GetModelMapping())
	c.Set(ctxkey.OriginalModel, modelName) // for retry
	c.Request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", channel.Key))
	c.Set(ctxkey.BaseURL, channel.GetBaseURL())
	cfg, _ := channel.LoadConfig()
	// this is for backward compatibility
	if channel.Other != nil {
		switch channelProtocol {
		case relaychannel.Azure:
			if cfg.APIVersion == "" {
				cfg.APIVersion = *channel.Other
			}
		case relaychannel.Xunfei:
			if cfg.APIVersion == "" {
				cfg.APIVersion = *channel.Other
			}
		case relaychannel.Gemini:
			if cfg.APIVersion == "" {
				cfg.APIVersion = *channel.Other
			}
		case relaychannel.AIProxyLibrary:
			if cfg.LibraryID == "" {
				cfg.LibraryID = *channel.Other
			}
		case relaychannel.Ali:
			if cfg.Plugin == "" {
				cfg.Plugin = *channel.Other
			}
		}
	}
	c.Set(ctxkey.Config, cfg)
}
