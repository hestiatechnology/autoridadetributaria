package test

import (
	"testing"
	// "time" // Not strictly needed for these test cases unless parsing dates for setup

	"github.com/hestiatechnology/autoridadetributaria/saft"
	"github.com/hestiatechnology/autoridadetributaria/saft/internal/validation/sourcedocuments"
)

// --- Helper functions (can be shared) ---
func containsErrorMsgMOG(errs []error, msg string) bool {
	for _, err := range errs {
		if err.Error() == msg {
			return true
		}
	}
	return false
}

func containsErrorMsgPrefixMOG(errs []error, prefix string) bool {
	for _, err := range errs {
		if len(err.Error()) >= len(prefix) && err.Error()[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// --- Mock data for dependencies (reused conceptually) ---
var (
	mockProductCodesMOG = map[string]bool{"PROD001": true, "PROD002": true}
	mockCustomerIDsMOG  = map[string]bool{"CUST001": true}
	mockSupplierIDsMOG  = map[string]bool{"SUP001": true}
)

func sampleValidStockMovement(docNum string, custID string, prodCode string, movementType string) saft.StockMovement {
	sm := saft.StockMovement{
		DocumentNumber: docNum,
		ATCUD:          "ATCUDMOG123",
		DocumentStatus: &saft.DocumentStatus{ // Assuming pointer
			MovementStatus:     "N", // Normal
			MovementStatusDate: "2023-11-01T11:00:00",
			SourceID:           "UserMOG",
			SourceBilling:      "P",
		},
		Hash:            "MOGHASH1234567890=",
		HashControl:     "1",
		Period:          "11",
		MovementDate:    "2023-11-01",
		MovementType:    movementType,
		SourceID:        "UserMOG_DocLevel",
		SystemEntryDate: "2023-11-01T10:00:00",
		CustomerID:      custID,    // Can be empty if SupplierID is used or neither for internal
		SupplierID:      "",        // Can be empty
		MovementStartTime: "2023-11-01T10:30:00",
		ATDocCodeID:     "DOCCODE123", // Optional, but can be present
		Line: []saft.Line{
			{
				LineNumber:         "1",
				ProductCode:        prodCode,
				ProductDescription: "Product for Movement",
				Quantity:           "10.00",
				UnitOfMeasure:      "KG",
				UnitPrice:          "2.50", // Can be 0 for non-valued movements
				TaxPointDate:       "2023-11-01",
				Description:        "Movement line item",
				Tax: &saft.Tax{ // Assuming pointer; Tax on movements can be complex (e.g. IVA for self-consumption)
					TaxType:          "IVA",
					TaxCountryRegion: "PT",
					TaxCode:          "NOR",
					TaxPercentage:    "23.00", // Required if TaxType is IVA and Tax block present
				},
			},
		},
		DocumentTotals: &saft.DocumentTotals{ // Assuming pointer
			TaxPayable: "5.75",  // 10 * 2.50 * 0.23 (if valued and taxed)
			NetTotal:   "25.00", // 10 * 2.50
			GrossTotal: "30.75", // 25 + 5.75
		},
	}
	if movementType == "GR" { // Example: Goods Receipt might involve a supplier
		sm.SupplierID = "SUP001"
		sm.CustomerID = "" // Typically not both
	}
	return sm
}

func TestValidateMovementOfGoods_Valid(t *testing.T) {
	mog := &saft.MovementOfGoods{
		NumberOfMovementLines: "1",
		TotalQuantityIssued:   "10.00", // Sum of all Line.Quantity
		StockMovement: []saft.StockMovement{
			sampleValidStockMovement("GR/1", "", "PROD001", "GR"), // Supplier set in sample func
		},
	}
	errs := sourcedocuments.ValidateMovementOfGoods(mog, mockProductCodesMOG, mockCustomerIDsMOG, mockSupplierIDsMOG)
	if len(errs) > 0 {
		t.Errorf("Expected no errors for valid MovementOfGoods, got %d: %v", len(errs), errs)
	}
}

func TestValidateMovementOfGoods_Invalid_NumberOfLinesMismatch(t *testing.T) {
	mog := &saft.MovementOfGoods{
		NumberOfMovementLines: "2", // Mismatch, only 1 line in sample
		TotalQuantityIssued:   "10.00",
		StockMovement: []saft.StockMovement{
			sampleValidStockMovement("GT/1", "CUST001", "PROD001", "GT"),
		},
	}
	errs := sourcedocuments.ValidateMovementOfGoods(mog, mockProductCodesMOG, mockCustomerIDsMOG, mockSupplierIDsMOG)
	if !containsErrorMsgPrefixMOG(errs, "mismatch in NumberOfMovementLines: declared 2, actual 1") {
		t.Errorf("Expected NumberOfMovementLines mismatch error, got: %v", errs)
	}
}

func TestValidateMovementOfGoods_Invalid_TotalQuantityMismatch(t *testing.T) {
	mog := &saft.MovementOfGoods{
		NumberOfMovementLines: "1",
		TotalQuantityIssued:   "5.00", // Mismatch, actual is 10.00 in sample
		StockMovement: []saft.StockMovement{
			sampleValidStockMovement("GT/1", "CUST001", "PROD001", "GT"),
		},
	}
	errs := sourcedocuments.ValidateMovementOfGoods(mog, mockProductCodesMOG, mockCustomerIDsMOG, mockSupplierIDsMOG)
	if !containsErrorMsgPrefixMOG(errs, "mismatch in TotalQuantityIssued: declared 5.000000, calculated 10.000000") {
		t.Errorf("Expected TotalQuantityIssued mismatch error, got: %v", errs)
	}
}

func TestValidateMovementOfGoods_StockMovement_Invalid_EmptyDocNumber(t *testing.T) {
	sm := sampleValidStockMovement("", "CUST001", "PROD001", "GT") // Empty DocumentNumber
	mog := &saft.MovementOfGoods{
		NumberOfMovementLines: "1", TotalQuantityIssued: "10.00",
		StockMovement: []saft.StockMovement{sm},
	}
	errs := sourcedocuments.ValidateMovementOfGoods(mog, mockProductCodesMOG, mockCustomerIDsMOG, mockSupplierIDsMOG)
	if !containsErrorMsgPrefixMOG(errs, "StockMovement[0 DocNum:]: DocumentNumber is required") {
		t.Errorf("Expected DocumentNumber required error, got: %v", errs)
	}
}

func TestValidateMovementOfGoods_StockMovement_Invalid_MovementStatus(t *testing.T) {
	sm := sampleValidStockMovement("GT/1", "CUST001", "PROD001", "GT")
	sm.DocumentStatus.MovementStatus = "X" // Invalid status
	mog := &saft.MovementOfGoods{
		NumberOfMovementLines: "1", TotalQuantityIssued: "10.00",
		StockMovement: []saft.StockMovement{sm},
	}
	errs := sourcedocuments.ValidateMovementOfGoods(mog, mockProductCodesMOG, mockCustomerIDsMOG, mockSupplierIDsMOG)
	expectedMsg := "StockMovement[0 DocNum:GT/1]: DocumentStatus.MovementStatus 'X' is invalid"
	if !containsErrorMsgMOG(errs, expectedMsg) {
		t.Errorf("Expected '%s', got: %v", expectedMsg, errs)
	}
}

func TestValidateMovementOfGoods_StockMovement_Invalid_MovementType(t *testing.T) {
	sm := sampleValidStockMovement("GT/1", "CUST001", "PROD001", "XX") // Invalid MovementType
	mog := &saft.MovementOfGoods{
		NumberOfMovementLines: "1", TotalQuantityIssued: "10.00",
		StockMovement: []saft.StockMovement{sm},
	}
	errs := sourcedocuments.ValidateMovementOfGoods(mog, mockProductCodesMOG, mockCustomerIDsMOG, mockSupplierIDsMOG)
	expectedMsg := "StockMovement[0 DocNum:GT/1]: MovementType 'XX' is invalid"
	if !containsErrorMsgMOG(errs, expectedMsg) {
		t.Errorf("Expected '%s', got: %v", expectedMsg, errs)
	}
}

func TestValidateMovementOfGoods_Line_Invalid_ProductCode_NotFound(t *testing.T) {
	sm := sampleValidStockMovement("GT/1", "CUST001", "PROD_INVALID", "GT") // Invalid ProductCode
	mog := &saft.MovementOfGoods{
		NumberOfMovementLines: "1", TotalQuantityIssued: "10.00",
		StockMovement: []saft.StockMovement{sm},
	}
	errs := sourcedocuments.ValidateMovementOfGoods(mog, mockProductCodesMOG, mockCustomerIDsMOG, mockSupplierIDsMOG)
	expectedMsg := "StockMovement[0 DocNum:GT/1] Line[1]: ProductCode 'PROD_INVALID' is invalid or not found"
	if !containsErrorMsgMOG(errs, expectedMsg) {
		t.Errorf("Expected '%s', got: %v", expectedMsg, errs)
	}
}

func TestValidateMovementOfGoods_Line_Invalid_Tax_IVAMissingPercentage(t *testing.T) {
	sm := sampleValidStockMovement("GT/1", "CUST001", "PROD001", "GT")
	if sm.Line[0].Tax != nil {
		sm.Line[0].Tax.TaxType = "IVA"
		sm.Line[0].Tax.TaxPercentage = "" // IVA type but percentage is empty
	}
	mog := &saft.MovementOfGoods{
		NumberOfMovementLines: "1", TotalQuantityIssued: "10.00",
		StockMovement: []saft.StockMovement{sm},
	}
	// Adjust DocumentTotals as tax might change
	sm.DocumentTotals.TaxPayable = "0.00"
	sm.DocumentTotals.GrossTotal = sm.DocumentTotals.NetTotal

	errs := sourcedocuments.ValidateMovementOfGoods(mog, mockProductCodesMOG, mockCustomerIDsMOG, mockSupplierIDsMOG)
	expectedMsg := "StockMovement[0 DocNum:GT/1] Line[1]: Tax.TaxPercentage is required for TaxType IVA"
	if !containsErrorMsgMOG(errs, expectedMsg) {
		t.Errorf("Expected '%s', got: %v", expectedMsg, errs)
	}
}

func TestValidateMovementOfGoods_StockMovement_Invalid_CustomerID_NotFound(t *testing.T) {
	sm := sampleValidStockMovement("GT/1", "CUST_INVALID", "PROD001", "GT") // Invalid CustomerID
	mog := &saft.MovementOfGoods{
		NumberOfMovementLines: "1", TotalQuantityIssued: "10.00",
		StockMovement: []saft.StockMovement{sm},
	}
	errs := sourcedocuments.ValidateMovementOfGoods(mog, mockProductCodesMOG, mockCustomerIDsMOG, mockSupplierIDsMOG)
	expectedMsg := "StockMovement[0 DocNum:GT/1]: CustomerID 'CUST_INVALID' is invalid or not found"
	if !containsErrorMsgMOG(errs, expectedMsg) {
		t.Errorf("Expected '%s', got: %v", expectedMsg, errs)
	}
}

func TestValidateMovementOfGoods_StockMovement_Invalid_SupplierID_NotFound(t *testing.T) {
	sm := sampleValidStockMovement("GR/1", "", "PROD001", "GR") // GR type, sample func sets SupplierID
	sm.SupplierID = "SUP_INVALID" // Set to an invalid one
	sm.CustomerID = "" // Ensure CustomerID is not set for this test
	mog := &saft.MovementOfGoods{
		NumberOfMovementLines: "1", TotalQuantityIssued: "10.00",
		StockMovement: []saft.StockMovement{sm},
	}
	errs := sourcedocuments.ValidateMovementOfGoods(mog, mockProductCodesMOG, mockCustomerIDsMOG, mockSupplierIDsMOG)
	expectedMsg := "StockMovement[0 DocNum:GR/1]: SupplierID 'SUP_INVALID' is invalid or not found"
	if !containsErrorMsgMOG(errs, expectedMsg) {
		t.Errorf("Expected '%s', got: %v", expectedMsg, errs)
	}
}


// Further tests could include:
// - Nil MovementOfGoods input.
// - More permutations of DocumentStatus.
// - Empty or invalid fields in StockMovement header (ATCUD, Hash, Dates, Period, etc.).
// - ATDocCodeID validation (if specific rules apply beyond "not empty if provided").
// - Line: Empty/invalid Quantity, UnitOfMeasure, UnitPrice, Description.
// - Tax (Line level): TaxType, TaxCountryRegion, TaxCode (if Tax block exists). Invalid TaxPercentage format.
// - DocumentTotals: Invalid formats for TaxPayable, NetTotal, GrossTotal. Currency validation.
// - Multiple StockMovement documents, some valid, some invalid.
// - Multiple lines, some valid, some invalid.
// - Cases where DocumentStatus, Tax, DocumentTotals are nil pointers.
// - Check "if Tax is provided" logic for lines - if line.Tax is nil, no tax validation errors should occur.
// - Check "ATDocCodeID (if provided) is not empty" - currently if ATDocCodeID="", no error. If ATDocCodeID=" ", it's not empty.
// - Check "CustomerID or SupplierID is provided if applicable" - current validator checks them if present, doesn't enforce presence.
//   This is usually fine as "applicability" depends on MovementType or other context.
