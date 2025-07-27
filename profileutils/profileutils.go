package profileutils

import (
	"strconv"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pprofile"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// GetPid extracts the process ID (PID) from a given pprofile.Sample using the provided
// pprofile.AttributeTableSlice. It searches for the attribute with the key matching
// semconv.ProcessPIDKey, attempts to parse its value as an integer, and returns the PID.
// If the PID is not found or parsing fails, it returns -1.
func GetPid(sample pprofile.Sample, attributeTable pprofile.AttributeTableSlice) int {
	pid := -1
	for _, attributeIdx := range sample.AttributeIndices().All() {
		key := attributeTable.At(int(attributeIdx)).Key()
		if key == string(semconv.ProcessPIDKey) {
			pidStr := attributeTable.At(int(attributeIdx)).Value().AsString()
			var err error
			pid, err = strconv.Atoi(pidStr)
			if err != nil {
				pid = -1 // Default to -1 if parsing fails
			}
			break
		}
	}
	return pid
}

type symbolAdder struct {
	stringTable pcommon.StringSlice
	functionTable pprofile.FunctionSlice
	attributeTable pprofile.AttributeTableSlice
	stringMap map[string]int
	functionMap map[pprofile.Function]int
}

func NewSymbolAdder(stringTable pcommon.StringSlice, functionTable pprofile.FunctionSlice, attributeTable pprofile.AttributeTableSlice) *symbolAdder {
	adder := &symbolAdder{
		stringTable: stringTable,
		functionTable: functionTable,
		attributeTable: attributeTable,
		stringMap: make(map[string]int),
		functionMap: make(map[pprofile.Function]int),
	}
	for idx, f := range functionTable.All() {
		adder.functionMap[f] = idx
	}

	for idx, s := range stringTable.All() {
		adder.stringMap[s] = idx
	}
	return adder
}

func (s *symbolAdder) AddSymbol(filename string, functionName string, lineNo int32, location pprofile.Location) {
	nameIdx, ok := s.stringMap[functionName]
	if !ok {
		s.stringTable.Append(functionName)
		nameIdx = s.stringTable.Len() - 1
		s.stringMap[functionName] = nameIdx
	}
	filenameIdx, ok := s.stringMap[filename]
	if !ok {
		s.stringTable.Append(filename)
		filenameIdx = s.stringTable.Len() - 1
		s.stringMap[filename] = filenameIdx
	}
	function := pprofile.NewFunction()
	function.SetNameStrindex(int32(nameIdx))
	function.SetFilenameStrindex(int32(filenameIdx))
	functionIdx, ok := s.functionMap[function]
	if !ok {
		s.functionTable.AppendEmpty()
		functionIdx = s.functionTable.Len() - 1
		s.functionMap[function] = functionIdx
		function.MoveTo(s.functionTable.At(int(functionIdx)))
	}
	line := location.Line().AppendEmpty()
	line.SetFunctionIndex(int32(functionIdx))
	line.SetLine(int64(lineNo))
	// For devfiler, we need to set the frame type to something other than "native"
	// because symbols are ignored for native frames:
	// https://github.com/elastic/devfiler/blob/297fe19e9ad0aa7bed93f7ffb97e2e6d09d5ffb2/src/collector/otlp/service.rs#L249
	for _, attributeIdx := range location.AttributeIndices().All() {
		key := s.attributeTable.At(int(attributeIdx)).Key()
		if key == string("profile.frame.type") {
			s.attributeTable.At(int(attributeIdx)).Value().SetStr("kernel")
		}
	}
}
