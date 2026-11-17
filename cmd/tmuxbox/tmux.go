package main

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// groupPrefix names the sessions tmuxbox makes for its own terminals.
const groupPrefix = "tmuxbox-"

// maxPanes is how many panes the bean shows: the layout is solved over every
// way of splitting them, which grows fast.
const maxPanes = 10

// Pane is one tmux pane as the page sees it.
type Pane struct {
	ID       string   `json:"id"`   // %12
	Key      string   `json:"key"`  // work:1.0, stable while the pane keeps its place
	Name     string   `json:"name"` // work:1.0 build vim
	W        int      `json:"w"`
	H        int      `json:"h"`
	Lines    []string `json:"lines"`
	Ansi     []string `json:"ansi,omitempty"`
	Alt      bool     `json:"alt,omitempty"`
	session  string   // $3
	window   string   // @7
	activity int64
	attached bool
}

var (
	paneID    = regexp.MustCompile(`^%[0-9]+$`)
	sessionID = regexp.MustCompile(`^\$[0-9]+$`)
	windowID  = regexp.MustCompile(`^@[0-9]+$`)
)

func tmux(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "tmux", args...).CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return "", errors.New(msg)
	}
	return string(out), nil
}

// tmux turns a tab in a format into an underscore, so the fields are split on
// spaces: the ones that cannot hold a space first, the session's name last.
const paneFormat = "#{pane_id} #{session_id} #{window_id} #{window_index} #{pane_index} #{pane_width} #{pane_height} " +
	"#{window_activity} #{session_attached} #{pane_current_command} #{session_name}"

// listPanes returns every pane once, without the sessions tmuxbox made
// itself, and how many there are in all. With more than maxPanes it keeps
// those in the most recently active windows.
func listPanes() ([]Pane, int, error) {
	out, err := tmux("list-panes", "-a", "-F", paneFormat)
	if err != nil {
		return nil, 0, err
	}
	// a window's name can hold spaces too, so it is read on its own
	names := map[string]string{}
	if wins, err := tmux("list-windows", "-a", "-F", "#{window_id} #{window_name}"); err == nil {
		for _, line := range strings.Split(strings.TrimRight(wins, "\n"), "\n") {
			if id, name, ok := strings.Cut(line, " "); ok {
				names[id] = name
			}
		}
	}
	seen := map[string]bool{}
	var panes []Pane
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		f := strings.SplitN(line, " ", 11)
		if len(f) != 11 || !paneID.MatchString(f[0]) || !sessionID.MatchString(f[1]) || !windowID.MatchString(f[2]) {
			continue
		}
		name := f[10]
		if strings.HasPrefix(name, groupPrefix) || seen[f[0]] {
			continue
		}
		seen[f[0]] = true
		w, _ := strconv.Atoi(f[5])
		h, _ := strconv.Atoi(f[6])
		act, _ := strconv.ParseInt(f[7], 10, 64)
		key := fmt.Sprintf("%s:%s.%s", name, f[3], f[4])
		label := key
		if n := names[f[2]]; n != "" {
			label += " " + n
		}
		if f[9] != names[f[2]] {
			label += " " + f[9]
		}
		panes = append(panes, Pane{ID: f[0], Key: key, Name: label, W: w, H: h,
			session: f[1], window: f[2], activity: act, attached: f[8] != "0"})
	}
	total := len(panes)
	if total > maxPanes {
		sort.SliceStable(panes, func(i, j int) bool { return panes[i].activity > panes[j].activity })
		panes = panes[:maxPanes]
	}
	sort.SliceStable(panes, func(i, j int) bool { return panes[i].Key < panes[j].Key })
	return panes, total, nil
}

// escapes matches the terminal control sequences tmux leaves in a capture
// taken with colours.
var escapes = regexp.MustCompile(`\x1b\[[0-9;:?]*[A-Za-z]|\x1b[()][0-9A-Za-z]|[\x0e\x0f]`)

// capture reads what a pane shows now: the plain text of each line, the same
// lines with their colour sequences when the pane has any, and whether a
// full-screen program has the pane.
func capture(id string) (plain, ansi []string, alt bool) {
	out, err := tmux("capture-pane", "-p", "-e", "-t", id)
	if err != nil {
		return nil, nil, false
	}
	ansi = strings.Split(out, "\n")
	plain = make([]string, len(ansi))
	coloured := false
	for i := range ansi {
		plain[i] = strings.TrimRight(escapes.ReplaceAllString(ansi[i], ""), " ")
		if plain[i] != strings.TrimRight(ansi[i], " ") {
			coloured = true
		}
	}
	n := len(plain)
	for n > 0 && plain[n-1] == "" {
		n--
	}
	plain, ansi = plain[:n], ansi[:n]
	if !coloured {
		ansi = nil
	}
	if on, err := tmux("display-message", "-p", "-t", id, "#{alternate_on}"); err == nil {
		alt = strings.TrimSpace(on) == "1"
	}
	return plain, ansi, alt
}

func findPane(id string) (Pane, error) {
	if !paneID.MatchString(id) {
		return Pane{}, errors.New("not a pane id")
	}
	panes, _, err := listPanes()
	if err != nil {
		return Pane{}, err
	}
	for _, p := range panes {
		if p.ID == id {
			return p, nil
		}
	}
	return Pane{}, errors.New("no such pane")
}
