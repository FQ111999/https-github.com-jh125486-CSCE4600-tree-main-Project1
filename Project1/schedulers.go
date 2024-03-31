package main

import (
	"container/heap"
	"fmt"
	"io"
	"sort"
)

type (
	Process struct {
		ProcessID     string
		ArrivalTime   int64
		BurstDuration int64
		Priority      int64
	}

	TimeSlice struct {
		PID   string
		Start int64
		Stop  int64
	}
)

//region Schedulers

// FCFSSchedule outputs a schedule of processes in a GANTT chart and a table of timing given:
// • an output writer
// • a title for the chart
// • a slice of processes
func FCFSSchedule(w io.Writer, title string, processes []Process) {
	var (
		serviceTime     int64
		totalWait       float64
		totalTurnaround float64
		lastCompletion  float64
		waitingTime     int64
		schedule        = make([][]string, len(processes))
		gantt           = make([]TimeSlice, 0)
	)

	for i := range processes {
		if processes[i].ArrivalTime > 0 {
			waitingTime = serviceTime - processes[i].ArrivalTime
		}
		totalWait += float64(waitingTime)

		start := waitingTime + processes[i].ArrivalTime

		turnaround := processes[i].BurstDuration + waitingTime
		totalTurnaround += float64(turnaround)

		completion := processes[i].BurstDuration + processes[i].ArrivalTime + waitingTime
		lastCompletion = float64(completion)

		schedule[i] = []string{
			fmt.Sprint(processes[i].ProcessID),
			fmt.Sprint(processes[i].Priority),
			fmt.Sprint(processes[i].BurstDuration),
			fmt.Sprint(processes[i].ArrivalTime),
			fmt.Sprint(waitingTime),
			fmt.Sprint(turnaround),
			fmt.Sprint(completion),
		}
		serviceTime += processes[i].BurstDuration

		gantt = append(gantt, TimeSlice{
			PID:   processes[i].ProcessID,
			Start: start,
			Stop:  serviceTime,
		})
	}

	count := float64(len(processes))
	aveWait := totalWait / count
	aveTurnaround := totalTurnaround / count
	aveThroughput := count / lastCompletion

	outputTitle(w, title)
	outputGantt(w, gantt)
	outputSchedule(w, schedule, aveWait, aveTurnaround, aveThroughput)
}

func SJFSchedule(w io.Writer, title string, processes []Process) {
	var (
		totalWait       float64
		totalTurnaround float64
		lastCompletion  float64
		currentTime     int64
		waitingTime     = make(map[string]int64)
		remainingTime   = make(map[string]int64)
		completed       = make(map[string]bool)
		readyQueue      = make([]Process, 0)
		gantt           = make([]TimeSlice, 0)
	)

	for _, p := range processes {
		remainingTime[p.ProcessID] = p.BurstDuration
	}

	for len(completed) < len(processes) {
		for i, p := range processes {
			if p.ArrivalTime <= currentTime && !completed[p.ProcessID] {
				readyQueue = append(readyQueue, processes[i])
				delete(processes, i)
			}
		}

		if len(readyQueue) == 0 {
			currentTime++
			continue
		}

		sort.SliceStable(readyQueue, func(i, j int) bool {
			return remainingTime[readyQueue[i].ProcessID] < remainingTime[readyQueue[j].ProcessID]
		})

		currentProcess := readyQueue[0]
		readyQueue = readyQueue[1:]

		if _, ok := waitingTime[currentProcess.ProcessID]; !ok {
			waitingTime[currentProcess.ProcessID] = currentTime - currentProcess.ArrivalTime
		}

		start := currentTime
		currentTime++

		gantt = append(gantt, TimeSlice{
			PID:   currentProcess.ProcessID,
			Start: start,
			Stop:  currentTime,
		})

		remainingTime[currentProcess.ProcessID]--
		if remainingTime[currentProcess.ProcessID] == 0 {
			completed[currentProcess.ProcessID] = true
			turnaround := currentTime - currentProcess.ArrivalTime
			totalTurnaround += float64(turnaround)
			totalWait += float64(waitingTime[currentProcess.ProcessID])
			lastCompletion = float64(currentTime)
			continue
		}

		for _, p := range processes {
			if p.ArrivalTime <= currentTime && !completed[p.ProcessID] && p.BurstDuration < remainingTime[currentProcess.ProcessID] {
				readyQueue = append(readyQueue, p)
				delete(processes, p.ProcessID)
			}
		}
	}

	count := float64(len(processes))
	aveWait := totalWait / count
	aveTurnaround := totalTurnaround / count
	aveThroughput := count / lastCompletion

	outputTitle(w, title)
	outputGantt(w, gantt)
	outputSchedule(w, schedule, aveWait, aveTurnaround, aveThroughput)
}

