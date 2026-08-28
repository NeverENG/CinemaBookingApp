// Package job 定时任务：只调用 application service 用例，不直接改库。
package job

import (
	"context"
	"log"
	"time"
)

type Job struct {
	Name string
	Run  func(ctx context.Context) error
}

// Runner 任务注册与周期执行。
type Runner struct {
	jobs []Job
}

func NewRunner() *Runner {
	return &Runner{}
}

func (r *Runner) Add(name string, run func(ctx context.Context) error) {
	r.jobs = append(r.jobs, Job{Name: name, Run: run})
}

// RunAll 执行全部任务，单个失败只记日志不阻塞其他任务。
func (r *Runner) RunAll(ctx context.Context) {
	for _, j := range r.jobs {
		if err := j.Run(ctx); err != nil {
			log.Printf("job %s: %v", j.Name, err)
		}
	}
}

// RunPeriodically 周期执行，直到 ctx 取消。
func (r *Runner) RunPeriodically(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.RunAll(ctx)
		}
	}
}
