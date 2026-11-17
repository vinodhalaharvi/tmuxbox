# Changelog

## v0.1.1 - 2026-11-17

- Panes on the bean keep their colours, so htop, coloured prompts and
  highlighted output look as they do in the terminal.
- A pane running a full-screen program is shown from its top line, not its
  last.

## v0.1.0 - 2026-11-17

- `tmuxbox` shows the panes of your tmux server on a bean in the browser,
  read once a second.
- Type to turn to the pane that matches best; Enter opens a terminal in the
  page, joined to that pane; Ctrl ] returns. Shift Enter opens the pane in a
  terminal window of your own.
- A dynamic program decides which panes go on the front and back, how each
  half is cut, and how big each pane is, and gives more room to the panes you
  open most.
- It listens on 127.0.0.1 only and asks for a token, new on every run, before
  it shows a pane or opens a terminal.

Tested on Linux with tmux 3.4. Not run on macOS.
