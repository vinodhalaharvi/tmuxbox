# tmuxbox

Your tmux panes, wrapped around a bean in the browser. Type to find one, press
Enter, and you are in it: a terminal in the page, joined to the real pane.

A dynamic program decides where every pane goes and how big it is, and gives
more room to the panes you open most.

## Install

```
brew install vinodhalaharvi/tap/tmuxbox
```

Or from a checkout, with Go 1.24 or newer:

```
make build
```

## Run

With tmux running:

```
tmuxbox
```

It prints an address on this machine, opens it in your browser, and runs until
you stop it with Ctrl-C.

| Flag | What it does |
|---|---|
| `-port N` | listen on port N instead of a free one |
| `-no-open` | print the address and do not open a browser |
| `-version` | print the version |

## Keys

The screen shows the bean and the text you type, and nothing else.

| Key | What it does |
|---|---|
| type | turn to the pane that matches best, by name or by what it shows |
| Enter | open that pane as a terminal in the page |
| Ctrl ] | leave the terminal, back to the bean |
| Shift Enter | open the pane in a terminal window of your own |
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

## What it does to your tmux

- It reads every pane once a second with `tmux capture-pane`, to draw it.
- Enter makes a session grouped with the pane's own, named `tmuxbox-…`,
  selects the pane, and attaches a terminal in the page to it. Nothing you
  have attached is detached. Leaving the terminal, or stopping tmuxbox,
  removes that session.
- A browser terminal is one more client. tmux sizes a window to the client
  that used it last, so a window you also have open elsewhere changes size
  while the page is on it.
- Shift Enter selects the pane and, when no terminal is attached to its
  session, opens one: Terminal.app on macOS.
- The bean shows ten panes. With more, it shows those in the ten most
  recently active windows.

## Who can reach it

tmuxbox gives a web page a shell, so it is strict about who may ask:

- it listens on 127.0.0.1 only;
- reading panes and opening a terminal need a token, made new on every run
  and carried in the address it prints;
- it refuses a request from another site's page, and one that names the
  machine by anything but `127.0.0.1` or `localhost`.

Anyone who can read that address while tmuxbox runs can use your panes. It is
in your shell's output and your browser's history.

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
looks an entry up; only a change of panes or of weights solves again. Ten
panes is 2,045 states and 67,743 alternatives, a few seconds; six is 125
states and immediate. The page runs the solver as WebAssembly in a worker.

`make check` compares the program with brute force on a five-pane case, every
whole layout one by one, and prints its timing on ten panes.

## The recorded demo

The page also works with no tmuxbox behind it. Then it shows ten panes of
output recorded from tmux, and Enter opens a pane as a small Linux machine
emulated in the page.

```
make demo
```

serves it on http://127.0.0.1:8080, after downloading the emulator and a
Linux disc.

## Layout of this repository

```
cmd/tmuxbox/     the command: reads tmux, serves the page, joins terminals to panes
cmd/beansolve/   the solver: solve.go, a WebAssembly entry point, a native check
web/             the page, and what it loads
scripts/         fetch-third-party.sh, for the demo
```

## Third-party code

In the repository, and inside the tmuxbox binary:

| File | From | Licence |
|---|---|---|
| `web/lib/three.min.js` | three.js r128 | MIT, `web/lib/three.LICENSE` |
| `web/lib/xterm/` | xterm.js 5.5.0 and its fit addon 0.10.0 | MIT, `web/lib/xterm/LICENSE` |

Copied in when you build:

| File | From | Licence |
|---|---|---|
| `web/lib/wasm_exec.js` | your Go installation | BSD 3-Clause |

Downloaded by `make demo` only, at pinned versions checked against recorded
SHA-256 sums, and never kept in the repository or the binary:

| File | From | Licence |
|---|---|---|
| v86 emulator | npm `v86@0.5.473` | BSD 2-Clause |
| SeaBIOS and VGA BIOS | [copy/v86](https://github.com/copy/v86), `bios/` | LGPL 3 |
| Linux 2.6 with BusyBox | [copy/images](https://github.com/copy/images), `linux.iso` | GPL 2 |
