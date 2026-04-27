package storage

import (
	"fmt"
	"path/filepath"
	"strings"
)

func ValidateExtension(purpose Purpose, filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	switch purpose {
	case PurposeProductImage, PurposeIdentityCard, PurposeEnterpriseLicense:
		if !allowedExt(ext, ".jpg", ".jpeg", ".png", ".webp") {
			return fmt.Errorf("%s files only support jpg, jpeg, png, webp", purpose)
		}
	case PurposeProductQualification, PurposePaymentVoucher, PurposeReceiptVoucher:
		if !allowedExt(ext, ".jpg", ".jpeg", ".png", ".webp", ".pdf") {
			return fmt.Errorf("%s files only support jpg, jpeg, png, webp, pdf", purpose)
		}
	default:
		return fmt.Errorf("invalid upload purpose: %s", purpose)
	}
	return nil
}
