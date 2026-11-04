'use client';

import * as React from 'react';
import { BellIcon, InboxIcon, PlayIcon, SettingsIcon } from 'lucide-react';
import { CheatSheetTable } from '@braswelljr/algorithm-ui/algorithm/cheat-sheet-table';
import { CodeTabs } from '@braswelljr/algorithm-ui/algorithm/code-tabs';
import { ComplexityBadge } from '@braswelljr/algorithm-ui/algorithm/complexity-badge';
import { EntryCard } from '@braswelljr/algorithm-ui/algorithm/entry-card';
import { KeyCombo } from '@braswelljr/algorithm-ui/algorithm/key-combo';
import { LevelBadge } from '@braswelljr/algorithm-ui/algorithm/level-badge';
import { MorphingDialog } from '@braswelljr/algorithm-ui/algorithm/morphing-dialog';
import { NumberTicker } from '@braswelljr/algorithm-ui/algorithm/number-ticker';
import { PathStepper } from '@braswelljr/algorithm-ui/algorithm/path-stepper';
import { PlayerControls } from '@braswelljr/algorithm-ui/algorithm/player-controls';
import { QuizChoiceCard } from '@braswelljr/algorithm-ui/algorithm/quiz-choice-card';
import { SpringTabs } from '@braswelljr/algorithm-ui/algorithm/spring-tabs';
import { TagChip } from '@braswelljr/algorithm-ui/algorithm/tag-chip';
import { TextScramble } from '@braswelljr/algorithm-ui/algorithm/text-scramble';
import { TLDRCard } from '@braswelljr/algorithm-ui/algorithm/tldr-card';
import { TraceView } from '@braswelljr/algorithm-ui/algorithm/trace-view';
import { Alert, AlertDescription, AlertTitle } from '@braswelljr/algorithm-ui/components/alert';
import { Badge } from '@braswelljr/algorithm-ui/components/badge';
import { Button } from '@braswelljr/algorithm-ui/components/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@braswelljr/algorithm-ui/components/card';
import { Checkbox } from '@braswelljr/algorithm-ui/components/checkbox';
import { CodeSample } from '@braswelljr/algorithm-ui/components/code-block';
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@braswelljr/algorithm-ui/components/empty';
import { Field, FieldDescription, FieldLabel } from '@braswelljr/algorithm-ui/components/field';
import { Frame, FramePanel } from '@braswelljr/algorithm-ui/components/frame';
import { Heading } from '@braswelljr/algorithm-ui/components/heading';
import { Input } from '@braswelljr/algorithm-ui/components/input';
import { Item, ItemContent, ItemDescription, ItemGroup, ItemMedia, ItemTitle } from '@braswelljr/algorithm-ui/components/item';
import { Kbd } from '@braswelljr/algorithm-ui/components/kbd';
import { Marquee } from '@braswelljr/algorithm-ui/components/marquee';
import { Progress } from '@braswelljr/algorithm-ui/components/progress';
import { Questionnaire, QuestionnaireChoices, QuestionnaireItem, QuestionnaireTitle } from '@braswelljr/algorithm-ui/components/questionnaire';
import { Separator } from '@braswelljr/algorithm-ui/components/separator';
import { Skeleton } from '@braswelljr/algorithm-ui/components/skeleton';
import { Spinner } from '@braswelljr/algorithm-ui/components/spinner';
import { Stat, StatLabel, StatValue } from '@braswelljr/algorithm-ui/components/stat';
import { Switch } from '@braswelljr/algorithm-ui/components/switch';
import { ToggleGroup, ToggleGroupItem } from '@braswelljr/algorithm-ui/components/toggle-group';
import { Tooltip, TooltipContent, TooltipTrigger } from '@braswelljr/algorithm-ui/components/tooltip';
import { TypingTerminal } from '@braswelljr/algorithm-ui/components/typing-terminal';
import { duration, ease } from '@braswelljr/algorithm-ui/lib/motion';

/** A component shown twice: in a light region and a dark one. */
function Specimen({ title, children, wide }: { title: string; children: React.ReactNode; wide?: boolean }) {
  return (
    <section aria-label={title} className="flex flex-col gap-2">
      <Heading level={3} size="overline">
        {title}
      </Heading>
      <Frame className={wide ? 'grid gap-1' : 'grid gap-1 lg:grid-cols-2'}>
        {(['light', 'dark'] as const).map((m) => (
          <FramePanel
            key={m}
            className={`scope-${m} mt-0! flex min-w-0 flex-wrap items-start gap-3 bg-bg text-fg`}
          >
            <span className="sr-only">{m} theme: </span>
            {children}
          </FramePanel>
        ))}
      </Frame>
    </section>
  );
}

