package btp

import (
	"testing"
)

func TestGetErrorInfos_ValidErrorBlock(t *testing.T) {
	input := `Response mapping: {"error":"ValidationFailed","description":"instance with same name exists for the current tenant"}`

	errorData, errorCode, err := GetErrorInfos(input)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if errorData.Error != "ValidationFailed" {
		t.Errorf("Expected error 'ValidationFailed', got '%s'", errorData.Error)
	}
	if errorCode != "INSTANCE_ALREADY_EXISTS" {
		t.Errorf("Expected error code 'INSTANCE_ALREADY_EXISTS', got '%s'", errorCode)
	}
}

func TestGetErrorInfos_MultipleBindingsError(t *testing.T) {
	input := `Response mapping: {"error":"BadRequest","description":"found multiple service bindings with the name test-binding"}`

	_, errorCode, err := GetErrorInfos(input)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if errorCode != "MULTIPLE_BINDINGS_FOUND" {
		t.Errorf("Expected error code 'MULTIPLE_BINDINGS_FOUND', got '%s'", errorCode)
	}
}

func TestGetErrorInfos_BindingAlreadyExistsError(t *testing.T) {
	input := `Response mapping: {"error":"BadRequest","description":"binding with same name exists for instance my-instance"}`

	_, errorCode, err := GetErrorInfos(input)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if errorCode != "BINDING_ALREADY_EXISTS" {
		t.Errorf("Expected error code 'BINDING_ALREADY_EXISTS', got '%s'", errorCode)
	}
}

func TestGetErrorInfos_ServiceInstanceNotFoundError(t *testing.T) {
	input := `Response mapping: {"error":"NotFound","description":"could not find such service instance"}`

	_, errorCode, err := GetErrorInfos(input)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if errorCode != "SERVICE_INSTANCE_NOT_FOUND" {
		t.Errorf("Expected error code 'SERVICE_INSTANCE_NOT_FOUND', got '%s'", errorCode)
	}
}

func TestGetErrorInfos_InstanceNotFoundError(t *testing.T) {
	input := `Response mapping: {"error":"NotFound","description":"could not find such instance"}`

	_, errorCode, err := GetErrorInfos(input)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if errorCode != "SERVICE_INSTANCE_NOT_FOUND" {
		t.Errorf("Expected error code 'SERVICE_INSTANCE_NOT_FOUND', got '%s'", errorCode)
	}
}

func TestGetErrorInfos_ServiceBindingNotFoundError(t *testing.T) {
	input := `Response mapping: {"error":"NotFound","description":"could not find such service binding"}`

	_, errorCode, err := GetErrorInfos(input)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if errorCode != "SERVICE_BINDING_NOT_FOUND" {
		t.Errorf("Expected error code 'SERVICE_BINDING_NOT_FOUND', got '%s'", errorCode)
	}
}

func TestGetErrorInfos_BindingNotFoundError(t *testing.T) {
	input := `Response mapping: {"error":"NotFound","description":"could not find such binding"}`

	_, errorCode, err := GetErrorInfos(input)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if errorCode != "SERVICE_BINDING_NOT_FOUND" {
		t.Errorf("Expected error code 'SERVICE_BINDING_NOT_FOUND', got '%s'", errorCode)
	}
}

func TestGetErrorInfos_UnknownError(t *testing.T) {
	input := `Response mapping: {"error":"InternalServerError","description":"An unknown error occurred"}`

	_, errorCode, err := GetErrorInfos(input)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if errorCode != "" {
		t.Errorf("Expected empty error code for unknown error, got '%s'", errorCode)
	}
}

func TestGetErrorInfos_NoErrorBlock(t *testing.T) {
	input := `Response mapping: success`

	_, _, err := GetErrorInfos(input)

	if err == nil {
		t.Error("Expected error for missing error block, got nil")
	}
}

func TestGetErrorInfos_MultipleErrorBlocks(t *testing.T) {
	input := `Response mapping: {"error":"Error1","description":"First error"} Response mapping: {"error":"Error2","description":"instance with same name exists for the current tenant"}`

	_, errorCode, err := GetErrorInfos(input)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	// Should extract the last error block
	if errorCode != "INSTANCE_ALREADY_EXISTS" {
		t.Errorf("Expected error code 'INSTANCE_ALREADY_EXISTS' from last block, got '%s'", errorCode)
	}
}

