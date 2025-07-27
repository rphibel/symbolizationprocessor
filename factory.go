package symbolizationprocessor

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/xconsumer"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/xprocessor"
)

var (
	typeStr         = component.MustNewType("symbolizationprocessor")
)

func createDefaultConfig() component.Config {
	return &Config{
	}
}

func createProfilesProcessor(_ context.Context, params processor.Settings, baseCfg component.Config, consumer xconsumer.Profiles) (xprocessor.Profiles, error) {

	logger := params.Logger
	symbolizationprocessorCfg := baseCfg.(*Config)

	profileProc := &symbolizationProcessor{
		logger:       logger,
		nextConsumer: consumer,
		config:       symbolizationprocessorCfg,
	}

	return profileProc, nil
}

// NewFactory creates a factory for symbolizationprocessor processor.
func NewFactory() xprocessor.Factory {
	return xprocessor.NewFactory(
		typeStr,
		createDefaultConfig,
		xprocessor.WithProfiles(createProfilesProcessor, component.StabilityLevelAlpha))
}

