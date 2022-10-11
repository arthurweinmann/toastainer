package supervisor

import (
	"fmt"

	"github.com/toastate/toastainer/internal/backgroundtasks"
	"github.com/toastate/toastainer/internal/config"
	"github.com/toastate/toastainer/internal/db/objectdb"
	"github.com/toastate/toastainer/internal/db/objectstorage"
	"github.com/toastate/toastainer/internal/db/redisdb"
	"github.com/toastate/toastainer/internal/email"
	"github.com/toastate/toastainer/internal/nodes"
	"github.com/toastate/toastainer/internal/runner"
	"github.com/toastate/toastainer/internal/utils"
)

func Start() (*Watcher, error) {
	err := nodes.Init()
	if err != nil {
		return nil, err
	}

	err = redisdb.Init()
	if err != nil {
		return nil, err
	}

	if config.IsAPI {
		err = objectdb.Init()
		if err != nil {
			return nil, err
		}
	}

	err = objectstorage.Init()
	if err != nil {
		return nil, err
	}

	if config.EmailProvider.Name != "" {
		err = email.Init()
		if err != nil {
			return nil, err
		}
	}

	if config.IsRunner {
		err = runner.Init()
		if err != nil {
			return nil, err
		}
	}

	err = backgroundtasks.Init()
	if err != nil {
		return nil, err
	}

	var srv *httpservers
	if config.IsAPI {
		// acme.init is in startServer because we need the http server to be running for HTTP Challenges
		// this needs to be at the end of initialization to take every dynamic routes into account
		srv, err = startServer()
		if err != nil {
			return nil, err
		}
	}

	wat := startWatcher(srv)

	fmt.Println("Toastainer is running..")

	return wat, nil
}

func StartNoServer() (*Watcher, error) {
	err := nodes.Init()
	if err != nil {
		return nil, err
	}

	err = redisdb.Init()
	if err != nil {
		return nil, err
	}

	if config.IsAPI {
		err = objectdb.Init()
		if err != nil {
			return nil, err
		}
	}

	err = objectstorage.Init()
	if err != nil {
		return nil, err
	}

	if config.EmailProvider.Name != "" {
		err = email.Init()
		if err != nil {
			return nil, err
		}
	}

	if config.IsRunner {
		err = runner.Init()
		if err != nil {
			return nil, err
		}
	}

	err = backgroundtasks.Init()
	if err != nil {
		return nil, err
	}

	wat := startWatcher(nil)

	utils.Info("msg", "Toastainer is running..")

	return wat, nil
}
