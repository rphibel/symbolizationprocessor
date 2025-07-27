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
	mappingTable := profilesDict.MappingTable()
	// functionMap := make(map[pprofile.Function]int)
	// for idx, f := range functionTable.All() {
	// 	functionMap[f] = idx
	// }
	// stringMap := make(map[string]int)
	// for idx, s := range stringTable.All() {
	// 	stringMap[s] = idx
	// }
	symbolAdder := profileutils.NewSymbolAdder(stringTable, functionTable, attributeTable)
	symbolizer := symbolizer.NewSymbolizer()
	defer symbolizer.Free()
	for _, resourceProfile := range td.ResourceProfiles().All() {
		for _, scopeProfile := range resourceProfile.ScopeProfiles().All() {
			for _, profile := range scopeProfile.Profiles().All() {
				symbolizationprocessorProc.logger.Info("Processing profile", zap.String("profile", profile.ProfileID().String()))
				locationIndices := profile.LocationIndices()
				for _, sample := range profile.Sample().All() {
					symbolizationprocessorProc.logger.Info(
						"Processing sample",
					)
					pid := profileutils.GetPid(sample, attributeTable)

					for sampleLocationIdx := 0; sampleLocationIdx < int(sample.LocationsLength()); sampleLocationIdx++ {
						locationIdx := locationIndices.At(int(sample.LocationsStartIndex()) + sampleLocationIdx)
						location := locationTable.At(int(locationIdx))
						mappingIdx := location.MappingIndex()
						mapping := mappingTable.At(int(mappingIdx))
						exe := stringTable.At(int(mapping.FilenameStrindex()))
						symbolizationprocessorProc.logger.Info(
							"Location binary",
							zap.Int("LocationIdx", sampleLocationIdx),
							zap.String("name", exe),
							zap.Int("address", int(location.Address())),
							zap.Int("memory_start", int(mapping.MemoryStart())),
							zap.Int("memory_limit", int(mapping.MemoryLimit())),
							zap.Int("offset", int(mapping.FileOffset())),
						)
						for _, attributeIdx := range mapping.AttributeIndices().All() {
							symbolizationprocessorProc.logger.Info(
								"Location attribute",
								zap.Int("LocationIdx", sampleLocationIdx),
								zap.String("key", attributeTable.At(int(attributeIdx)).Key()),
								zap.String("value", attributeTable.At(int(attributeIdx)).Value().AsString()),
							)
						}
						if location.Line().Len() == 0 && pid != -1 {
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
								symbolizationprocessorProc.logger.Info(
									"DEBUGRP Symbolized location for "+exe,
									zap.String("function", symbol.Name),
									zap.String("file", symbol.CodeInfo.File),
									zap.Int("line", int(symbol.CodeInfo.Line)),
								)
								functionName := symbol.Name
								lineNo := symbol.CodeInfo.Line
								filename := symbol.CodeInfo.File
								symbolAdder.AddSymbol(filename, functionName, lineNo, location)
								//addSymbol(stringMap, functionName, stringTable, filename, functionMap, functionTable, location, lineNo, attributeTable)
							}
							// symbolizationprocessorProc.logger.Info(
							// 	"Location needs symbolization",
							// 	zap.Int("LocationIdx", sampleLocationIdx),
							// 	zap.String("name", exe),
							// )
							// functionName := "test_function"
							// filename := "test_file"
							// var nameIdx int
							// var ok bool
							// nameIdx, ok = stringMap[functionName]
							// if !ok {
							// 	stringTable.Append(functionName)
							// 	nameIdx = stringTable.Len() - 1
							// 	stringMap[functionName] = nameIdx
							// }
							// var filenameIdx int
							// filenameIdx, ok = stringMap[filename]
							// if !ok {
							// 	stringTable.Append(filename)
							// 	filenameIdx = stringTable.Len() - 1
							// 	stringMap[filename] = filenameIdx
							// }
							// function := pprofile.NewFunction()
							// function.SetNameStrindex(int32(nameIdx))
							// function.SetFilenameStrindex(int32(filenameIdx))
							// functionIdx, ok := functionMap[function]
							// if !ok {
							// 	functionTable.AppendEmpty()
							// 	functionIdx = functionTable.Len() - 1
							// 	function := functionTable.At(int(functionIdx))
							// 	functionMap[function] = functionIdx
							// }
							//
							// line := location.Line().AppendEmpty()
							// line.SetFunctionIndex(1)
							// line.SetLine(1)
							// // line.SetFunctionIndex(int32(functionIdx))
							// for _, attributeIdx := range location.AttributeIndices().All() {
							// 	key := attributeTable.At(int(attributeIdx)).Key()
							// 	if key == "profile.frame.type" {
							// 		attributeTable.At(int(attributeIdx)).Value().SetStr("kernel")
							// 	}
							// }
						}
						for _, line := range location.Line().All() {
							function := functionTable.At(int(line.FunctionIndex()))
							functionName := stringTable.At(int(function.NameStrindex()))
							filename := stringTable.At(int(function.FilenameStrindex()))
							symbolizationprocessorProc.logger.Info(
								"Location",
								zap.Int("LocationIdx", sampleLocationIdx),
								zap.Int64("line", line.Line()),
								zap.String("function", functionName),
								zap.String("filename", filename),
							)
						}
						for _, attributeIdx := range location.AttributeIndices().All() {
							symbolizationprocessorProc.logger.Info(
								"Location attribute for "+exe,
								zap.String("key", attributeTable.At(int(attributeIdx)).Key()),
								zap.String("value", attributeTable.At(int(attributeIdx)).Value().AsString()),
							)
						}
					}
					for _, attributeIdx := range sample.AttributeIndices().All() {
						symbolizationprocessorProc.logger.Info(
							"Sample attribute",
							zap.String("key", attributeTable.At(int(attributeIdx)).Key()),
							zap.String("value", attributeTable.At(int(attributeIdx)).Value().AsString()),
						)
					}
				}
			}
		}
	}
	return symbolizationprocessorProc.nextConsumer.ConsumeProfiles(ctx, td)
}

