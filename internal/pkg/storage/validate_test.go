package storage

import "testing"

func TestValidateExtensionAllowsEnterpriseServiceDocuments(t *testing.T) {
	tests := []struct {
		name    string
		purpose Purpose
		file    string
	}{
		{name: "enterprise license pdf", purpose: PurposeEnterpriseLicense, file: "business-license.pdf"},
		{name: "service document pdf", purpose: PurposeEnterpriseServiceDoc, file: "transport-license.pdf"},
		{name: "service document image", purpose: PurposeEnterpriseServiceDoc, file: "carrier-certificate.jpeg"},
		{name: "title voucher pdf", purpose: PurposeTitleVoucher, file: "title-voucher.pdf"},
		{name: "title voucher image", purpose: PurposeTitleVoucher, file: "title-voucher.png"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateExtension(tt.purpose, tt.file); err != nil {
				t.Fatalf("ValidateExtension returned error: %v", err)
			}
		})
	}
}

func TestValidateExtensionRejectsUnsupportedEnterpriseServiceDocument(t *testing.T) {
	if err := ValidateExtension(PurposeEnterpriseServiceDoc, "transport-license.exe"); err == nil {
		t.Fatal("ValidateExtension returned nil error, want unsupported extension error")
	}
}
