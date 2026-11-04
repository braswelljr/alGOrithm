package design

import (
	"fmt"
	"math/rand"
	"sort"
)

// VideoConfig is a video platform's uploads and its transcoding workers.
type VideoConfig struct {
	Videos     int   // videos uploaded at once
	Chunks     int   // most chunks a video is split into (each a few seconds)
	Renditions int   // resolutions each chunk is transcoded to (240p, 480p, 720p, 1080p...)
	Workers    int   // transcoding workers
	Seed       int64 // video lengths
}

// VideoPipeline models a video upload pipeline (YouTube, Netflix): the
// upload is split into chunks stored in a blob store; a job per chunk and
// rendition goes on a queue; workers transcode jobs in parallel, one tick
// each and longer for higher resolutions; a video is ready when its last job
// finishes, and its renditions are pushed to the CDN. Splitting into chunks
// is what lets a long video finish in about the time of its longest chunk on
// enough workers. It leaves out audio, DRM and thumbnails.
//
// Time: O(jobs log jobs + jobs * Workers). Space: O(jobs).
func VideoPipeline(c VideoConfig) (Report, error) {
	if err := check(c.Videos >= 1 && c.Videos <= 1000 && c.Chunks >= 1 && c.Chunks <= 1000 && c.Renditions >= 1 && c.Renditions <= 8 && c.Workers >= 1 && c.Workers <= 1000,
		"1 to 1,000 videos, chunks and workers, 1 to 8 renditions"); err != nil {
		return Report{}, err
	}
	rng := rand.New(rand.NewSource(c.Seed))
	type job struct{ video, cost int }
	var jobs []job
	chunks := make([]int, c.Videos)
	for v := range chunks {
		chunks[v] = 1 + rng.Intn(c.Chunks)
		for k := 0; k < chunks[v]; k++ {
			for q := 0; q < c.Renditions; q++ {
				jobs = append(jobs, job{v, 1 + q})
			}
		}
	}
	free := make([]int, c.Workers)
	ready := make([]int, c.Videos)
	busy := 0
	for _, j := range jobs {
		w := 0
		for i := range free {
			if free[i] < free[w] {
				w = i
			}
		}
		free[w] += j.cost
		busy += j.cost
		ready[j.video] = max(ready[j.video], free[w])
	}
	makespan := 0
	for _, f := range free {
		makespan = max(makespan, f)
	}
	a := arch([]Component{
		{ID: "creator", Kind: Client, Label: "Creator"},
		{ID: "upload", Kind: Service, Label: "Upload service", Notes: "splits into chunks"},
		{ID: "raw", Kind: Blob, Label: "Raw chunks"},
		{ID: "queue", Kind: Queue, Label: "Transcode jobs"},
		{ID: "workers", Kind: Worker, Label: fmt.Sprintf("%d transcoders", c.Workers)},
		{ID: "out", Kind: Blob, Label: "Renditions"},
		{ID: "cdn", Kind: CDN, Label: "CDN"},
		{ID: "viewer", Kind: Client, Label: "Viewer"},
	}, "creator>upload", "upload>raw", "upload>queue", "queue>workers", "workers>out", "out>cdn", "cdn>viewer:stream")
	for v := 0; v < c.Videos && a.recording(); v++ {
		f := a.frame(fmt.Sprintf("video %d: %d chunks x %d renditions = %d jobs, ready at tick %d", v, chunks[v], c.Renditions, chunks[v]*c.Renditions, ready[v]))
		a.send(f, "creator", "upload", v, OK)
		a.send(f, "upload", "raw", v, OK)
		a.send(f, "upload", "queue", v, Queued)
		a.send(f, "queue", "workers", v, OK)
		a.send(f, "workers", "out", v, OK)
		a.send(f, "out", "cdn", v, OK)
		f.load("queue", len(jobs), len(jobs), -1)
	}
	last := a.frame("all renditions on the CDN")
	a.send(last, "cdn", "viewer", 0, OK)
	sorted := append([]int(nil), ready...)
	sort.Ints(sorted)
	var r Report
	r.Arch = a
	r.say("%d videos, %d jobs on %d workers: everything ready at tick %d, workers %d%% busy", c.Videos, len(jobs), c.Workers, makespan, percent(busy, makespan*c.Workers))
	r.say("median video ready at tick %d; one worker alone would need %d ticks", sorted[len(sorted)/2], busy)
	r.put("jobs", float64(len(jobs)))
	r.put("ready (ticks)", float64(makespan))
	r.put("worker utilization (%)", float64(percent(busy, makespan*c.Workers)))
	return r, nil
}

// ABRConfig is a viewer's network over a playback.
type ABRConfig struct {
	Seconds      int   // seconds of playback
	StartKbps    int   // the network's bandwidth at the start
	BufferTarget int   // seconds of video the player tries to keep buffered
	Seed         int64 // how bandwidth wanders
}

// Ladder is the bitrate ladder of a stream, in kbps: 240p to 4K.
var Ladder = []int{400, 800, 1500, 3000, 6000, 15000}

// AdaptiveBitrate models adaptive bitrate streaming (HLS, DASH): every second
// the player downloads the next segment at the highest bitrate on the ladder
// below 80% of the bandwidth it measured, dropping a step when the buffer runs
// low and holding a lower step until the buffer recovers. When a segment
// takes longer to download than the buffer holds, playback stalls
// (rebuffers). Bandwidth wanders as a seeded random walk. It leaves out
// segment-size variation and startup delay.
//
// Time: O(Seconds). Space: O(Seconds).
func AdaptiveBitrate(c ABRConfig) (Report, error) {
	if err := check(c.Seconds >= 1 && c.Seconds <= 3600 && c.StartKbps >= 100 && c.BufferTarget >= 1 && c.BufferTarget <= 60, "1 to 3,600 seconds, at least 100 kbps, a buffer target of 1 to 60 seconds"); err != nil {
		return Report{}, err
	}
	rng := rand.New(rand.NewSource(c.Seed))
	bw := float64(c.StartKbps)
	buffer := 2.0
	var r Report
	t := r.table("bandwidth (kbps)", "bitrate (kbps)", "buffer (s x10)")
	stalls, switches, prev, sum := 0, 0, -1, 0
	for s := 0; s < c.Seconds; s++ {
		bw *= 0.75 + rng.Float64()*0.5
		bw = max(150, min(bw, 25_000))
		step := 0
		for i, b := range Ladder {
			if float64(b) <= 0.8*bw {
				step = i
			}
		}
		if buffer < float64(c.BufferTarget)/3 && step > 0 {
			step--
		}
		download := float64(Ladder[step]) / bw // seconds to fetch one second of video
		if download > buffer {
			stalls++
			buffer = 0
		} else {
			buffer -= download
		}
		buffer = min(buffer+1, float64(c.BufferTarget))
		if prev >= 0 && step != prev {
			switches++
		}
		prev = step
		sum += Ladder[step]
		if len(t.Rows) < 60 {
			t.add(fmt.Sprintf("second %d", s), int(bw), Ladder[step], int(buffer*10))
		}
	}
	r.say("average bitrate %d kbps over %d seconds, %d quality switches, %d stalls", sum/c.Seconds, c.Seconds, switches, stalls)
	r.say("the player trades quality for smoothness: a lower step when the buffer is thin")
	r.put("average bitrate (kbps)", float64(sum/c.Seconds))
	r.put("switches", float64(switches))
	r.put("stalls", float64(stalls))
	return r, nil
}
