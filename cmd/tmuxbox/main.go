// Command tmuxbox shows the panes of your tmux server on a bean in the
// browser, and connects a terminal in the page to the one you pick.
//
// It listens on this machine only, and every request for pane contents or a
// terminal must carry the token printed in the address it opens.
package main

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/vinodhalaharvi/tmuxbox/web"
)

// Version is what -version prints.
const Version = "v0.1.1"

func main() {
	port := flag.Int("port", 0, "port to listen on, on this machine only; 0 picks a free one")
	noOpen := flag.Bool("no-open", false, "print the address and do not open a browser")
	version := flag.Bool("version", false, "print the version and stop")
	flag.Parse()
	if *version {
		fmt.Println("tmuxbox", Version)
		return
	}
	if _, err := exec.LookPath("tmux"); err != nil {
		fmt.Fprintln(os.Stderr, "tmuxbox: tmux is not installed or not on PATH")
		os.Exit(1)
	}

	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", *port))
	if err != nil {
		fmt.Fprintln(os.Stderr, "tmuxbox:", err)
		os.Exit(1)
	}
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		fmt.Fprintln(os.Stderr, "tmuxbox:", err)
		os.Exit(1)
	}
	s := &server{token: hex.EncodeToString(raw), port: ln.Addr().(*net.TCPAddr).Port}
	url := fmt.Sprintf("http://127.0.0.1:%d/?t=%s", s.port, s.token)
	fmt.Println("tmuxbox is at", url)
	fmt.Println("Ctrl-C stops it.")
	if !*noOpen {
		openBrowser(url)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stop
		killGroups()
		os.Exit(0)
	}()
	if err := http.Serve(ln, s.handler()); err != nil {
		fmt.Fprintln(os.Stderr, "tmuxbox:", err)
		os.Exit(1)
	}
}

type server struct {
	token string
	port  int
}

func (s *server) handler() http.Handler {
	files, err := fs.Sub(web.Files, ".")
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServerFS(files))
	mux.HandleFunc("GET /api/panes", s.guard(servePanes))
	mux.HandleFunc("POST /api/native", s.guard(serveNative))
	mux.HandleFunc("GET /ws", s.guard(serveTerm))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// a page on another site cannot reach this server under a name of
		// its own choosing
		if h := r.Host; h != fmt.Sprintf("127.0.0.1:%d", s.port) && h != fmt.Sprintf("localhost:%d", s.port) {
			http.Error(w, "wrong host", http.StatusForbidden)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		mux.ServeHTTP(w, r)
	})
}

// guard lets a request through only with the token, and only from this
// server's own page.
func (s *server) guard(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if subtle.ConstantTimeCompare([]byte(r.URL.Query().Get("t")), []byte(s.token)) != 1 {
			http.Error(w, "missing or wrong token", http.StatusForbidden)
			return
		}
		if o := r.Header.Get("Origin"); o != "" && o != "http://"+r.Host {
			http.Error(w, "wrong origin", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

func servePanes(w http.ResponseWriter, _ *http.Request) {
	var out struct {
		Panes []Pane `json:"panes"`
		Total int    `json:"total"`
		Error string `json:"error,omitempty"`
	}
	panes, total, err := listPanes()
	if err != nil {
		out.Error = err.Error()
	}
	for i := range panes {
		panes[i].Lines, panes[i].Ansi, panes[i].Alt = capture(panes[i].ID)
	}
	out.Panes, out.Total = panes, total
	if out.Panes == nil {
		out.Panes = []Pane{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

// serveNative brings a pane up in a terminal application of the machine's
// own. A session someone is already attached to only needs the pane selected.
func serveNative(w http.ResponseWriter, r *http.Request) {
	p, err := findPane(r.URL.Query().Get("pane"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	tmux("select-window", "-t", p.window)
	tmux("select-pane", "-t", p.ID)
	w.Header().Set("Content-Type", "application/json")
	if p.attached {
		json.NewEncoder(w).Encode(map[string]string{"did": "selected the pane in the terminal already attached to " + p.Key})
		return
	}
	// the session id is $ and digits, checked when the pane was listed
	attach := "tmux attach-session -t '" + p.session + "'"
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("osascript", "-e", `tell application "Terminal" to do script "`+attach+`"`, "-e", `tell application "Terminal" to activate`)
	default:
		cmd = exec.Command("x-terminal-emulator", "-e", "sh", "-c", attach)
	}
	if err := cmd.Start(); err != nil {
		json.NewEncoder(w).Encode(map[string]string{"did": "could not start a terminal; run: " + attach})
		return
	}
	go cmd.Wait()
	json.NewEncoder(w).Encode(map[string]string{"did": "opened " + p.Key + " in a terminal window"})
}

func openBrowser(url string) {
	name := "xdg-open"
	if runtime.GOOS == "darwin" {
		name = "open"
	}
	if err := exec.Command(name, url).Start(); err != nil {
		fmt.Println("Open that address in a browser.")
	}
}