func SJFPrioritySchedule(w io.Writer, title string, processes []Process) {
	var (
		totalWait       float64
		totalTurnaround float64
		lastCompletion  float64
		currentTime     int64
		waitingTime     = make(map[string]int64)
		completed       = make(map[string]bool)
		readyQueue      = make(PriorityQueue, 0)
		gantt           = make([]TimeSlice, 0)
	)

	for len(completed) < len(processes) {
		for i, p := range processes {
			if p.ArrivalTime <= currentTime && !completed[p.ProcessID] {
				heap.Push(&readyQueue, &Item{
					Value:    p,
					Priority: p.Priority,
				})
				delete(processes, i)
			}
		}

		if len(readyQueue) == 0 {
			currentTime++
			continue
		}

		currentProcess := heap.Pop(&readyQueue).(*Item).Value.(Process)

		if _, ok := waitingTime[currentProcess.ProcessID]; !ok {
			waitingTime[currentProcess.ProcessID] = currentTime - currentProcess.ArrivalTime
		}

		start := currentTime
		currentTime++

		gantt = append(gantt, TimeSlice{
			PID:   currentProcess.ProcessID,
			Start: start,
			Stop:  currentTime,
		})

		remainingTime := currentProcess.BurstDuration - 1
		if remainingTime == 0 {
			completed[currentProcess.ProcessID] = true
			turnaround := currentTime - currentProcess.ArrivalTime
			totalTurnaround += float64(turnaround)
			totalWait += float64(waitingTime[currentProcess.ProcessID])
			lastCompletion = float64(currentTime)
			continue
		}

		for _, p := range processes {
			if p.ArrivalTime <= currentTime && !completed[p.ProcessID] && p.Priority < currentProcess.Priority {
				heap.Push(&readyQueue, &Item{
					Value:    p,
					Priority: p.Priority,
				})
				delete(processes, p.ProcessID)
			}
		}

		heap.Push(&readyQueue, &Item{
			Value:    currentProcess,
			Priority: currentProcess.Priority,
		})
	}

	count := float64(len(processes))
	aveWait := totalWait / count
	aveTurnaround := totalTurnaround / count
	aveThroughput := count / lastCompletion

	outputTitle(w, title)
	outputGantt(w, gantt)
	outputSchedule(w, schedule, aveWait, aveTurnaround, aveThroughput)
}

func RRSchedule(w io.Writer, title string, processes []Process) {
	var (
		totalWait       float64
		totalTurnaround float64
		lastCompletion  float64
		currentTime     int64
		waitingTime     = make(map[string]int64)
		remainingTime   = make(map[string]int64)
		completed       = make(map[string]bool)
		readyQueue      = make([]Process, 0)
		gantt           = make([]TimeSlice, 0)
		timeQuantum     = int64(4)
	)

	for _, p := range processes {
		remainingTime[p.ProcessID] = p.BurstDuration
	}

	for len(completed) < len(processes) {
		for i, p := range processes {
			if p.ArrivalTime <= currentTime && !completed[p.ProcessID] {
				readyQueue = append(readyQueue, processes[i])
				delete(processes, i)
			}
		}

		if len(readyQueue) == 0 {
			currentTime++
			continue
		}

		currentProcess := readyQueue[0]
		readyQueue = readyQueue[1:]

		if _, ok := waitingTime[currentProcess.ProcessID]; !ok {
			waitingTime[currentProcess.ProcessID] = currentTime - currentProcess.ArrivalTime
		}

		start := currentTime
		executionTime := min(currentProcess.BurstDuration, timeQuantum)
		currentTime += executionTime
		remainingTime[currentProcess.ProcessID] -= executionTime

		gantt = append(gantt, TimeSlice{
			PID:   currentProcess.ProcessID,
			Start: start,
			Stop:  currentTime,
		})

		if remainingTime[currentProcess.ProcessID] == 0 {
			completed[currentProcess.ProcessID] = true
			turnaround := currentTime - currentProcess.ArrivalTime
			totalTurnaround += float64(turnaround)
			totalWait += float64(waitingTime[currentProcess.ProcessID])
			lastCompletion = float64(currentTime)
			continue
		}

		for _, p := range processes {
			if p.ArrivalTime <= currentTime && !completed[p.ProcessID] {
				readyQueue = append(readyQueue, p)
				delete(processes, p.ProcessID)
			}
		}

		readyQueue = append(readyQueue, currentProcess)
	}

	count := float64(len(processes))
	aveWait := totalWait / count
	aveTurnaround := totalTurnaround / count
	aveThroughput := count / lastCompletion

	outputTitle(w, title)
	outputGantt(w, gantt)
	outputSchedule(w, schedule, aveWait, aveTurnaround, aveThroughput)
}

//endregion

// Helper function to find minimum of two integers
func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

