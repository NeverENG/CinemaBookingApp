package port

import "context"

/*
Server 层启动事务
*/
type TXManager interface {
	Run(ctx context.Context, fn func(ctx context.Context) error) error
}
