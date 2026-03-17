package btp

import (
	"testing"
)

func TestGetErrorInfos_ValidErrorBlock(t *testing.T) {
	input := `Response mapping: {"error":"ValidationFailed","description":"instance with same name exists for the current tenant"}`

	errorData, err := GetErrorInfos(input)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if errorData.Error != "ValidationFailed" {
		t.Errorf("Expected error 'ValidationFailed', got '%s'", errorData.Error)
	}
}

func TestExtractLastErrorBlock_SingleErrorBlock(t *testing.T) {
	input := `Response mapping: {"error":"TestError","description":"Test description"}`

	errorBlock, err := extractLastErrorBlock(input)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if errorBlock != `{"error":"TestError","description":"Test description"}` {
		t.Errorf("Expected error block to match, got '%s'", errorBlock)
	}
}

func TestExtractLastErrorBlock_MultipleErrorBlocks(t *testing.T) {
	input := `Response mapping: {"error":"FirstError","description":"First"} Response mapping: {"error":"LastError","description":"Last"}`

	errorBlock, err := extractLastErrorBlock(input)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if errorBlock != `{"error":"LastError","description":"Last"}` {
		t.Errorf("Expected last error block, got '%s'", errorBlock)
	}
}

func TestExtractLastErrorBlock_NoErrorBlock(t *testing.T) {
	input := `Response mapping: success without error block`

	_, err := extractLastErrorBlock(input)

	if err == nil {
		t.Error("Expected error when no error block found, got nil")
	}
}

func TestExtractLastErrorBlock_WithWhitespace(t *testing.T) {
	input := `Response mapping: { "error" : "TestError" , "description" : "Test with spaces" }`

	errorBlock, err := extractLastErrorBlock(input)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if errorBlock == "" {
		t.Error("Expected error block to be extracted with whitespace")
	}
}

func TestExtractLastErrorBlock_MultipleResponseMappings(t *testing.T) {
	input := `Response mapping: something Response mapping: {"error":"FinalError","description":"Final"}`

	errorBlock, err := extractLastErrorBlock(input)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if errorBlock != `{"error":"FinalError","description":"Final"}` {
		t.Errorf("Expected final error block after multiple response mappings, got '%s'", errorBlock)
	}
}