// func addSymbol(stringMap map[string]int, functionName string, stringTable pcommon.StringSlice, filename string, functionMap map[pprofile.Function]int, functionTable pprofile.FunctionSlice, location pprofile.Location, lineNo int32, attributeTable pprofile.AttributeTableSlice) {
// 	nameIdx, ok := stringMap[functionName]
// 	if !ok {
// 		stringTable.Append(functionName)
// 		nameIdx = stringTable.Len() - 1
// 		stringMap[functionName] = nameIdx
// 	}
// 	filenameIdx, ok := stringMap[filename]
// 	if !ok {
// 		stringTable.Append(filename)
// 		filenameIdx = stringTable.Len() - 1
// 		stringMap[filename] = filenameIdx
// 	}
// 	function := pprofile.NewFunction()
// 	function.SetNameStrindex(int32(nameIdx))
// 	function.SetFilenameStrindex(int32(filenameIdx))
// 	functionIdx, ok := functionMap[function]
// 	if !ok {
// 		functionTable.AppendEmpty()
// 		functionIdx = functionTable.Len() - 1
// 		functionMap[function] = functionIdx
// 		function.MoveTo(functionTable.At(int(functionIdx)))
// 	}
// 	line := location.Line().AppendEmpty()
// 	line.SetFunctionIndex(int32(functionIdx))
// 	line.SetLine(int64(lineNo))
// 	for _, attributeIdx := range location.AttributeIndices().All() {
// 		key := attributeTable.At(int(attributeIdx)).Key()
// 		if key == "profile.frame.type" {
// 			attributeTable.At(int(attributeIdx)).Value().SetStr("kernel")
// 		}
// 	}
// }
