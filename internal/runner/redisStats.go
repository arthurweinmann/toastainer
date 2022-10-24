package runner

import (
	"context"
	"time"

	red "github.com/go-redis/redis/v8"
	"github.com/toastate/toastainer/internal/db/redisdb"
	"github.com/toastate/toastainer/internal/utils"
)

const (
	queueInternalSeconds = 10
	queueMaxPerPipeline  = 12
)

var incrQueue chan string

type DecrReq struct {
	Monthyear  string
	CodeID     string
	UserID     string
	DurationMS int
	RAMGBS     float64 // ram gigabytes-seconds
	CPUS       int     // CPU seconds
	NetIngress float64 // bytes
	NetEgress  float64 // bytes
}

var decrQueue chan *DecrReq

func startRedisQueueRoutine() {
	go func() {
		ticker := time.NewTicker(queueInternalSeconds * time.Second)
		defer ticker.Stop()

		var codeids []string
		var decrs []*DecrReq
		var t string
		var d *DecrReq
		var p red.Pipeliner
		var err error

		for {
			select {
			case t = <-incrQueue:
				codeids = append(codeids, t)
				continue
			case d = <-decrQueue:
				decrs = append(decrs, d)
				continue
			case <-ticker.C:
			}

			if len(codeids) > 0 {
				for c := 0; c < len(codeids)/queueMaxPerPipeline+1; c++ {
					p = redisdb.GetClient().Pipeline()
					for i := c * queueMaxPerPipeline; i < (c+1)*queueMaxPerPipeline && i < len(codeids); i++ {
						p.Incr(context.Background(), "toastercount_"+codeids[i])
					}
					_, err = p.Exec(context.Background())
					if err != nil {
						utils.Error("origin", "redis stats incr routine", "error", err)
					}
				}

				codeids = codeids[:0]
			}

			if len(decrs) > 0 {
				for c := 0; c < len(decrs)/queueMaxPerPipeline+1; c++ {
					p = redisdb.GetClient().Pipeline()
					for i := c * queueMaxPerPipeline; i < (c+1)*queueMaxPerPipeline && i < len(decrs); i++ {
						p.Decr(context.Background(), "toastercount_"+decrs[i].CodeID)

						p.HIncrBy(context.Background(), "toasterstats_"+decrs[i].CodeID+"_"+decrs[i].Monthyear, "executions", 1)
						p.HIncrBy(context.Background(), "toasterstats_"+decrs[i].CodeID+"_"+decrs[i].Monthyear, "durationms", int64(decrs[i].DurationMS))
						p.HIncrBy(context.Background(), "toasterstats_"+decrs[i].CodeID+"_"+decrs[i].Monthyear, "cpus", int64(decrs[i].CPUS))
						p.HIncrByFloat(context.Background(), "toasterstats_"+decrs[i].CodeID+"_"+decrs[i].Monthyear, "ramgbs", decrs[i].RAMGBS)
						p.HIncrByFloat(context.Background(), "toasterstats_"+decrs[i].CodeID+"_"+decrs[i].Monthyear, "ingress", decrs[i].NetIngress)
						p.HIncrByFloat(context.Background(), "toasterstats_"+decrs[i].CodeID+"_"+decrs[i].Monthyear, "egress", decrs[i].NetEgress)

						p.HIncrBy(context.Background(), "userstats_"+decrs[i].UserID+"_"+decrs[i].Monthyear, "executions", 1)
						p.HIncrBy(context.Background(), "userstats_"+decrs[i].UserID+"_"+decrs[i].Monthyear, "durationms", int64(decrs[i].DurationMS))
						p.HIncrBy(context.Background(), "userstats_"+decrs[i].UserID+"_"+decrs[i].Monthyear, "cpus", int64(decrs[i].CPUS))
						p.HIncrByFloat(context.Background(), "userstats_"+decrs[i].UserID+"_"+decrs[i].Monthyear, "ramgbs", decrs[i].RAMGBS)
						p.HIncrByFloat(context.Background(), "userstats_"+decrs[i].UserID+"_"+decrs[i].Monthyear, "ingress", decrs[i].NetIngress)
						p.HIncrByFloat(context.Background(), "userstats_"+decrs[i].UserID+"_"+decrs[i].Monthyear, "egress", decrs[i].NetEgress)
					}
					_, err = p.Exec(context.Background())
					if err != nil {
						utils.Error("origin", "redis stats incr routine", "error", err)
					}
				}

				decrs = decrs[:0]
			}
		}
	}()
}
