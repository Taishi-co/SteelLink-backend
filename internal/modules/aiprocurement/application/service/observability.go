package service

import (
	"github.com/Taishi-co/SteelLink-backend/pkg/logger"
	"go.uber.org/zap"
)

func logAIProcurementInfo(msg string, fields ...zap.Field) {
	if logger.Logger == nil {
		return
	}
	logger.Info(msg, fields...)
}

func logAIProcurementWarn(msg string, fields ...zap.Field) {
	if logger.Logger == nil {
		return
	}
	logger.Warn(msg, fields...)
}
