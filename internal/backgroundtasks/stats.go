package backgroundtasks

import (
	"context"
	"strconv"
	"time"

	"github.com/toastate/toastainer/internal/db/objectdb"
	"github.com/toastate/toastainer/internal/db/redisdb"
	"github.com/toastate/toastainer/internal/model"
	"github.com/toastate/toastainer/internal/utils"
)

func statsRoutine() {
	var lock *taskLock
	var err error

F:
	for {
		if lock != nil {
			lock.release()
			lock = nil
		}
		// time.Sleep(30 * time.Minute)
		time.Sleep(1 * time.Minute)

		lock, err = refreshTaskLock("statistics", 30*time.Minute)
		if err != nil {
			utils.Error("origin", "statistics fetch background routine", "error", err)
			continue F
		}

		month := strconv.Itoa(int(time.Now().Month()))
		year := strconv.Itoa(time.Now().Year())

		var users []model.User
		next := true
		var cursor string
		for next {
			cursor, next, users, err = objectdb.Client.RangeUsers(100, cursor)
			if err != nil {
				lock.release()
				lock = nil
				utils.Error("origin", "statistics fetch background routine", "error", err)
				continue F
			}

			if len(users) > 0 {
				for i := 0; i < len(users); i++ {
					stats, err := redisdb.GetClient().HGetAll(context.Background(), "userstats_"+users[i].ID+"_"+month+year).Result()
					if err != nil && err != redisdb.ErrNil {
						lock.release()
						lock = nil
						cursor = ""
						utils.Error("origin", "statistics fetch background routine", "error", "error", err)
						continue F
					}
					if err == nil {
						var dms, cpus, executions int64
						var ramgbs, ingress, egress float64
						if _, ok := stats["durationms"]; ok {
							dms, err = strconv.ParseInt(stats["durationms"], 10, 64)
							if err != nil {
								utils.Error("origin", "statistics fetch background routine", "error", "error", err)
								continue
							}
						}

						if _, ok := stats["cpus"]; ok {
							cpus, err = strconv.ParseInt(stats["cpus"], 10, 64)
							if err != nil {
								utils.Error("origin", "statistics fetch background routine", "error", "error", err)
								continue
							}
						}

						if _, ok := stats["executions"]; ok {
							executions, err = strconv.ParseInt(stats["executions"], 10, 64)
							if err != nil {
								utils.Error("origin", "statistics fetch background routine", "error", "error", err)
								continue
							}
						}

						if _, ok := stats["ramgbs"]; ok {
							ramgbs, err = strconv.ParseFloat(stats["ramgbs"], 64)
							if err != nil {
								utils.Error("origin", "statistics fetch background routine", "error", "error", err)
								continue
							}
						}

						if _, ok := stats["ingress"]; ok {
							ingress, err = strconv.ParseFloat(stats["ingress"], 64)
							if err != nil {
								utils.Error("origin", "statistics fetch background routine", "error", "error", err)
								continue
							}
						}

						if _, ok := stats["egress"]; ok {
							egress, err = strconv.ParseFloat(stats["egress"], 64)
							if err != nil {
								utils.Error("origin", "statistics fetch background routine", "error", "error", err)
								continue
							}
						}

						usrstat := &model.UserStatistics{
							UserID:     users[i].ID,
							Monthyear:  month + year,
							DurationMS: int(dms),
							CPUS:       int(cpus),
							Executions: int(executions),
							RAMGBS:     ramgbs,
							NetIngress: ingress,
							NetEgress:  egress,
						}

						// Otherwise we may get an SQL out of range value error
						if usrstat.RAMGBS < 0.00000001 {
							usrstat.RAMGBS = 0
						}
						if usrstat.NetIngress < 0.00000001 {
							usrstat.NetIngress = 0
						}
						if usrstat.NetEgress < 0.00000001 {
							usrstat.NetEgress = 0
						}

						if usrstat.Executions > 0 || usrstat.DurationMS > 0 || usrstat.CPUS > 0 || usrstat.RAMGBS > 0 || usrstat.NetIngress > 0 || usrstat.NetEgress > 0 {
							p := redisdb.GetClient().Pipeline()
							p.HIncrBy(context.Background(), "userstats_"+users[i].ID+"_"+month+year, "executions", -int64(usrstat.Executions))
							p.HIncrBy(context.Background(), "userstats_"+users[i].ID+"_"+month+year, "durationms", -int64(usrstat.DurationMS))
							p.HIncrBy(context.Background(), "userstats_"+users[i].ID+"_"+month+year, "cpus", -int64(usrstat.CPUS))
							p.HIncrByFloat(context.Background(), "userstats_"+users[i].ID+"_"+month+year, "ramgbs", -usrstat.RAMGBS)
							p.HIncrByFloat(context.Background(), "userstats_"+users[i].ID+"_"+month+year, "ingress", -usrstat.NetIngress)
							p.HIncrByFloat(context.Background(), "userstats_"+users[i].ID+"_"+month+year, "egress", -usrstat.NetEgress)
							_, err = p.Exec(context.Background())
							if err != nil {
								lock.release()
								lock = nil
								cursor = ""
								utils.Error("origin", "statistics fetch background routine", "error", err)
								continue F
							}

							err = objectdb.Client.IncrUserStatistics(usrstat)
							if err != nil {
								lock.release()
								lock = nil
								cursor = ""
								utils.Error("origin", "statistics fetch background routine", "error", err)
								continue F
							}
						}
					}
				}
			}

			if !next {
				continue F
			} else {
				if !lock.active() {
					lock, err = refreshTaskLock("statistics", 30*time.Minute)
					if err != nil {
						utils.Error("origin", "certificates background routine", "error", "error", err)
						continue F
					}
					if lock == nil {
						continue F
					}
				}
			}
		}
	}
}
