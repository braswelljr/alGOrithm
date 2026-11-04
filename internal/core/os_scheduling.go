package core

import (
	"fmt"
	"strconv"

	"github.com/braswelljr/alGOrithm/pkg/os/scheduling"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// schedulingEntries is every entry of the scheduling category.
func schedulingEntries() []Entry {
	return []Entry{
		edfSchedulingEntry(),
		fcfsSchedulingEntry(),
		lotterySchedulingEntry(),
		mlfqSchedulingEntry(),
		prioritySchedulingEntry(),
		rateMonotonicEntry(),
		roundRobinEntry(),
		sjfSchedulingEntry(),
		srtfSchedulingEntry(),
	}
}

const tasksExample = "name period burst\nT1 5 2\nT2 7 4"

func schedulingRt(id, name, fn, summary string, run func([]scheduling.Task) (scheduling.RealTime, error)) Entry {
	e := entry(id, name, "os/scheduling/realtime.go", fn)
	e.Summary, e.Time, e.Space, e.Input = summary, "O(H * n) for a hyperperiod of H", "O(H)", Processes
	e.Example = &Input{Text: tasksExample}
	e.Limits = "tasks as \"name period burst\" lines"
	e.Run = func(in Input, _ trace.Recorder) (Output, error) {
		ts, err := scheduling.ParseTasks(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		res, err := run(ts)
		if err != nil {
			return Output{}, wrap(err)
		}
		out := Output{Bool: ptr(len(res.Missed) == 0), Summary: map[string]float64{
			"utilization": res.Utilization, "Liu-Layland bound": res.Bound, "missed deadlines": float64(len(res.Missed)), "hyperperiod": float64(res.Horizon),
		}}
		for _, sl := range res.Timeline {
			lane, state := sl.Process, "running"
			if lane == "" {
				lane, state = "idle", "idle"
			}
			out.Timeline = append(out.Timeline, Span{Lane: lane, Label: lane, Start: sl.Start, End: sl.End, State: state})
		}
		for _, m := range res.Missed {
			out.Lines = append(out.Lines, fmt.Sprintf("%s's job released at %d missed its deadline %d with %d ticks left", m.Task, m.Release, m.Deadline, m.Left))
		}
		if len(res.Missed) == 0 {
			out.Lines = []string{"every deadline met"}
		}
		return out, nil
	}
	e.Call = func(in Input) (string, string) {
		return fmt.Sprintf("ts, err := scheduling.ParseTasks(%s)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", "`"+in.Text+"`"), "scheduling." + fn + "(ts)"
	}
	return e
}

func lotterySchedulingEntry() Entry {
	lottery := entry("lottery-scheduling", "Lottery Scheduling", "os/scheduling/lottery.go", "Lottery")
	lottery.Summary, lottery.Time, lottery.Space, lottery.Input = "Each tick, draw a ticket at random: a process gets the CPU in proportion to its tickets (the priority column).", "O(T * n)", "O(T)", Processes
	lottery.Options = []Option{{"seed", "draw seed", 1, 0, 1_000_000}}
	lottery.Example = &Input{Text: "name arrival burst tickets\nA 0 6 6\nB 0 6 3\nC 0 6 1", Options: map[string]int{"seed": 1}}
	lottery.Run = func(in Input, _ trace.Recorder) (Output, error) {
		ps, err := scheduling.Parse(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		s, err := scheduling.Lottery(ps, int64(in.opt("seed", 1)))
		if err != nil {
			return Output{}, wrap(err)
		}
		return scheduleOutput(s), nil
	}
	lottery.Call = func(in Input) (string, string) {
		return fmt.Sprintf("ps, err := scheduling.Parse(%s)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", "`"+in.Text+"`"), fmt.Sprintf("scheduling.Lottery(ps, %d)", in.opt("seed", 1))
	}
	return lottery
}

func edfSchedulingEntry() Entry {
	return schedulingRt("edf-scheduling", "Earliest Deadline First", "EDF", "Periodic real-time tasks: always run the job whose deadline is nearest; optimal on one CPU when utilization is at most 1.", scheduling.EDF)
}

func rateMonotonicEntry() Entry {
	return schedulingRt("rate-monotonic", "Rate-Monotonic Scheduling", "RateMonotonic", "Periodic real-time tasks with fixed priorities, shorter period first; safe below the Liu-Layland bound.", scheduling.RateMonotonic)
}

const processesExample = "name arrival burst priority\nP1 0 8 3\nP2 1 4 1\nP3 2 9 4\nP4 3 5 2\nP5 6 2 5"

// scheduleOutput turns a Schedule into a timeline, a statistics table and a
// summary.
func scheduleOutput(s scheduling.Schedule) Output {
	out := Output{Summary: map[string]float64{
		"average waiting": s.AvgWaiting, "average turnaround": s.AvgTurnaround, "average response": s.AvgResponse,
		"context switches": float64(s.ContextSwitches), "makespan": float64(s.Makespan),
	}}
	for _, sl := range s.Timeline {
		lane, state, label := sl.Process, "running", sl.Process
		if lane == "" {
			lane, state, label = "idle", "idle", "idle"
		}
		if sl.Queue > 0 {
			label += " (queue " + strconv.Itoa(sl.Queue) + ")"
		}
		out.Timeline = append(out.Timeline, Span{Lane: lane, Label: label, Start: sl.Start, End: sl.End, State: state})
	}
	cols := []string{"arrival", "burst", "priority", "start", "finish", "turnaround", "waiting", "response"}
	t := &TableOutput{Rows: len(s.Stats), Cols: len(cols), ColLabels: cols}
	for _, st := range s.Stats {
		t.RowLabels = append(t.RowLabels, st.Name)
		t.Cells = append(t.Cells, st.Arrival, st.Burst, st.Priority, st.Start, st.Finish, st.Turnaround, st.Waiting, st.Response)
	}
	out.Table = t
	return out
}

func schedulingMk(id, name, fn, summary string, opts []Option, run func(ps []scheduling.Process, in Input) (scheduling.Schedule, error)) Entry {
	e := entry(id, name, "os/scheduling/algorithms.go", fn)
	e.Summary, e.Time, e.Space, e.Input, e.Options = summary, "O(T * n) for T ticks", "O(T)", Processes, opts
	e.Example = &Input{Text: processesExample, Options: map[string]int{}}
	e.Limits = fmt.Sprintf("at most %d processes and %d ticks of work", scheduling.MaxProcesses, scheduling.MaxTime)
	e.Run = func(in Input, _ trace.Recorder) (Output, error) {
		ps, err := scheduling.Parse(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		s, err := run(ps, in)
		if err != nil {
			return Output{}, wrap(err)
		}
		return scheduleOutput(s), nil
	}
	return e
}

func fcfsSchedulingEntry() Entry {
	return schedulingMk("fcfs-scheduling", "First Come, First Served", "FCFS", "Run processes to completion in order of arrival; one long job makes everyone wait (the convoy effect).", nil,
		func(ps []scheduling.Process, _ Input) (scheduling.Schedule, error) { return scheduling.FCFS(ps) })
}

func sjfSchedulingEntry() Entry {
	return schedulingMk("sjf-scheduling", "Shortest Job First", "SJF", "When the CPU frees up, run the shortest ready job to completion; least average waiting without preemption.", nil,
		func(ps []scheduling.Process, _ Input) (scheduling.Schedule, error) { return scheduling.SJF(ps) })
}

func srtfSchedulingEntry() Entry {
	return schedulingMk("srtf-scheduling", "Shortest Remaining Time First", "SRTF", "Preemptive SJF: a new arrival with less work left takes the CPU; least average waiting of all.", nil,
		func(ps []scheduling.Process, _ Input) (scheduling.Schedule, error) { return scheduling.SRTF(ps) })
}

func roundRobinEntry() Entry {
	return schedulingMk("round-robin", "Round Robin", "RoundRobin", "Each ready process runs for one time quantum in turn; good response, nothing starves.",
		[]Option{{"quantum", "time quantum", 2, 1, 100}},
		func(ps []scheduling.Process, in Input) (scheduling.Schedule, error) {
			return scheduling.RoundRobin(ps, in.opt("quantum", 2))
		})
}

func prioritySchedulingEntry() Entry {
	return schedulingMk("priority-scheduling", "Priority Scheduling", "Priority", "Run the most urgent ready process (lowest number); with preemption a more urgent arrival takes over.",
		[]Option{{"preempt", "preemptive (1 yes, 0 no)", 0, 0, 1}},
		func(ps []scheduling.Process, in Input) (scheduling.Schedule, error) {
			return scheduling.Priority(ps, in.opt("preempt", 0) == 1)
		})
}

func mlfqSchedulingEntry() Entry {
	return schedulingMk("mlfq-scheduling", "Multilevel Feedback Queue", "MLFQ", "Round-robin queues by urgency; using a whole quantum moves a process down, so short jobs finish fast.",
		[]Option{{"quantum", "top quantum (doubles each level)", 2, 1, 50}, {"levels", "levels", 3, 1, 6}, {"boost", "priority boost every N ticks (0 off)", 0, 0, 10000}},
		func(ps []scheduling.Process, in Input) (scheduling.Schedule, error) {
			q := make([]int, in.opt("levels", 3))
			for i := range q {
				q[i] = in.opt("quantum", 2) << i
			}
			return scheduling.MLFQ(ps, q, in.opt("boost", 0))
		})
}
