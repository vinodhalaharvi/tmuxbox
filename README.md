# tmuxbox

Terminal panes wrapped around a bean you can turn, resize and search, with a
dynamic program deciding where every pane goes and how big it is.

What is here is a browser demo. A program that attaches to the tmux on your
own machine is not written.

## Run it

Needs Go 1.24 or newer, `curl`, `openssl`, `tar` and `python3`.

```
make serve
```

Then open http://127.0.0.1:8080. `make serve` builds the solver to
WebAssembly, downloads the third-party files listed below, and serves the
`web` folder on localhost.

## Keys

The screen shows the bean and the text you type, and nothing else.

| Key | What it does |
|---|---|
| type | turn to the pane that matches best, by name or contents |
| Enter | open that pane as a real Linux terminal |
| Ctrl ] | leave the terminal, back to the bean |
| Tab, Shift Tab | next and previous pane |
| Esc | clear what you typed |
| arrows | turn the bean |
| ⌘ or Ctrl + ↑ ↓ | bigger, smaller |
| ⌘ or Ctrl + ← → | shorter, longer |
| Home | reset shape and view |
| ⌘ or Ctrl + ⌫ | forget usage |
| Alt Enter | ask the browser for full screen |
| F1 | keys, the solver's numbers, credits |

On a touch screen: drag to turn, pinch to resize, tap a pane to open it, tap
the text to type.

## What decides the layout

One dynamic program, in `cmd/beansolve`, written with
[Dyna](https://github.com/vinodhalaharvi/dyna) and the
[collage](https://github.com/vinodhalaharvi/collage) engine. It decides three
things together:

- which panes go on the front half of the bean and which on the back;
- how each half is cut into panes;
- how big each pane is.

Each pane has a weight: one plus the number of times it has been opened. The
program maximises the smallest share divided by its weight, so a pane used
more gets more room and none is starved.

Its answer is a table with one entry per bean proportion. Resizing the bean
looks an entry up; only a change of weights solves again. Ten panes is 2,045
states and 67,743 alternatives. The page runs the solver as WebAssembly in a
worker, which takes a few seconds.

`make check` compares the program with brute force on a five-pane case, every
whole layout one by one, and prints its timing on ten panes.

## What is real and what is not

- The search, the layout and the re-planning are real and run in the page.
- Until a pane is opened it shows output recorded from tmux: six windows of
  one session running [tilegen](https://github.com/vinodhalaharvi/tilegen).
  It does not update.
- An opened pane is a real Linux kernel with BusyBox on an emulated PC, in
  the page, one machine per pane and two kept running at a time. It is not
  your computer, has no network, and is not connected to tmux.
- Usage is remembered in the browser's local storage only.

## Layout of this repository

```
cmd/beansolve/   the solver: solve.go, a WebAssembly entry point, a native check
web/index.html   the page
web/lib/         the worker that runs the solver; built and fetched files land here
scripts/         fetch-third-party.sh
```

## Third-party files

None are kept in this repository. `scripts/fetch-third-party.sh` downloads
each from its publisher at a pinned version and checks it against a recorded
SHA-256.

| File | From | Licence |
|---|---|---|
| three.js r128 | npm `three@0.128.0` | MIT |
| v86 emulator | npm `v86@0.5.473` | BSD 2-Clause |
| SeaBIOS and VGA BIOS | [copy/v86](https://github.com/copy/v86), `bios/` | LGPL 3 |
| Linux 2.6 with BusyBox | [copy/images](https://github.com/copy/images), `linux.iso` | GPL 2 |
| `wasm_exec.js` | your Go installation | BSD 3-Clause |
