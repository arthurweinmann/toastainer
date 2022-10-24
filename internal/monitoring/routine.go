package monitoring

import (
	"context"
	"math"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/toastate/toastainer/internal/config"
	"github.com/toastate/toastainer/internal/db/redisdb"
	"github.com/toastate/toastainer/internal/model"
	"github.com/toastate/toastainer/internal/utils"
)

func routine() {
	gather := make([]metric, 0, 6)
	var prevcpustats []cpu.TimesStat

	for {
		time.Sleep(10 * time.Second)

		memstats, err := mem.VirtualMemory()
		if err != nil {
			utils.Error("origin", "monitoring:routine:mem.VirtualMemory", "error", err)
			continue
		}

		cputimestat, err := cpu.Times(false)
		if err != nil {
			utils.Error("origin", "monitoring:routine:cpu.Times", "error", err)
			continue
		}

		var totalcpuseconds, percentUsed float64
		if prevcpustats != nil {
			totalcpuseconds, percentUsed = calculateBusy(cputimestat[0], prevcpustats[0])
		}
		prevcpustats = cputimestat

		diskStats, err := disk.Usage(config.Runner.BTRFSMountPoint)
		if err != nil {
			utils.Error("origin", "monitoring:disk.Usage", "error", err)
			continue
		}

		gather = append(gather, metric{
			freecpupercent:  1.0 - percentUsed,
			totalcpuseconds: totalcpuseconds,
			freememorybytes: memstats.Available,
			freediskbytes:   diskStats.Free,
		})

		if len(gather) < 6 {
			continue
		}

		var avgfreemembytes uint64
		var avgfreecpupercent uint64
		var avgfreediskbytes uint64
		var sumtotalcpuseconds uint64
		for i := 0; i < 6; i++ {
			avgfreemembytes += gather[i].freememorybytes
			avgfreecpupercent += uint64(gather[i].freecpupercent * 10000.0)
			avgfreediskbytes += gather[i].freediskbytes

			sumtotalcpuseconds += uint64(gather[i].totalcpuseconds)
		}
		avgfreemembytes = avgfreemembytes / 6
		avgfreecpupercent = avgfreecpupercent / 6
		avgfreediskbytes = avgfreediskbytes / 6

		sm := &model.ServerMonitoring{
			TotalMemoryBytes:  memstats.Total,
			FreeMemoryBytes:   avgfreemembytes,
			TotalCPUSeconds:   sumtotalcpuseconds,
			FreeCPUPercentage: avgfreecpupercent,
			TotalDiskBytes:    diskStats.Total,
			FreeDiskBytes:     avgfreediskbytes,
			// ToasterStartupLatencyNanoseconds: ,
		}

		var ipid string
		if config.LocalPrivateIP != "" {
			ipid = config.LocalPrivateIP
		} else if config.LocalPublicIP != "" {
			ipid = config.LocalPublicIP
		}

		if ipid != "" {
			err = redisdb.GetClient().Set(context.Background(), "mon_"+ipid, sm.Marshal(), 2*time.Minute).Err()
			if err != nil {
				utils.Error("origin", "monitoring:redisdb.Set", "error", err)
				continue
			}
		}

		gather = gather[:0]
	}
}

type metric struct {
	freecpupercent  float64
	totalcpuseconds float64
	freememorybytes uint64
	freediskbytes   uint64
}

func calculateBusy(t1, t2 cpu.TimesStat) (float64, float64) {
	t1All, t1Busy := getAllBusy(t1)
	t2All, t2Busy := getAllBusy(t2)

	if t2Busy <= t1Busy {
		return t2All - t1All, 0
	}
	if t2All <= t1All {
		return t2All - t1All, 100
	}
	return t2All - t1All, math.Min(100, math.Max(0, (t2Busy-t1Busy)/(t2All-t1All)*100))
}

func getAllBusy(t cpu.TimesStat) (float64, float64) {
	tot := t.Total()
	if runtime.GOOS == "linux" {
		tot -= t.Guest     // Linux 2.6.24+
		tot -= t.GuestNice // Linux 3.2.0+
	}

	busy := tot - t.Idle - t.Iowait

	return tot, busy
}
