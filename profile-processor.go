package symbolizationprocessor

import (
	"context"

	//"time"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/xconsumer"
	"go.opentelemetry.io/collector/pdata/pprofile"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/symbolizationprocessor/profileutils"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/symbolizationprocessor/symbolizer"
)

type symbolizationProcessor struct {
	host         component.Host
	cancel       context.CancelFunc
	logger       *zap.Logger
	nextConsumer xconsumer.Profiles
	config       *Config
}

func (symbolizationprocessorProc *symbolizationProcessor) Start(ctx context.Context, host component.Host) error {
	symbolizationprocessorProc.host = host
	ctx = context.Background()
	ctx, symbolizationprocessorProc.cancel = context.WithCancel(ctx)

	/*interval, _ := time.ParseDuration(symbolizationprocessorProc.config.Interval)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
				case <-ticker.C:
					symbolizationprocessorProc.logger.Info("I should start processing profiles now!")
				case <-ctx.Done():
					return
			}
		}
	}()*/

	return nil

}

func (symbolizationprocessorProc *symbolizationProcessor) Shutdown(ctx context.Context) error {
	if symbolizationprocessorProc.cancel != nil {
		symbolizationprocessorProc.cancel()
	}
	return nil
}

func (symbolizationprocessorProc *symbolizationProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}

func (symbolizationprocessorProc *symbolizationProcessor) ConsumeProfiles(ctx context.Context, td pprofile.Profiles) error {
	symbolizationprocessorProc.logger.Info("Received new profiles!", zap.Int("num_profiles", td.SampleCount()))
	profilesDict := td.ProfilesDictionary()
	locationTable := profilesDict.LocationTable()
	attributeTable := profilesDict.AttributeTable()
	functionTable := profilesDict.FunctionTable()
	stringTable := profilesDict.StringTable()
	symbolAdder := profileutils.NewSymbolAdder(stringTable, functionTable, attributeTable)
	symbolizer := symbolizer.NewSymbolizer()
	defer symbolizer.Free()
	for _, resourceProfile := range td.ResourceProfiles().All() {
		for _, scopeProfile := range resourceProfile.ScopeProfiles().All() {
			for _, profile := range scopeProfile.Profiles().All() {
				locationIndices := profile.LocationIndices()
				for _, sample := range profile.Sample().All() {
					pid := profileutils.GetPid(sample, attributeTable)

					if pid == -1 {
						continue // Skip if PID is not found
					}

					for sampleLocationIdx := 0; sampleLocationIdx < int(sample.LocationsLength()); sampleLocationIdx++ {
						locationIdx := locationIndices.At(int(sample.LocationsStartIndex()) + sampleLocationIdx)
						location := locationTable.At(int(locationIdx))
						if location.Line().Len() == 0 {
							address := location.Address()
							symbol, err := symbolizer.Symbolize(pid, address)
							if err != nil {
								symbolizationprocessorProc.logger.Error(
									"Failed to symbolize address",
									zap.Int("pid", pid),
									zap.Uint64("address", address),
									zap.Error(err),
								)
							} else {
								functionName := symbol.Name
								lineNo := symbol.CodeInfo.Line
								filename := symbol.CodeInfo.File
								symbolAdder.AddSymbol(filename, functionName, lineNo, location)
							}
						}
					}
				}
			}
		}
	}
	return symbolizationprocessorProc.nextConsumer.ConsumeProfiles(ctx, td)
}
