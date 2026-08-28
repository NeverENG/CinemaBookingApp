package uid

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// OrderNo 生成业务订单号：O + 时间戳 + 随机后缀。
func OrderNo() string {
	return "O" + time.Now().Format("20060102150405") + randHex(4)
}

// LockToken 生成座位锁令牌。
func LockToken() string {
	return randHex(16)
}

// TransactionNo 生成支付交易号。
func TransactionNo() string {
	return "T" + time.Now().Format("20060102150405") + randHex(4)
}

// TicketNo 生成取票码。
func TicketNo() string {
	return "TK" + randHex(6)
}

// RefundNo 生成退款单号。
func RefundNo() string {
	return "RF" + time.Now().Format("20060102150405") + randHex(4)
}

func randHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err) // crypto/rand 失败属于环境级故障
	}
	return hex.EncodeToString(b)
}
