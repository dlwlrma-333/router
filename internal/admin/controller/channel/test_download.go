package channel

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yeying-community/router/internal/admin/model"
)

// DownloadChannelTestArtifact godoc
// @Summary Download channel test artifact (admin)
// @Tags admin
// @Security BearerAuth
// @Produce application/octet-stream
// @Param id path string true "Channel ID"
// @Param model query string true "Model"
// @Param endpoint query string true "Endpoint"
// @Success 200 {file} binary
// @Failure 401 {object} docs.ErrorResponse
// @Router /api/v1/admin/channel/{id}/tests/artifact [get]
func DownloadChannelTestArtifact(c *gin.Context) {
	channelID := strings.TrimSpace(c.Param("id"))
	modelID := strings.TrimSpace(c.Query("model"))
	endpoint := strings.TrimSpace(c.Query("endpoint"))
	if channelID == "" || modelID == "" || endpoint == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "渠道 ID、模型或端点无效",
		})
		return
	}
	testRow, err := model.GetLatestChannelTestByModelEndpointWithDB(
		model.DB,
		channelID,
		modelID,
		endpoint,
	)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if strings.TrimSpace(testRow.ArtifactPath) == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "当前测试结果没有可下载文件",
		})
		return
	}
	absPath, ok := isChannelTestArtifactPathSafe(testRow.ArtifactPath)
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "测试文件路径无效",
		})
		return
	}
	data, err := os.ReadFile(absPath)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.Header("Content-Disposition", `attachment; filename="`+buildArtifactDownloadFilename(testRow)+`"`)
	c.Data(http.StatusOK, detectContentTypeFromFile(data, testRow.ArtifactContentType), data)
}
