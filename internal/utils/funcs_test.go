package utils

import (
	"testing"
)

func TestParseCIDR(t *testing.T) {
	tests := []struct {
		cidr          string
		expectedIP    string
		expectedMask  string
		expectedError bool
	}{
		{"192.168.1.0/24", "192.168.1.0", "ffffff00", false},
		{"10.0.0.0/8", "10.0.0.0", "ff000000", false},
		{"invalid-cidr", "", "", true},
	}

	for _, tt := range tests {
		ip, mask, err := ParseCIDR(tt.cidr)
		if (err != nil) != tt.expectedError {
			t.Errorf("ParseCIDR(%s) error = %v, expectedError %v", tt.cidr, err, tt.expectedError)
			continue
		}
		if ip != tt.expectedIP {
			t.Errorf("ParseCIDR(%s) ip = %v, expected %v", tt.cidr, ip, tt.expectedIP)
		}
		if mask != tt.expectedMask {
			t.Errorf("ParseCIDR(%s) mask = %v, expected %v", tt.cidr, mask, tt.expectedMask)
		}
	}
}
func TestGetLastIP(t *testing.T) {
	tests := []struct {
		cidr          string
		expectedIP    string
		expectedError bool
	}{
		{"192.168.1.0/24", "192.168.1.254", false},
		{"10.0.0.0/8", "10.255.255.254", false},
		{"invalid-cidr", "", true},
	}

	for _, tt := range tests {
		ip, err := GetLastIP(tt.cidr)
		if (err != nil) != tt.expectedError {
			t.Errorf("GetLastIP(%s) error = %v, expectedError %v", tt.cidr, err, tt.expectedError)
			continue
		}
		if ip != tt.expectedIP {
			t.Errorf("GetLastIP(%s) ip = %v, expected %v", tt.cidr, ip, tt.expectedIP)
		}
	}
}
