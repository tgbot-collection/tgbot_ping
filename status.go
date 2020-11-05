// DailyGakki - status
// 2020-11-05 12:15
// Benny <benny.think@gmail.com>

package tgbot_ping

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

import "code.cloudfoundry.org/bytefmt"

type containerStart struct {
	State struct{ StartedAt time.Time }
}

// https://github.com/moby/moby/blob/c1d090fcc88fa3bc5b804aead91ec60e30207538/api/types/stats.go

type cpuUsage struct {
	TotalUsage        uint64   `json:"total_usage"`
	PercpuUsage       []uint64 `json:"percpu_usage,omitempty"`
	UsageInKernelmode uint64   `json:"usage_in_kernelmode"`
	UsageInUsermode   uint64   `json:"usage_in_usermode"`
}

type cpuStats struct {
	CPUUsage    cpuUsage `json:"cpu_usage"`
	SystemUsage uint64   `json:"system_cpu_usage,omitempty"`
	OnlineCPUs  uint32   `json:"online_cpus,omitempty"`
}

type memoryStats struct {
	Usage             uint64            `json:"usage,omitempty"`
	MaxUsage          uint64            `json:"max_usage,omitempty"`
	Stats             map[string]uint64 `json:"stats,omitempty"`
	Failcnt           uint64            `json:"failcnt,omitempty"`
	Limit             uint64            `json:"limit,omitempty"`
	Commit            uint64            `json:"commitbytes,omitempty"`
	CommitPeak        uint64            `json:"commitpeakbytes,omitempty"`
	PrivateWorkingSet uint64            `json:"privateworkingset,omitempty"`
}

type blkioStatEntry struct {
	Major uint64 `json:"major"`
	Minor uint64 `json:"minor"`
	Op    string `json:"op"`
	Value uint64 `json:"value"`
}

type blkioStats struct {
	IoServiceBytesRecursive []blkioStatEntry `json:"io_service_bytes_recursive"`
	IoServicedRecursive     []blkioStatEntry `json:"io_serviced_recursive"`
	IoQueuedRecursive       []blkioStatEntry `json:"io_queue_recursive"`
	IoServiceTimeRecursive  []blkioStatEntry `json:"io_service_time_recursive"`
	IoWaitTimeRecursive     []blkioStatEntry `json:"io_wait_time_recursive"`
	IoMergedRecursive       []blkioStatEntry `json:"io_merged_recursive"`
	IoTimeRecursive         []blkioStatEntry `json:"io_time_recursive"`
	SectorsRecursive        []blkioStatEntry `json:"sectors_recursive"`
}

type networkStats struct {
	RxBytes    uint64 `json:"rx_bytes"`
	RxPackets  uint64 `json:"rx_packets"`
	RxErrors   uint64 `json:"rx_errors"`
	RxDropped  uint64 `json:"rx_dropped"`
	TxBytes    uint64 `json:"tx_bytes"`
	TxPackets  uint64 `json:"tx_packets"`
	TxErrors   uint64 `json:"tx_errors"`
	TxDropped  uint64 `json:"tx_dropped"`
	EndpointID string `json:"endpoint_id,omitempty"`
	InstanceID string `json:"instance_id,omitempty"`
}

type stats struct {
	Read    time.Time `json:"read"`
	PreRead time.Time `json:"preread"`

	BlkioStats  blkioStats  `json:"blkio_stats,omitempty"`
	CPUStats    cpuStats    `json:"cpu_stats,omitempty"`
	PreCPUStats cpuStats    `json:"precpu_stats,omitempty"` // "Pre"="Previous"
	MemoryStats memoryStats `json:"memory_stats,omitempty"`
}

type statsJSON struct {
	stats
	Name string `json:"name,omitempty"`
	ID   string `json:"id,omitempty"`

	// Networks request version >=1.21
	Networks map[string]networkStats `json:"networks,omitempty"`
}

