package domain

type RefundStatus string

const (
	RefundPending RefundStatus = "PENDING"
	RefundSuccess RefundStatus = "SUCCESS"
	RefundFailed  RefundStatus = "FAILED"
)

// Refund 退款单：一单一退。
type Refund struct {
	RefundNo         string
	OrderNo          string
	UserID           int64
	AmountCents      int64
	Reason           string
	Status           RefundStatus
	ExternalRefundNo string
}
