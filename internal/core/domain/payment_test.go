package domain

import "testing"

func TestPaymentTransition(t *testing.T) {
	tests := []struct {
		name  string
		from  PaymentStatus
		event PaymentEvent
		want  PaymentStatus
		ok    bool
	}{
		{"PENDING->SUCCESS", PaymentPending, PaymentEventSuccess, PaymentSuccess, true},
		{"PENDING->FAILED", PaymentPending, PaymentEventFail, PaymentFailed, true},
		{"PENDING->CLOSED", PaymentPending, PaymentEventClose, PaymentClosed, true},
		{"SUCCESS->REFUNDED", PaymentSuccess, PaymentEventRefunded, PaymentRefunded, true},
		{"SUCCESS->FAILED 非法", PaymentSuccess, PaymentEventFail, "", false},
		{"PENDING->REFUNDED 非法", PaymentPending, PaymentEventRefunded, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PaymentTransaction{Status: tt.from}
			err := p.Transition(tt.event)
			if tt.ok && err != nil {
				t.Fatalf("expected ok, got %v", err)
			}
			if !tt.ok && err == nil {
				t.Fatalf("expected error for %s -> %s", tt.from, tt.event)
			}
			if tt.ok && p.Status != tt.want {
				t.Fatalf("expected %s, got %s", tt.want, p.Status)
			}
		})
	}
}
