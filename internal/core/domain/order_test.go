package domain

import "testing"

func TestOrderTransition(t *testing.T) {
	tests := []struct {
		name  string
		from  OrderStatus
		event OrderEvent
		want  OrderStatus
		ok    bool
	}{
		{"待支付->已支付", OrderPendingPayment, OrderEventPaySuccess, OrderPaid, true},
		{"待支付->用户取消", OrderPendingPayment, OrderEventUserCancel, OrderCanceled, true},
		{"待支付->超时", OrderPendingPayment, OrderEventTimeout, OrderExpired, true},
		{"已支付->完成", OrderPaid, OrderEventComplete, OrderCompleted, true},
		{"已支付->申请退款", OrderPaid, OrderEventApplyRefund, OrderRefunding, true},
		{"退款中->退款成功", OrderRefunding, OrderEventRefundSuccess, OrderRefunded, true},
		{"退款中->退款失败回到已支付", OrderRefunding, OrderEventRefundFail, OrderPaid, true},
		{"已支付->再支付(非法)", OrderPaid, OrderEventPaySuccess, "", false},
		{"已取消->已支付(非法)", OrderCanceled, OrderEventPaySuccess, "", false},
		{"已过期->已支付(非法)", OrderExpired, OrderEventPaySuccess, "", false},
		{"已完成->申请退款(非法)", OrderCompleted, OrderEventApplyRefund, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Order{Status: tt.from}
			err := o.Transition(tt.event)
			if tt.ok && err != nil {
				t.Fatalf("expected ok, got %v", err)
			}
			if !tt.ok && err == nil {
				t.Fatalf("expected error for %s -> %s", tt.from, tt.event)
			}
			if tt.ok && o.Status != tt.want {
				t.Fatalf("expected status %s, got %s", tt.want, o.Status)
			}
		})
	}
}

func TestOrderSettle(t *testing.T) {
	o := &Order{}
	if err := o.Settle(10000, 0, 2000); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if o.PaidCents != 8000 {
		t.Fatalf("expected paid 8000, got %d", o.PaidCents)
	}
	if err := o.Settle(10000, 0, 10001); err == nil {
		t.Fatal("expected ErrMoneyInvalid when coupon > total")
	}
}
