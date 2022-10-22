package hound

import (
	"fmt"
	authplugin "github.com/parasource/hound/sdk/plugins/auth"
	"github.com/rs/zerolog/log"
	"net"
	"sync"
)

type Config struct {
	Address     string
	MasterToken string
}

type Hound struct {
	cfg Config

	mu          sync.RWMutex
	shutdown    bool
	authModules map[string]authplugin.Auth
	lis         net.Listener
	shutdownC   chan struct{}
}

func New(cfg Config) (*Hound, error) {
	h := &Hound{
		cfg:         cfg,
		authModules: make(map[string]authplugin.Auth),

		shutdownC: make(chan struct{}),
	}

	lis, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return nil, err
	}

	h.lis = lis

	return h, err
}

func (h *Hound) NotifyShutdown() <-chan struct{} {
	return h.shutdownC
}

func (h *Hound) Shutdown() {
	h.mu.RLock()
	if h.shutdown {
		h.mu.RUnlock()
		return
	}
	h.mu.RUnlock()

	close(h.shutdownC)

	err := h.lis.Close()
	if err != nil {
		log.Err(fmt.Errorf("error closing net listener: %w", err))
	}

	log.Info().Msg("Hound server is closed")

	h.mu.Lock()
	h.shutdown = true
	h.mu.Unlock()
}
