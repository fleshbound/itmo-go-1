package main

import (
	"bufio"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	serverURL    = "http://srv.msk01.gigacorp.local/_stats"
	pollInterval = 5 * time.Second
	loadThreshold      = 30
	memoryThreshold    = 80  // 80%
	diskThreshold      = 90  // 90%
	networkThreshold   = 90  // 90%
	bytesInMb     = 1024 * 1024
	bytesInMbit   = 125000 // 1 Mbit/s = 125,000 bytes/s
)

func main() {
	errorCount := 0
	
	for {
		stats, err := fetchStats()
		if err != nil {
			errorCount++
			fmt.Printf("Error fetching stats: %v\n", err)
			
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
				errorCount = 0
			}
			
			time.Sleep(pollInterval)
			continue
		}
		
		errorCount = 0
		checkThresholds(stats)
		time.Sleep(pollInterval)
	}
}

type ServerStats struct {
	LoadAverage         uint64
	TotalMemory         uint64
	UsedMemory          uint64
	TotalDisk           uint64
	UsedDisk            uint64
	TotalNetwork        uint64
	UsedNetwork         uint64
}

func fetchStats() (*ServerStats, error) {
	resp, err := http.Get(serverURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP status: %s", resp.Status)
	}
	
	scanner := bufio.NewScanner(resp.Body)
	if !scanner.Scan() {
		return nil, fmt.Errorf("empty response")
	}
	
	line := scanner.Text()
	parts := strings.Split(line, ",")
	
	if len(parts) != 7 {
		return nil, fmt.Errorf("invalid data format: expected 7 values, got %d", len(parts))
	}
	
	stats := &ServerStats{}
	
	if stats.LoadAverage, err = strconv.ParseUint(parts[0], 10, 64); err != nil {
		return nil, fmt.Errorf("invalid load average: %v", err)
	}
	
	if stats.TotalMemory, err = strconv.ParseUint(parts[1], 10, 64); err != nil {
		return nil, fmt.Errorf("invalid total memory: %v", err)
	}
	
	if stats.UsedMemory, err = strconv.ParseUint(parts[2], 10, 64); err != nil {
		return nil, fmt.Errorf("invalid used memory: %v", err)
	}
	
	if stats.TotalDisk, err = strconv.ParseUint(parts[3], 10, 64); err != nil {
		return nil, fmt.Errorf("invalid total disk: %v", err)
	}
	
	if stats.UsedDisk, err = strconv.ParseUint(parts[4], 10, 64); err != nil {
		return nil, fmt.Errorf("invalid used disk: %v", err)
	}
	
	if stats.TotalNetwork, err = strconv.ParseUint(parts[5], 10, 64); err != nil {
		return nil, fmt.Errorf("invalid total network: %v", err)
	}
	
	if stats.UsedNetwork, err = strconv.ParseUint(parts[6], 10, 64); err != nil {
		return nil, fmt.Errorf("invalid used network: %v", err)
	}
	
	return stats, nil
}

func checkThresholds(stats *ServerStats) {
	// Load Average
	if stats.LoadAverage > uint64(loadThreshold) {
		fmt.Printf("Load Average is too high: %d\n", stats.LoadAverage)
	}
	
	// Memory usage
	if stats.TotalMemory > 0 {
		memoryPercent := (stats.UsedMemory * 100) / stats.TotalMemory
		if memoryPercent > uint64(memoryThreshold) {
			fmt.Printf("Memory usage too high: %d%%\n", memoryPercent)
		}
	}
	
	// Disk space
	if stats.TotalDisk > 0 {
		diskPercent := (stats.UsedDisk * 100) / stats.TotalDisk
		if diskPercent > uint64(diskThreshold) {
			freeSpaceMB := (stats.TotalDisk - stats.UsedDisk) / bytesInMb
			fmt.Printf("Free disk space is too low: %d Mb left\n", freeSpaceMB)
		}
	}
	
	// Network bandwidth
	if stats.TotalNetwork > 0 {
		networkPercent := (stats.UsedNetwork * 100) / stats.TotalNetwork
		if networkPercent > uint64(networkThreshold) {
			freeBandwidthMbit := (stats.TotalNetwork - stats.UsedNetwork) / bytesInMbit
			fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", freeBandwidthMbit)
		}
	}
}
