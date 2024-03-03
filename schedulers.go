package main

import (
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
		serviceTime     int64
		totalWait       float64
		totalTurnaround float64
		lastCompletion  float64
		schedule        = make([][]string, len(processes))
		gantt           = make([]TimeSlice, 0)
	)
	remaining := make([]Process, len(processes))
	copy(remaining, processes)

	byArrivalTime := func(p1, p2 *Process) bool {
		return p1.ArrivalTime < p2.ArrivalTime
	}

	sort.SliceStable(remaining, byArrivalTime)

	for len(remaining) > 0 {
		next := findShortestJob(remaining, serviceTime)
		if next == nil {
			// No available jobs
			serviceTime++
			continue
		}

		process := *next
		remaining = removeProcess(remaining, process)

		waitingTime := serviceTime - process.ArrivalTime
		if waitingTime < 0 {
			waitingTime = 0
		}
		totalWait += float64(waitingTime)

		start := serviceTime

		turnaround := process.BurstDuration + waitingTime
		totalTurnaround += float64(turnaround)

		completion := process.BurstDuration + serviceTime
		lastCompletion = float64(completion)

		schedule[process.ProcessID-1] = []string{
			fmt.Sprint(process.ProcessID),
			fmt.Sprint(process.Priority),
			fmt.Sprint(process.BurstDuration),
			fmt.Sprint(process.ArrivalTime),
			fmt.Sprint(waitingTime),
			fmt.Sprint(turnaround),
			fmt.Sprint(completion),
		}

		gantt = append(gantt, TimeSlice{
			PID:   process.ProcessID,
			Start: start,
			Stop:  completion,
		})

		serviceTime += process.BurstDuration
	}

	count := float64(len(processes))
	aveWait := totalWait / count
	aveTurnaround := totalTurnaround / count
	aveThroughput := count / lastCompletion

	outputTitle(w, title)
	outputGantt(w, gantt)
	outputSchedule(w, schedule, aveWait, aveTurnaround, aveThroughput)
}

func findShortestJob(remaining []Process, serviceTime int64) *Process {
	var shortest *Process
	for i := range remaining {
		if remaining[i].ArrivalTime > serviceTime {
			break
		}
		if shortest == nil || remaining[i].BurstDuration < shortest.BurstDuration {
			shortest = &remaining[i]
		}
	}
	return shortest
}

func removeProcess(processes []Process, process Process) []Process {
	var remaining []Process
	for i := range processes {
		if processes[i].ProcessID != process.ProcessID {
			remaining = append(remaining, processes[i])
		}
	}
	return remaining
}

type Process1 struct {
	Name       string
	Burst      int
	Arrival    int
	Priority   int
	Completed  bool
	Turnaround int
	Waiting    int
}

func SJFPrioritySchedule(w io.Writer, title string, processes []Process1) {
	fmt.Fprintf(w, "------ %s ------\n", title)

	completed := 0
	currentTime := 0
	var waiting []Process1
	var active *Process1

	for completed < len(processes) {
		for i := range processes {
			if !processes[i].Completed && processes[i].Arrival <= currentTime {
				waiting = append(waiting, processes[i])
			}
		}
		sort.Slice(waiting, func(i, j int) bool {
			return waiting[i].Burst < waiting[j].Burst
		})
		if active == nil && len(waiting) > 0 {
			active = &waiting[0]
			waiting = waiting[1:]
		}
		if active != nil {
			active.Burst--
			if active.Burst == 0 {
				active.Completed = true
				completed++
				active.Turnaround = currentTime + 1 - active.Arrival
				active.Waiting = active.Turnaround - active.Priority
				active = nil
			}
		}
		currentTime++
	}

	var totalTurnaround, totalWaiting int
	for i := range processes {
		totalTurnaround += processes[i].Turnaround
		totalWaiting += processes[i].Waiting
	}

	fmt.Fprintf(w, "Average turnaround time: %.2f\n", float64(totalTurnaround)/float64(len(processes)))
	fmt.Fprintf(w, "Average waiting time: %.2f\n", float64(totalWaiting)/float64(len(processes)))
	fmt.Fprintf(w, "Throughput: %.2f\n", float64(len(processes))/float64(currentTime))
}

func RRSchedule(w io.Writer, title string, processes []Process) {}

//endregion
