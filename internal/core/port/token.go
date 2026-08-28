package port

// TokenManager 登录用例需要的令牌签发能力。
// 实现由 pkg/jwt 提供；请求时的解析属于 transport 层职责，不在这里定义。
type TokenManager interface {
	Generate(userID int64, role string) (string, error)
}
