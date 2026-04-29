package service

import "testing"

func TestFixedDevSmsCode(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		code     string
		want     string
	}{
		{name: "default provider accepts fixed dev code", code: "123456", want: "123456"},
		{name: "dev provider accepts fixed dev code", provider: "dev", code: "654321", want: "654321"},
		{name: "tencent provider ignores fixed dev code", provider: "tencent", code: "123456", want: ""},
		{name: "rejects short code", code: "12345", want: ""},
		{name: "rejects non-digit code", code: "12a456", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SMS_PROVIDER", tt.provider)
			t.Setenv("SMS_DEV_CODE", tt.code)

			if got := fixedDevSmsCode(); got != tt.want {
				t.Fatalf("fixedDevSmsCode() = %q, want %q", got, tt.want)
			}
		})
	}
}
