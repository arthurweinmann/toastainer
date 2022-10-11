package backgroundtasks

import (
	"context"
	"strconv"
	"time"

	"github.com/toastate/toastcloud/internal/db/objectdb"
	"github.com/toastate/toastcloud/internal/db/redisdb"
	"github.com/toastate/toastcloud/internal/model"
	"github.com/toastate/toastcloud/internal/utils"
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
		time.Sleep(30 * time.Minute)

		lock, err = refreshTaskLock("statistics", 30*time.Minute)
		if err != nil {
			utils.Error("msg", "statistics fetch background routine", err)
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
				utils.Error("msg", "statistics fetch background routine", err)
				continue F
			}

			if len(users) > 0 {
				for i := 0; i < len(users); i++ {
					stats, err := redisdb.GetClient().HGetAll(context.Background(), "userstats_"+users[i].ID+"_"+month+year).Result()
					if err != nil && err != redisdb.ErrNil {
						lock.release()
						lock = nil
						cursor = ""
						utils.Error("msg", "statistics fetch background routine", "error", err)
						continue F
					}
					if err == nil {
						var dms, cpus, executions int64
						var ramgbs, ingress, egress float64
						if _, ok := stats["durationms"]; ok {
							dms, err = strconv.ParseInt(stats["durationms"], 10, 64)
							if err != nil {
								utils.Error("msg", "statistics fetch background routine", "error", err)
								continue
							}
						}

						if _, ok := stats["cpus"]; ok {
							cpus, err = strconv.ParseInt(stats["cpus"], 10, 64)
							if err != nil {
								utils.Error("msg", "statistics fetch background routine", "error", err)
								continue
							}
						}

						if _, ok := stats["executions"]; ok {
							executions, err = strconv.ParseInt(stats["executions"], 10, 64)
							if err != nil {
								utils.Error("msg", "statistics fetch background routine", "error", err)
								continue
							}
						}

						if _, ok := stats["ramgbs"]; ok {
							ramgbs, err = strconv.ParseFloat(stats["ramgbs"], 64)
							if err != nil {
								utils.Error("msg", "statistics fetch background routine", "error", err)
								continue
							}
						}

						if _, ok := stats["ingress"]; ok {
							ingress, err = strconv.ParseFloat(stats["ingress"], 64)
							if err != nil {
								utils.Error("msg", "statistics fetch background routine", "error", err)
								continue
							}
						}

						if _, ok := stats["egress"]; ok {
							egress, err = strconv.ParseFloat(stats["egress"], 64)
							if err != nil {
								utils.Error("msg", "statistics fetch background routine", "error", err)
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

						err = objectdb.Client.UpsertUserStatistics(usrstat)
						if err != nil {
							lock.release()
							lock = nil
							cursor = ""
							utils.Error("msg", "statistics fetch background routine", "error", err)
							continue F
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
						utils.Error("msg", "certificates background routine", "error", err)
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