func calculateCPUPercentUnix(previousCPU, previousSystem uint64, v statsJSON) float64 {
	var (
		cpuPercent = 0.0
		// calculate the change for the cpu usage of the container in between readings
		cpuDelta = float64(v.CPUStats.CPUUsage.TotalUsage) - float64(previousCPU)
		// calculate the change for the entire system between readings
		systemDelta = float64(v.CPUStats.SystemUsage) - float64(previousSystem)
	)

	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * float64(len(v.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	}
	return cpuPercent
}

func calculateBlockIO(blkio blkioStats) (blkRead uint64, blkWrite uint64) {
	for _, bioEntry := range blkio.IoServiceBytesRecursive {
		switch strings.ToLower(bioEntry.Op) {
		case "read":
			blkRead = blkRead + bioEntry.Value
		case "write":
			blkWrite = blkWrite + bioEntry.Value
		}
	}
	return
}

func calculateNetwork(network map[string]networkStats) (float64, float64) {
	var rx, tx float64

	for _, v := range network {
		rx += float64(v.RxBytes)
		tx += float64(v.TxBytes)
	}
	return rx, tx
}

func formatSince(t time.Time) string {
	const (
		Decisecond = 100 * time.Millisecond
		Day        = 24 * time.Hour
	)
	ts := time.Since(t)
	sign := time.Duration(1)
	if ts < 0 {
		sign = -1
		ts = -ts
	}
	ts += +Decisecond / 2
	d := sign * (ts / Day)
	ts = ts % Day
	h := ts / time.Hour
	ts = ts % time.Hour
	m := ts / time.Minute
	ts = ts % time.Minute
	s := ts / time.Second
	ts = ts % time.Second
	f := ts / Decisecond
	return fmt.Sprintf("%d days %d hours %d minutes %d.%d seconds", d, h, m, s, f)
}

func getContainerInfo(containerName, displayName string) string {
	var (
		statData  statsJSON
		container containerStart

		cpuPercent        = 0.0
		blkRead, blkWrite uint64 // Only used on Linux
		mem               = 0.0
		previousCPU       uint64
		previousSystem    uint64

		statUrl    = fmt.Sprintf("http://socat:2375/containers/%s/stats?stream=0", containerName)
		runtimeUrl = fmt.Sprintf("http://socat:2375/containers/%s/json", containerName)
		template   = "%s has been running for 😊%s😭 from 😊%s😭😀\n" +
			"CPU: 😊%.2f%%😭\n" +
			"RAM: 😊%s😭\n" +
			"Network RX/TX: 😊%s/%s😭\n" +
			"IO R/W: 😊%s/%s😭"
	)
	// calculate start time
	runtimeResponse, _ := http.Get(runtimeUrl)

	_ = json.NewDecoder(runtimeResponse.Body).Decode(&container)
	_ = runtimeResponse.Body.Close()
	if container.State.StartedAt.Nanosecond() == 0 {
		return "Runtime information is not available outside of docker container.\n"
	}
	startTime := container.State.StartedAt.Local().Format("2006-01-02 15:04:05 -0700")

	// calculate cpu, memory....
	statResponse, _ := http.Get(statUrl)
	_ = json.NewDecoder(statResponse.Body).Decode(&statData)
	_ = statResponse.Body.Close()

	previousCPU = statData.PreCPUStats.CPUUsage.TotalUsage
	previousSystem = statData.PreCPUStats.SystemUsage
	cpuPercent = calculateCPUPercentUnix(previousCPU, previousSystem, statData)
	blkRead, blkWrite = calculateBlockIO(statData.BlkioStats)
	mem = float64(statData.MemoryStats.Usage)
	netRx, netTx := calculateNetwork(statData.Networks)

	return fmt.Sprintf(template, displayName,
		formatSince(container.State.StartedAt), startTime,
		cpuPercent, bytefmt.ByteSize(uint64(mem)),
		bytefmt.ByteSize(uint64(netRx)), bytefmt.ByteSize(uint64(netTx)),
		bytefmt.ByteSize(blkRead), bytefmt.ByteSize(blkWrite))
}

func GetRuntime(containerName, displayName, parseMode string) string {
	//container_name: str, display_name: str = "This bot", parse_mode: str = "markdown"

	info := getContainerInfo(containerName, displayName)

	if parseMode == "" {
		parseMode = "markdown"
	}
	switch parseMode {
	case "html":
		info = strings.Replace(info, "😊", "<pre>", -1)
		info = strings.Replace(info, "😭", "</pre>", -1)
	default:
		info = strings.Replace(info, "😊", "`", -1)
		info = strings.Replace(info, "😭", "`", -1)
	}

	return info
}
