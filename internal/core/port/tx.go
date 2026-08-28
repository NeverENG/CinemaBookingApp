package port

import "context"

// TxManager 由 service 层启动事务：一个用例一个事务。
// 实现方（infra）负责把事务放入 ctx，repo 只接收 ctx、不自己开事务。
type TxManager interface {
	Run(ctx context.Context, fn func(ctx context.Context) error) error
}
