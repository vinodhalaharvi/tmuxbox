package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"

	"github.com/coder/websocket"
	"github.com/creack/pty"
)

var (
	groupSeq atomic.Int64
	groupsMu sync.Mutex
	groups   = map[string]bool{}
)

// killGroups removes every session tmuxbox made, when it stops.
func killGroups() {
	groupsMu.Lock()
	defer groupsMu.Unlock()
	for g := range groups {
		tmux("kill-session", "-t", g)
	}
}

// serveTerm connects a browser terminal to a tmux pane.
//
// It makes a session grouped with the pane's own, so the browser has its own
// view of the same windows, selects the pane, and attaches to it on a pty.
// Bytes from the browser are keystrokes; a text message is a resize.
func serveTerm(w http.ResponseWriter, r *http.Request) {
	p, err := findPane(r.URL.Query().Get("pane"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	c, err := websocket.Accept(w, r, nil) // refuses another origin
	if err != nil {
		return
	}
	defer c.CloseNow()
	c.SetReadLimit(1 << 20)

	group := fmt.Sprintf("%s%d-%d", groupPrefix, os.Getpid(), groupSeq.Add(1))
	if _, err := tmux("new-session", "-d", "-s", group, "-t", p.session); err != nil {
		c.Close(websocket.StatusInternalError, "tmux: "+err.Error())
		return
	}
	groupsMu.Lock()
	groups[group] = true
	groupsMu.Unlock()
	defer func() {
		tmux("kill-session", "-t", group)
		groupsMu.Lock()
		delete(groups, group)
		groupsMu.Unlock()
	}()
	tmux("select-window", "-t", group+":"+p.window)
	tmux("select-pane", "-t", p.ID)

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	cmd := exec.CommandContext(ctx, "tmux", "attach-session", "-t", group)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")
	tty, err := pty.StartWithSize(cmd, &pty.Winsize{Cols: 80, Rows: 24})
	if err != nil {
		c.Close(websocket.StatusInternalError, "pty: "+err.Error())
		return
	}
	defer tty.Close()

	go func() { // the pane's output, to the browser
		defer cancel()
		buf := make([]byte, 32<<10)
		for {
			n, err := tty.Read(buf)
			if n > 0 {
				if c.Write(ctx, websocket.MessageBinary, buf[:n]) != nil {
					return
				}
			}
			if err != nil {
				return
			}
		}
	}()
	for { // keystrokes and resizes, from the browser
		typ, data, err := c.Read(ctx)
		if err != nil {
			break
		}
		if typ == websocket.MessageBinary {
			tty.Write(data)
			continue
		}
		var msg struct {
			Resize [2]uint16 `json:"resize"`
		}
		if json.Unmarshal(data, &msg) == nil && msg.Resize[0] > 0 && msg.Resize[1] > 0 {
			pty.Setsize(tty, &pty.Winsize{Cols: msg.Resize[0], Rows: msg.Resize[1]})
		}
	}
	cancel()
	cmd.Wait()
	c.Close(websocket.StatusNormalClosure, "")
}
