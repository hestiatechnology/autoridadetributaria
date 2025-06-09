package test

import (
	"testing"

	"github.com/hestiatechnology/autoridadetributaria/saft"
	"github.com/hestiatechnology/autoridadetributaria/saft/internal/validation/masterfiles"
)

// Helper functions (can be shared or defined per file if kept separate)
func containsErrorMsgProduct(errs []error, msg string) bool {
	for _, err := range errs {
		if err.Error() == msg {
			return true
		}
	}
	return false
}

func containsErrorMsgPrefixProduct(errs []error, prefix string) bool {
	for _, err := range errs {
		if len(err.Error()) >= len(prefix) && err.Error()[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

func TestValidateProduct_Valid(t *testing.T) {
	product := &saft.Product{
		ProductType:        "P", // Product
		ProductCode:        "PROD001",
		ProductDescription: "Valid Test Product",
		ProductNumberCode:  "EAN1234567890123",
	}
	errs := masterfiles.ValidateProduct(product)
	if len(errs) > 0 {
		t.Errorf("Expected no errors for valid product, got %d: %v", len(errs), errs)
	}
}

func TestValidateProduct_Valid_AllTypes(t *testing.T) {
	validTypes := []string{"P", "S", "O", "I", "E"}
	for _, pType := range validTypes {
		product := &saft.Product{
			ProductType:        pType,
			ProductCode:        "CODE-" + pType,
			ProductDescription: "Desc " + pType,
			ProductNumberCode:  "NUM-" + pType,
		}
		errs := masterfiles.ValidateProduct(product)
		if len(errs) > 0 {
			t.Errorf("Expected no errors for valid product type %s, got %d: %v", pType, len(errs), errs)
		}
	}
}

func TestValidateProduct_Invalid_NilProduct(t *testing.T) {
	errs := masterfiles.ValidateProduct(nil)
	if !containsErrorMsgProduct(errs, "product is nil") {
		t.Errorf("Expected 'product is nil' error, got: %v", errs)
	}
}

func TestValidateProduct_Invalid_ProductType(t *testing.T) {
	product := &saft.Product{
		ProductType:        "X", // Invalid type
		ProductCode:        "PROD002",
		ProductDescription: "Invalid Type Product",
		ProductNumberCode:  "EAN_X",
	}
	errs := masterfiles.ValidateProduct(product)
	expectedMsg := "ProductType 'X' for ProductCode 'PROD002' is invalid. Must be one of P, S, O, I, E"
	if !containsErrorMsgProduct(errs, expectedMsg) {
		t.Errorf("Expected error message '%s', got: %v", expectedMsg, errs)
	}
}

func TestValidateProduct_Invalid_EmptyProductCode(t *testing.T) {
	product := &saft.Product{
		ProductType:        "P",
		ProductCode:        "", // Empty code
		ProductDescription: "No Code Product",
		ProductNumberCode:  "EAN_NOCODE",
	}
	errs := masterfiles.ValidateProduct(product)
	if !containsErrorMsgProduct(errs, "ProductCode is required") {
		t.Errorf("Expected 'ProductCode is required' error, got: %v", errs)
	}
}

func TestValidateProduct_Invalid_EmptyProductDescription(t *testing.T) {
	product := &saft.Product{
		ProductType:        "S",
		ProductCode:        "SERV001",
		ProductDescription: "", // Empty description
		ProductNumberCode:  "EAN_NODESC",
	}
	errs := masterfiles.ValidateProduct(product)
	expectedMsg := "ProductDescription is required for ProductCode 'SERV001'"
	if !containsErrorMsgProduct(errs, expectedMsg) {
		t.Errorf("Expected error message '%s', got: %v", expectedMsg, errs)
	}
}

func TestValidateProduct_Invalid_EmptyProductNumberCode(t *testing.T) {
	product := &saft.Product{
		ProductType:        "O",
		ProductCode:        "OTHER001",
		ProductDescription: "Other Product",
		ProductNumberCode:  "", // Empty number code
	}
	errs := masterfiles.ValidateProduct(product)
	expectedMsg := "ProductNumberCode is required for ProductCode 'OTHER001'"
	if !containsErrorMsgProduct(errs, expectedMsg) {
		t.Errorf("Expected error message '%s', got: %v", expectedMsg, errs)
	}
}

func TestValidateProduct_MultipleErrors(t *testing.T) {
	product := &saft.Product{
		ProductType:        "Z",  // Invalid type
		ProductCode:        "",   // Empty code
		ProductDescription: "",   // Empty description
		ProductNumberCode:  "",   // Empty number code
	}
	errs := masterfiles.ValidateProduct(product)
	if len(errs) != 4 { // Expecting 4 errors
		t.Errorf("Expected 4 errors, got %d: %v", len(errs), errs)
	}
	if !containsErrorMsgPrefixProduct(errs, "ProductType 'Z' for ProductCode '' is invalid") { // ProductCode is empty in msg
		t.Errorf("Missing ProductType error or wrong message format, got: %v", errs)
	}
	if !containsErrorMsgProduct(errs, "ProductCode is required") {
		t.Errorf("Missing ProductCode error, got: %v", errs)
	}
	if !containsErrorMsgPrefixProduct(errs, "ProductDescription is required for ProductCode ''") { // ProductCode is empty
		t.Errorf("Missing ProductDescription error or wrong message format, got: %v", errs)
	}
	if !containsErrorMsgPrefixProduct(errs, "ProductNumberCode is required for ProductCode ''") { // ProductCode is empty
		t.Errorf("Missing ProductNumberCode error or wrong message format, got: %v", errs)
	}
}
