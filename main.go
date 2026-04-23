package main

import (
	"context"
	"errors"
	"flag"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"

	"mercantile/api"
	"mercantile/tui"
)

func main() {
	var (
		host     = flag.String("host", "0.0.0.0", "SSH bind host")
		port     = flag.String("port", "2222", "SSH bind port")
		hostKey  = flag.String("host-key", ".ssh/mercantile_ed25519", "path to host key (generated if missing)")
		baseURL  = flag.String("api", "https://mercantile.wordpress.org/wp-json/wc/store/v1", "Store API base URL")
		siteURL  = flag.String("site", "https://mercantile.wordpress.org", "Website URL (used for checkout handoff)")
		local    = flag.Bool("local", false, "run the TUI locally instead of over SSH (dev convenience)")
	)
	flag.Parse()

	if *local {
		runLocal(*baseURL, *siteURL)
		return
	}

	if err := os.MkdirAll(".ssh", 0o700); err != nil {
		log.Fatal("mkdir .ssh", "err", err)
	}

	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(*host, *port)),
		wish.WithHostKeyPath(*hostKey),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler(*baseURL, *siteURL)),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Fatal("could not start server", "err", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	log.Info("starting mercantile ssh", "addr", net.JoinHostPort(*host, *port))
	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("server error", "err", err)
			done <- nil
		}
	}()

	<-done
	log.Info("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("shutdown", "err", err)
	}
}

func teaHandler(apiBase, siteURL string) bubbletea.Handler {
	return func(sess ssh.Session) (tea.Model, []tea.ProgramOption) {
		_, _, active := sess.Pty()
		if !active {
			wish.Fatalln(sess, "mercantile requires an interactive terminal (try: ssh -t ...)")
			return nil, nil
		}
		client := api.New(apiBase)
		m := tui.NewModel(client, siteURL)
		return m, []tea.ProgramOption{tea.WithAltScreen()}
	}
}

func runLocal(apiBase, siteURL string) {
	client := api.New(apiBase)
	m := tui.NewModel(client, siteURL)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal("run", "err", err)
	}
}
