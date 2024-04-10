package logging

import "go.uber.org/zap"

func GetLogger() *zap.Logger {
	return zap.Must(zap.NewDevelopmentConfig().Build())
}