const events = [
  { kind: 'compare', indices: [0, 1], values: [5, 3] },
  { kind: 'swap', indices: [0, 1], values: [3, 5] },
  { kind: 'compare', indices: [1, 2], values: [5, 9] },
  { kind: 'mark', indices: [2], values: [], note: 'sorted' },
  { kind: 'set', indices: [3], values: [1] }
];

const notes = {
  tldr: 'Pick a pivot, put smaller values left and larger right, then sort each side.',
  keyIdea: 'After partitioning the pivot is in its final place.',
  whenToUse: 'General in-memory sorting where average speed matters.',
  pitfall: 'A bad pivot every time makes it O(n^2).'
};

/** The component gallery for /ui. */
export function Gallery() {
  const [tab, setTab] = React.useState('go');
  const [toggle, setToggle] = React.useState(['bars']);
  const [at, setAt] = React.useState(2);
  const [ticks, setTicks] = React.useState(128);

  return (
    <div className="flex flex-col gap-10">
      <section aria-labelledby="tokens" className="flex flex-col gap-4">
        <Heading level={2} id="tokens">
          Tokens
        </Heading>
        <div className="grid gap-4 md:grid-cols-2">
          <Card>
            <CardHeader>
              <CardTitle>Radius</CardTitle>
              <CardDescription>One scale; nested corners are the outer radius minus the padding. Squircles where supported.</CardDescription>
            </CardHeader>
            <CardContent className="flex flex-wrap items-end gap-3">
              {['rounded-sm', 'rounded-md', 'rounded-lg', 'rounded-xl', 'rounded-2xl'].map((r) => (
                <span key={r} className="flex flex-col items-center gap-1 text-xs text-muted">
                  <span className={`smooth-corners size-14 bg-accent/15 ring-1 ring-accent/40 ${r}`} />
                  {r.replace('rounded-', '')}
                </span>
              ))}
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>Elevation</CardTitle>
              <CardDescription>Layered, low-alpha shadows; dark mode leans on borders and tints.</CardDescription>
            </CardHeader>
            <CardContent className="flex flex-wrap items-end gap-4">
              {['shadow-xs', 'shadow-sm', 'shadow-md', 'shadow-lg', 'shadow-xl'].map((sh) => (
                <span key={sh} className="flex flex-col items-center gap-2 text-xs text-muted">
                  <span className={`smooth-corners size-14 rounded-xl bg-card ${sh}`} />
                  {sh.replace('shadow-', '')}
                </span>
              ))}
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>Motion</CardTitle>
              <CardDescription>One set of curves and durations for CSS, Motion and GSAP. Hover to play.</CardDescription>
            </CardHeader>
            <CardContent className="flex flex-col gap-2">
              {(
                [
                  ['ease-out-expo', ease.outExpo, 'var(--ease-out-expo)'],
                  ['ease-out-strong', ease.outStrong, 'var(--ease-out-strong)'],
                  ['ease-in-out-strong', ease.inOutStrong, 'var(--ease-in-out-strong)']
                ] as const
              ).map(([name, curve, v]) => (
                <div key={name} className="group flex items-center gap-3 text-xs">
                  <span className="w-32 font-mono text-muted">{name}</span>
                  <span className="relative h-2 flex-1 rounded-full bg-subtle">
                    <span
                      className="absolute top-0 left-0 size-2 rounded-full bg-accent transition-transform duration-(--duration-slow) group-hover:translate-x-[calc(100cqw)]"
                      style={{ transitionTimingFunction: v }}
                    />
                  </span>
                  <span className="font-mono text-muted">[{curve.join(', ')}]</span>
                </div>
              ))}
              <p className="text-xs text-muted">
                Durations: fast {duration.fast * 1000}ms, base {duration.base * 1000}ms, slow {duration.slow * 1000}ms.
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>Type</CardTitle>
              <CardDescription>Ysabeau Office for text, JetBrains Mono for code, both through next/font.</CardDescription>
            </CardHeader>
            <CardContent className="flex flex-col gap-1">
              <p className="text-2xl font-semibold">Sorting from zero</p>
              <p className="text-base">A humanist sans with a generous x-height, readable at small sizes.</p>
              <p className="font-mono text-sm">sorting.Quick([]int&#123;5, 3, 9, 1&#125;)</p>
              <p className="text-sm text-muted tabular-nums">0123456789 tabular figures for counters</p>
            </CardContent>
          </Card>
        </div>
      </section>

      <section aria-labelledby="primitives" className="flex flex-col gap-6">
        <Heading level={2} id="primitives">
          Primitives
        </Heading>
        <Specimen title="Button: one primary per view, ghost for the rest">
          <Button variant="primary">
            <PlayIcon aria-hidden /> Run
          </Button>
          <Button variant="ghost">
            <SettingsIcon aria-hidden /> Options
          </Button>
          <Button variant="secondary">Secondary</Button>
          <Button variant="link">Read the code</Button>
          <Button variant="destructive">Delete</Button>
          <Button variant="ghost" size="icon" aria-label="Notifications">
            <BellIcon />
          </Button>
        </Specimen>
        <Specimen title="Badge, LevelBadge, ComplexityBadge, TagChip">
          <Badge>default</Badge>
          <Badge variant="secondary">secondary</Badge>
          <LevelBadge level="beginner" />
          <LevelBadge level="intermediate" />
          <LevelBadge level="advanced" />
          <ComplexityBadge time="O(n log n)" space="O(n)" why="log n levels, each a pass over n values." />
          <TagChip tag="divide-and-conquer" />
        </Specimen>
        <Specimen title="Card variants">
          {(['default', 'ghost', 'interactive', 'selected', 'accent'] as const).map((v) => (
            <Card key={v} variant={v} size="sm" className="w-40 px-(--card-spacing)">
              <p className="font-medium">{v}</p>
              <p className="text-xs text-muted">Card variant</p>
            </Card>
          ))}
        </Specimen>
        <Specimen title="Field, Input, Checkbox, Switch, ToggleGroup">
          <Field className="w-56">
            <FieldLabel htmlFor="g-values">Values</FieldLabel>
            <Input id="g-values" placeholder="5 3 9 1" />
            <FieldDescription>Numbers separated by spaces.</FieldDescription>
          </Field>
          <label className="flex items-center gap-2 text-sm">
            <Checkbox defaultChecked /> Trace every step
          </label>
          <label className="flex items-center gap-2 text-sm">
            <Switch defaultChecked /> Reduced motion preview
          </label>
          <ToggleGroup variant="chip" value={toggle} onValueChange={(v) => setToggle(v)} aria-label="View">
            <ToggleGroupItem value="bars">Bars</ToggleGroupItem>
            <ToggleGroupItem value="cells">Cells</ToggleGroupItem>
          </ToggleGroup>
        </Specimen>
        <Specimen title="Tabs, Tooltip, Kbd and key combos">
          <SpringTabs
            label="Example"
            value={tab}
            onValueChange={setTab}
            tabs={[
              { value: 'go', label: 'Go' },
              { value: 'cli', label: 'CLI' },
              { value: 'json', label: 'JSON' }
            ]}
          />
          <Tooltip>
            <TooltipTrigger render={<Button variant="ghost" />}>Hover me</TooltipTrigger>
            <TooltipContent>A tooltip names or explains a control.</TooltipContent>
          </Tooltip>
          <Kbd>Esc</Kbd>
          <KeyCombo keys="mod+k" />
          <KeyCombo keys="g>l" />
        </Specimen>
        <Specimen title="Alert, Progress, Stat, Skeleton, Spinner">
          <Alert variant="info" className="w-72">
            <AlertTitle>Tip</AlertTitle>
            <AlertDescription>Press ? for every shortcut.</AlertDescription>
          </Alert>
          <Progress value={60} aria-label="60 percent" className="w-40" />
          <Stat className="w-40">
            <StatLabel>Comparisons</StatLabel>
            <StatValue>
              <NumberTicker value={ticks} />
            </StatValue>
          </Stat>
          <Button variant="ghost" onClick={() => setTicks((t) => t + Math.round(Math.random() * 50))}>
            Roll the number
          </Button>
          <Skeleton className="h-8 w-32" />
          <Spinner />
        </Specimen>
        <Specimen title="Empty, Item, Separator">
          <Empty className="w-64">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <InboxIcon />
              </EmptyMedia>
              <EmptyTitle>No runs yet</EmptyTitle>
              <EmptyDescription>Run an entry and it appears here.</EmptyDescription>
            </EmptyHeader>
          </Empty>
          <ItemGroup className="w-72">
            <Item variant="muted" size="sm">
              <ItemMedia>
                <PlayIcon aria-hidden />
              </ItemMedia>
              <ItemContent>
                <ItemTitle>Quick Sort</ItemTitle>
                <ItemDescription>Divide and conquer around a pivot.</ItemDescription>
              </ItemContent>
            </Item>
          </ItemGroup>
          <Separator className="w-full" />
        </Specimen>
      </section>

      <section aria-labelledby="ours" className="flex flex-col gap-6">
        <Heading level={2} id="ours">
          The algorithm components
        </Heading>
        <Specimen title="EntryCard (tilts toward the pointer)">
          <EntryCard
            name="Quick Sort"
            href="#"
            summary="Pick a pivot, partition around it, sort each side."
            level="intermediate"
            time="O(n log n)"
            why="log n levels, each a pass over n values."
            tags={['divide-and-conquer', 'in-place']}
            animated
            className="w-72"
          />
        </Specimen>
        <Specimen title="TLDRCard" wide>
          <TLDRCard notes={notes} time="O(n log n)" space="O(log n)" why="log n levels of n work each." />
        </Specimen>
        <Specimen title="PlayerControls and TraceView" wide>
          <PlayerControls
            playing={false}
            index={at}
            total={events.length}
            speed={1}
            onToggle={() => {}}
            onStep={(by) => setAt((i) => Math.min(events.length, Math.max(0, i + by)))}
            onSeek={setAt}
            onSpeed={() => {}}
            onReplay={() => setAt(0)}
            className="w-full"
          />
          <TraceView events={events} current={at - 1} onSeek={(i) => setAt(i + 1)} height={180} className="w-full" />
        </Specimen>
        <Specimen title="PathStepper">
          <PathStepper
            current={1}
            steps={[
              { title: 'Bubble Sort', note: 'Count the swaps on reversed input.', done: true },
              { title: 'Insertion Sort', note: 'Watch it fly on nearly sorted input.' },
              { title: 'Merge Sort', note: 'Count the levels: log n.' }
            ]}
          />
        </Specimen>
        <Specimen title="QuizChoiceCard on the Questionnaire">
          <Questionnaire shortcuts="numbers" onSubmit={(e) => e.preventDefault()} className="w-80">
            <QuestionnaireItem name="demo">
              <QuestionnaireTitle>Which two get swapped next?</QuestionnaireTitle>
              <QuestionnaireChoices>
                <QuizChoiceCard value="a" result="correct">[0] and [1]</QuizChoiceCard>
                <QuizChoiceCard value="b" result="wrong">[1] and [2]</QuizChoiceCard>
                <QuizChoiceCard value="c">[2] and [3]</QuizChoiceCard>
              </QuestionnaireChoices>
            </QuestionnaireItem>
          </Questionnaire>
        </Specimen>
        <Specimen title="CodeSample and CodeTabs" wide>
          <CodeSample
            title="main.go"
            language="go"
            lineNumbers
            highlight={[3]}
            code={'package main\n\nfunc main() { println(sorting.Quick([]int{5, 3, 9, 1})) }\n'}
          />
          <CodeTabs
            className="w-full"
            tabs={[
              { value: 'cli', label: 'CLI', language: 'bash', code: 'algorithm run quick-sort 5 3 9 1 --trace' },
              { value: 'json', label: 'JSON', language: 'json', code: '{ "values": [5, 3, 9, 1] }' }
            ]}
          />
        </Specimen>
        <Specimen title="CheatSheetTable" wide>
          <CheatSheetTable
            caption="Sorting"
            className="w-full"
            rows={[
              { id: 'a', name: 'Merge Sort', time: 'O(n log n)', space: 'O(n)', level: 'intermediate', keyIdea: 'Merging two sorted halves takes one pass.' },
              { id: 'b', name: 'Insertion Sort', time: 'O(n^2)', space: 'O(1)', level: 'beginner', keyIdea: 'Slide each value into the sorted prefix.' }
            ]}
          />
        </Specimen>
        <Specimen title="TextScramble, MorphingDialog, Marquee, TypingTerminal" wide>
          <TextScramble as="p" text="Sorting from zero" className="text-2xl font-semibold" />
          <MorphingDialog title="Quick Sort" trigger={<span className="font-medium">Open the details (it morphs)</span>}>
            <p className="text-sm text-muted">The card grows into this dialog, transform only.</p>
          </MorphingDialog>
          <Marquee pauseOnHover className="w-full">
            {['Sorting', 'Graphs', 'Cryptography', 'Scheduling', 'Machine learning'].map((t) => (
              <Badge key={t} variant="secondary">
                {t}
              </Badge>
            ))}
          </Marquee>
          <TypingTerminal
            label="The algorithm command running quick sort"
            title="Terminal"
            scenes={[{ command: 'algorithm run quick-sort 5 3 9 1', output: 'sorted  [1 3 5 9]' }]}
            className="w-full"
          />
        </Specimen>
      </section>
    </div>
  );
}
