package domain

import "time"

type PaymentStatus string

const (
	PaymentPending  PaymentStatus = "PENDING"
	PaymentSuccess  PaymentStatus = "SUCCESS"
	PaymentFailed   PaymentStatus = "FAILED"
	PaymentClosed   PaymentStatus = "CLOSED"
	PaymentRefunded PaymentStatus = "REFUNDED"
)

type PaymentEvent string

const (
	PaymentEventSuccess  PaymentEvent = "SUCCESS"
	PaymentEventFail     PaymentEvent = "FAIL"
	PaymentEventClose    PaymentEvent = "CLOSE"
	PaymentEventRefunded PaymentEvent = "REFUNDED"
)

var paymentTransitions = map[PaymentStatus]map[PaymentEvent]PaymentStatus{
	PaymentPending: {
		PaymentEventSuccess: PaymentSuccess,
		PaymentEventFail:    PaymentFailed,
		PaymentEventClose:   PaymentClosed,
	},
	PaymentSuccess: {
		PaymentEventRefunded: PaymentRefunded,
	},
}

func (s PaymentStatus) CanTransition(event PaymentEvent) bool {
	_, ok := paymentTransitions[s][event]
	return ok
}

// PaymentTransaction 支付交易：一笔订单一次支付。
// 下单阶段不创建，支付阶段才创建（PENDING），回调成功后 SUCCESS。
type PaymentTransaction struct {
	TransactionNo   string
	OrderNo         string
	UserID          int64
	AmountCents     int64
	Channel         string
	Status          PaymentStatus
	ExternalTradeNo string
	Version         int32
	CreatedAt       time.Time
	PaidAt          *time.Time
	ClosedAt        *time.Time
}

func (p *PaymentTransaction) Transition(event PaymentEvent) error {
	next, ok := paymentTransitions[p.Status][event]
	if !ok {
		return ErrInvalidTransition
	}
	p.Status = next
	return nil
}

type CallbackStatus string

const (
	CallbackReceived  CallbackStatus = "RECEIVED"
	CallbackProcessed CallbackStatus = "PROCESSED"
	CallbackDuplicate CallbackStatus = "DUPLICATE"
	CallbackFailed    CallbackStatus = "FAILED"
)

// PaymentCallback 支付回调记录：只增不改，event_id 是幂等键。
type PaymentCallback struct {
	EventID       string
	TransactionNo string
	AmountCents   int64
	Payload       string
	Status        CallbackStatus
	RetryCount    int
	CreatedAt     time.Time
	ProcessedAt   *time.Time
}
