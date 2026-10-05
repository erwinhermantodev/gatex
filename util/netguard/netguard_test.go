package netguard

import "testing"

func TestCheckHost(t *testing.T) {
	AllowLoopback = false
	for _, h := range []string{"169.254.169.254", "0.0.0.0", "127.0.0.1", "::1", "localhost", "metadata.google.internal", ""} {
		if CheckHost(h) == nil {
			t.Errorf("%q should be blocked", h)
		}
	}
	for _, h := range []string{"10.1.2.3", "192.168.0.5", "93.184.216.34"} {
		if err := CheckHost(h); err != nil {
			t.Errorf("%q should be allowed: %v", h, err)
		}
	}
	AllowLoopback = true
	defer func() { AllowLoopback = false }()
	if err := CheckHost("127.0.0.1"); err != nil {
		t.Errorf("loopback should be allowed when enabled: %v", err)
	}
	if CheckHost("169.254.169.254") == nil {
		t.Error("metadata address must stay blocked even with AllowLoopback")
	}
}
