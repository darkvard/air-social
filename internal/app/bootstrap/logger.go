package bootstrap

import "go.uber.org/zap"

func NewLogger(env string) (*zap.SugaredLogger, error) {
	var baseLogger *zap.Logger
	var err error
	if env == "development" {
		baseLogger, err = zap.NewDevelopment()
	} else {
		baseLogger, err = zap.NewProduction()
	}
	if err != nil {
		return nil, err
	}
	sugar := baseLogger.Sugar()
	return sugar, nil
}
