package citysocialrate

import (
	"neverbroke/common/logger"
	"neverbroke/internal/http/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler 城市社保比例处理器
type Handler struct{}

// NewHandler 创建城市社保比例处理器
func NewHandler() *Handler {
	return &Handler{}
}

// ListCitySocialRatesHandler GET /city-social-rates
// 返回所有城市的当前生效社保比例列表
func (h *Handler) ListCitySocialRatesHandler(c *gin.Context) {
	rates := ListRates()

	// 只返回 active 的记录
	activeRates := []CitySocialRate{}
	for _, rate := range rates {
		if rate.IsActive {
			activeRates = append(activeRates, rate)
		}
	}

	logger.Info("List city social rates", zap.Int("count", len(activeRates)))
	response.OK(c, activeRates)
}

// GetCitySocialRateDetailHandler GET /city-social-rates/:cityCode
// 返回指定城市的社保比例详情(含各项拆分)
func (h *Handler) GetCitySocialRateDetailHandler(c *gin.Context) {
	cityCode := c.Param("cityCode")
	if cityCode == "" {
		response.BadRequest(c, "cityCode is required")
		return
	}

	detail := GetRateDetail(cityCode)
	if detail == nil {
		response.BadRequest(c, "city not found")
		return
	}

	logger.Info("Get city social rate detail",
		zap.String("cityCode", cityCode),
		zap.Float64("totalRate", detail.TotalRate),
	)
	response.OK(c, detail)
}
