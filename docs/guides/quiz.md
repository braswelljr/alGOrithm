# Quiz: predict what happens next

The quiz stops a real run just before a step and asks what the algorithm does next. Every answer comes from the run's own events and output, never from a second implementation, so the quiz cannot disagree with the code.

## In the terminal

```bash
algorithm quiz bubble-sort                      # 5 questions on the example
algorithm quiz quick-sort 5 3 9 1 --count 3     # on your own values
algorithm quiz round-robin --seed 7             # the same seed asks the same questions
algorithm quiz lru-page-replacement --json      # the questions as JSON, without asking
```

Type a choice's number. After each answer it shows whether you were right, why, and a score line: points (100 times a streak multiplier), the streak, and a row of `+`, `x` and `.` for right, wrong and still to come. Color appears only on a terminal (and never with `NO_COLOR` set). The input ending stops the quiz with the score so far.

## What it asks

| The run | The question | Paused at |
|---------|--------------|-----------|
| A sort | Which two positions does it compare (or swap) next? | the event before it |
| A search | Which position does it look at next? | the event before it |
| A grid or a graph search | Which cell or node does it explore next? The other choices come from its frontier. | the event before it |
| CPU scheduling | At time t, which process runs? | the time t |
| Disk scheduling | Where does the head go next? The other choices are the waiting requests nearest the head. | the seek before it |
| Page replacement | Page p is referenced: a hit, a fault into an empty frame, or a fault evicting which page? | the reference |
| A data structure | What does `pop`, `peek`, `get 5` give back? | the operation before it |
| Anything else with a number | What is the result? The other choices are near it. | the end |

Each question has 3 or 4 different choices and one right answer, and `TestQuizOnEveryExample` checks that for every entry's example against a fresh run.

## On the site

The Quiz page (`/quiz/`) is a game built on the same questions, with the visualizer's own view of the run paused at each question's step. An answer plays that step on the view before the explanation appears.

- **Classic**: 10 questions, 3 lives.
- **Time attack**: as many as you can in 60 seconds.
- **Streak**: endless, until the first miss.

A right answer scores 100 times the streak multiplier (x1 for streaks of 1 to 3, x2 for 4 to 6, x3 for 7 to 9, x4 after) plus up to 50 for answering fast. Choose a topic and a difficulty, or press "Quiz me" on any entry's run or docs page for 5 questions on that entry. The link keeps the settings and the seed, so a shared link replays the same questions; best scores stay in your browser only. With reduced motion, nothing tweens and answers reveal at once, but the timers still count.

The rules are `gameReducer` in [`web/src/visualizer/core/game.ts`](../../web/src/visualizer/core/game.ts), a pure reducer tested transition by transition; the questions come from `core.Quiz` in [`internal/core/quiz.go`](../../internal/core/quiz.go).
