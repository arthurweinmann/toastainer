package supervisor

import (
	"os"
	"os/signal"
	"sync/atomic"

	"github.com/toastate/toastainer/internal/config"
	"github.com/toastate/toastainer/internal/runner"
	"github.com/toastate/toastainer/internal/utils"
)

type Watcher struct {
	shuttingdown      uint32    // set to 1 atomically when shutdown is triggered
	shuttingdownEvent chan bool // closed when shutdown is triggered

	shutdownDone chan bool

	sigchan chan os.Signal

	srv *httpservers
}

func startWatcher(srv *httpservers) *Watcher {
	wat := &Watcher{
		srv: srv,
	}

	wat.shuttingdownEvent = make(chan bool)
	wat.shutdownDone = make(chan bool)
	wat.sigchan = make(chan os.Signal, 1)

	signal.Notify(wat.sigchan, os.Interrupt)

	if config.IsAPI && wat.srv != nil {
		go wat.httpWatcher()
	}
	go wat.signalWatcher()

	return wat
}

func (wat *Watcher) WaitForShutdown() {
	<-wat.shutdownDone
}

func (wat *Watcher) Shutdown() {
	if atomic.CompareAndSwapUint32(&wat.shuttingdown, 0, 1) {
		utils.Info("msg", "Toastainer is shutting down..")

		close(wat.shuttingdownEvent)

		if wat.srv != nil {
			wat.srv.Close()
		}

		if config.IsRunner {
			runner.Stop()
		}

		close(wat.shutdownDone)
	}
}

func (wat *Watcher) httpWatcher() {
	err := <-wat.srv.errs
	if err != nil && atomic.LoadUint32(&wat.shuttingdown) == 0 {
		utils.Error("msg", "API Server", err)
	}
	wat.Shutdown()
}

func (wat *Watcher) signalWatcher() {
	<-wat.sigchan
	wat.Shutdown()
}
