package profileutils

import (
	"testing"

	"go.opentelemetry.io/collector/pdata/pprofile"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func TestGetPid_FoundValidPid(t *testing.T) {
	sample := pprofile.NewSample()
	attributeTable := pprofile.NewAttributeTableSlice()
	sample.AttributeIndices().Append(addAttribute(attributeTable, "foo", "bar"))
	sample.AttributeIndices().Append(addAttribute(attributeTable, string(semconv.ProcessPIDKey), "1234"))

	pid := GetPid(sample, attributeTable)
	if pid != 1234 {
		t.Errorf("expected pid 1234, got %d", pid)
	}
}

func TestGetPid_FoundInvalidPid(t *testing.T) {
	sample := pprofile.NewSample()
	attributeTable := pprofile.NewAttributeTableSlice()
	sample.AttributeIndices().Append(addAttribute(attributeTable, "foo", "bar"))
	sample.AttributeIndices().Append(addAttribute(attributeTable, string(semconv.ProcessPIDKey), "notanumber"))

	pid := GetPid(sample, attributeTable)
	if pid != -1 {
		t.Errorf("expected pid -1 for invalid value, got %d", pid)
	}
}

func TestGetPid_NotFound(t *testing.T) {
	sample := pprofile.NewSample()
	attributeTable := pprofile.NewAttributeTableSlice()
	sample.AttributeIndices().Append(addAttribute(attributeTable, "foo", "bar"))

	pid := GetPid(sample, attributeTable)
	if pid != -1 {
		t.Errorf("expected pid -1 when not found, got %d", pid)
	}
}

// addAttribute appends a new attribute with the specified key and value to the given AttributeTableSlice.
// It returns the index of the newly added attribute.
// Parameters:
//   - attributeTable: The slice to which the attribute will be added.
//   - key: The key for the new attribute.
//   - value: The value for the new attribute.
//
// Returns:
//   - int32: The index of the newly added attribute.
func addAttribute(attributeTable pprofile.AttributeTableSlice, key string, value string) int32 {
	attribute := attributeTable.AppendEmpty()
	attribute.SetKey(key)
	attribute.Value().SetStr(value)
	attributeIdx := attributeTable.Len() - 1
	return int32(attributeIdx)
}
