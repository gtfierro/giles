package stats

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	netio "github.com/shirou/gopsutil/net"
	"net/http"
	"sync/atomic"
	"time"
)

// Struct to return as json
type systemStats struct {
	Time        int64     `json:"time"`
	UptimeSec   int       `json:"uptime_sec"`
	CpuPercents []float64 `json:"cpu_percents"`
	DiskUsedGB  float64   `json:"disk_used_gb"`
	DiskTotalGB float64   `json:"disk_total_gb"`
	DiskFreeGB  float64   `json:"disk_free_gb"`
	InodesUsed  uint64    `json:"inodes_used"`
	InodesTotal uint64    `json:"inodes_total"`
	InodesFree  uint64    `json:"inodes_free"`
	MemFreeGB   float64   `json:"mem_free_gb"`
	MemTotalGB  float64   `json:"mem_total_gb"`
	NetSentMB   float64   `json:"net_sent_mb"`
	NetRecvMB   float64   `json:"net_recv_mb"`
}

var curSystemStats atomic.Value

// collectSystemStats() will keep curSystemStats updated with stats
// about the local system. It uses atomic operations when updating
// curSystemStats.
func collectSystemStats() {
	for {
		bootTime, _ := host.BootTime()
		uptime_sec :=
			int(time.Since(time.Unix(int64(bootTime), 0)).Seconds())

		cpu_percents, err :=
			cpu.CPUPercent(time.Duration(0)*time.Second, true)
		if err != nil {
			//log.Info("CPUPercent fail", err)
			//return
		}

		diskstat, _ := disk.DiskUsage(".")
		btogb := float64(1000000000)
		btomb := float64(1000000)

		vmem, err := mem.VirtualMemory()
		if err != nil {
			log.Info("VirtualMemory fail", err)
			return
		}

		nicCounters, err := netio.NetIOCounters(true)
		if err != nil {
			log.Info("NetIOCounters fail", err)
			return
		}

		newSystemStats := systemStats{
			Time:        time.Now().Unix(),
			UptimeSec:   uptime_sec,
			CpuPercents: cpu_percents,
			DiskUsedGB:  float64(diskstat.Used) / btogb,
			DiskTotalGB: float64(diskstat.Total) / btogb,
			DiskFreeGB:  float64(diskstat.Free) / btogb,
			InodesUsed:  diskstat.InodesUsed,
			InodesTotal: diskstat.InodesTotal,
			InodesFree:  diskstat.InodesFree,
			MemFreeGB:   float64(vmem.Available) / btogb,
			MemTotalGB:  float64(vmem.Total) / btogb,
			NetSentMB:   float64(nicCounters[0].BytesSent) / btomb,
			NetRecvMB:   float64(nicCounters[0].BytesRecv) / btomb,
		}
		curSystemStats.Store(newSystemStats)

		time.Sleep(time.Second)
	}
}

func systemStatsHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	jsonSystemStats, err := json.Marshal(curSystemStats.Load())
	if err != nil {
		fmt.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonSystemStats)
}