func TestGetErrorInfos_MalformedJSON(t *testing.T) {
	input := `Response mapping: {"error":"BadRequest","description":"this is not valid json but has error block format}`

	_, _, err := GetErrorInfos(input)

	// Should return an error when JSON parsing fails
	if err == nil {
		t.Error("Expected error for malformed JSON, got nil")
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

func TestMapErrorMessageToCode_MultipleBindings(t *testing.T) {
	tests := []struct {
		message  string
		expected string
	}{
		{"found multiple service bindings with the name test", "MULTIPLE_BINDINGS_FOUND"},
		{"Found Multiple Service Bindings With The Name test", "MULTIPLE_BINDINGS_FOUND"},
		{"FOUND MULTIPLE SERVICE BINDINGS WITH THE NAME test", "MULTIPLE_BINDINGS_FOUND"},
	}

	for _, test := range tests {
		result := mapErrorMessageToCode(test.message)
		if result != test.expected {
			t.Errorf("For message '%s': expected '%s', got '%s'", test.message, test.expected, result)
		}
	}
}

func TestMapErrorMessageToCode_BindingExists(t *testing.T) {
	tests := []struct {
		message  string
		expected string
	}{
		{"binding with same name exists for instance test", "BINDING_ALREADY_EXISTS"},
		{"Binding With Same Name Exists For Instance test", "BINDING_ALREADY_EXISTS"},
		{"BINDING WITH SAME NAME EXISTS FOR INSTANCE test", "BINDING_ALREADY_EXISTS"},
	}

	for _, test := range tests {
		result := mapErrorMessageToCode(test.message)
		if result != test.expected {
			t.Errorf("For message '%s': expected '%s', got '%s'", test.message, test.expected, result)
		}
	}
}

func TestMapErrorMessageToCode_InstanceNotFound(t *testing.T) {
	tests := []struct {
		message  string
		expected string
	}{
		{"could not find such instance", "SERVICE_INSTANCE_NOT_FOUND"},
		{"could not find such service instance", "SERVICE_INSTANCE_NOT_FOUND"},
		{"Could Not Find Such Instance", "SERVICE_INSTANCE_NOT_FOUND"},
		{"COULD NOT FIND SUCH SERVICE INSTANCE", "SERVICE_INSTANCE_NOT_FOUND"},
	}

	for _, test := range tests {
		result := mapErrorMessageToCode(test.message)
		if result != test.expected {
			t.Errorf("For message '%s': expected '%s', got '%s'", test.message, test.expected, result)
		}
	}
}

func TestMapErrorMessageToCode_BindingNotFound(t *testing.T) {
	tests := []struct {
		message  string
		expected string
	}{
		{"could not find such binding", "SERVICE_BINDING_NOT_FOUND"},
		{"could not find such service binding", "SERVICE_BINDING_NOT_FOUND"},
		{"Could Not Find Such Binding", "SERVICE_BINDING_NOT_FOUND"},
		{"COULD NOT FIND SUCH SERVICE BINDING", "SERVICE_BINDING_NOT_FOUND"},
	}

	for _, test := range tests {
		result := mapErrorMessageToCode(test.message)
		if result != test.expected {
			t.Errorf("For message '%s': expected '%s', got '%s'", test.message, test.expected, result)
		}
	}
}

func TestMapErrorMessageToCode_InstanceExists(t *testing.T) {
	tests := []struct {
		message  string
		expected string
	}{
		{"instance with same name exists for the current tenant", "INSTANCE_ALREADY_EXISTS"},
		{"Instance With Same Name Exists For The Current Tenant", "INSTANCE_ALREADY_EXISTS"},
		{"INSTANCE WITH SAME NAME EXISTS FOR THE CURRENT TENANT", "INSTANCE_ALREADY_EXISTS"},
	}

	for _, test := range tests {
		result := mapErrorMessageToCode(test.message)
		if result != test.expected {
			t.Errorf("For message '%s': expected '%s', got '%s'", test.message, test.expected, result)
		}
	}
}

func TestMapErrorMessageToCode_UnknownMessage(t *testing.T) {
	message := "This is an unknown error message"
	result := mapErrorMessageToCode(message)

	if result != "" {
		t.Errorf("Expected empty string for unknown message, got '%s'", result)
	}
}

func TestMapErrorMessageToCode_EmptyMessage(t *testing.T) {
	result := mapErrorMessageToCode("")

	if result != "" {
		t.Errorf("Expected empty string for empty message, got '%s'", result)
	}
}

func TestMapErrorMessageToCode_Priority(t *testing.T) {
	// Test that first matching regex is used (priority matters)
	message := "found multiple service bindings with the name test and instance with same name exists"
	result := mapErrorMessageToCode(message)

	// Should match the first regex (multipleBindingsRegex)
	if result != "MULTIPLE_BINDINGS_FOUND" {
		t.Errorf("Expected 'MULTIPLE_BINDINGS_FOUND' based on regex priority, got '%s'", result)
	}
}
