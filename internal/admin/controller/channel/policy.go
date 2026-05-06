package channel

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/yeying-community/router/internal/admin/model"
)

type updateChannelEndpointPolicyRequest struct {
	ID             string `json:"id"`
	Model          string `json:"model"`
	Endpoint       string `json:"endpoint"`
	Enabled        *bool  `json:"enabled"`
	TemplateKey    string `json:"template_key"`
	Capabilities   string `json:"capabilities"`
	RequestPolicy  string `json:"request_policy"`
	ResponsePolicy string `json:"response_policy"`
	Reason         string `json:"reason"`
	Source         string `json:"source"`
	LastVerifiedAt int64  `json:"last_verified_at"`
}

// GetChannelEndpointPolicies godoc
// @Summary List channel endpoint policies (admin)
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Param id path string true "Channel ID"
// @Param model query string false "Model"
// @Param endpoint query string false "Endpoint"
// @Success 200 {object} docs.StandardResponse
// @Failure 401 {object} docs.ErrorResponse
// @Router /api/v1/admin/channel/{id}/policies [get]
func GetChannelEndpointPolicies(c *gin.Context) {
	channelID := strings.TrimSpace(c.Param("id"))
	if channelID == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "渠道 ID 无效",
		})
		return
	}
	modelName := strings.TrimSpace(c.Query("model"))
	endpoint := strings.TrimSpace(c.Query("endpoint"))
	rows, err := model.ListChannelModelEndpointPoliciesByChannelIDWithDB(model.DB, channelID, modelName, endpoint)
	if err != nil {
		logChannelAdminWarn(c, "list_endpoint_policies", stringField("channel_id", channelID), stringField("model", modelName), stringField("endpoint", endpoint), stringField("reason", err.Error()))
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"items": rows,
			"total": len(rows),
		},
	})
}

// UpdateChannelEndpointPolicy godoc
// @Summary Upsert channel endpoint policy (admin)
// @Tags admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Channel ID"
// @Param body body docs.StandardResponse true "Endpoint policy payload"
// @Success 200 {object} docs.StandardResponse
// @Failure 401 {object} docs.ErrorResponse
// @Router /api/v1/admin/channel/{id}/policies [put]
func UpdateChannelEndpointPolicy(c *gin.Context) {
	channelID := strings.TrimSpace(c.Param("id"))
	if channelID == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "渠道 ID 无效",
		})
		return
	}
	req := updateChannelEndpointPolicyRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		logChannelAdminWarn(c, "update_endpoint_policy", stringField("channel_id", channelID), stringField("reason", err.Error()))
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	modelName := strings.TrimSpace(req.Model)
	endpoint := model.NormalizeRequestedChannelModelEndpoint(req.Endpoint)
	if modelName == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "model 不能为空",
		})
		return
	}
	if endpoint == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "endpoint 无效",
		})
		return
	}
	endpointRows, err := model.ListChannelModelEndpointCandidatesByChannelIDWithDB(model.DB, channelID, modelName, endpoint)
	if err != nil {
		logChannelAdminWarn(c, "update_endpoint_policy", stringField("channel_id", channelID), stringField("model", modelName), stringField("endpoint", endpoint), stringField("reason", err.Error()))
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if !model.HasChannelModelEndpoint(endpointRows, modelName, endpoint) {
		message := "该渠道模型未声明该端点，不能保存策略"
		logChannelAdminWarn(c, "update_endpoint_policy", stringField("channel_id", channelID), stringField("model", modelName), stringField("endpoint", endpoint), stringField("reason", message))
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": message,
		})
		return
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	row := model.ChannelModelEndpointPolicy{
		ID:             strings.TrimSpace(req.ID),
		ChannelId:      channelID,
		Model:          modelName,
		Endpoint:       endpoint,
		Enabled:        enabled,
		TemplateKey:    strings.TrimSpace(req.TemplateKey),
		Capabilities:   strings.TrimSpace(req.Capabilities),
		RequestPolicy:  strings.TrimSpace(req.RequestPolicy),
		ResponsePolicy: strings.TrimSpace(req.ResponsePolicy),
		Reason:         strings.TrimSpace(req.Reason),
		Source:         strings.TrimSpace(req.Source),
		LastVerifiedAt: req.LastVerifiedAt,
	}
	saved, err := model.UpsertChannelModelEndpointPolicyWithDB(model.DB, row)
	if err != nil {
		logChannelAdminWarn(c, "update_endpoint_policy", stringField("channel_id", channelID), stringField("model", modelName), stringField("endpoint", endpoint), stringField("reason", err.Error()))
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	logChannelAdminInfo(c, "update_endpoint_policy", stringField("channel_id", channelID), stringField("model", modelName), stringField("endpoint", endpoint), stringField("enabled", map[bool]string{true: "true", false: "false"}[enabled]))
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    saved,
	})
}
