package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/alexflint/go-arg"
	"github.com/mitchellh/go-homedir"
	"github.com/toastate/toastcloud/internal/config"
	"github.com/toastate/toastcloud/internal/db/objectdb"
	"github.com/toastate/toastcloud/internal/db/objectstorage"
	"github.com/toastate/toastcloud/internal/db/redisdb"
	"github.com/toastate/toastcloud/internal/email"
	"github.com/toastate/toastcloud/internal/nodes"
	"github.com/toastate/toastcloud/internal/runner"
	"github.com/toastate/toastcloud/internal/utils"
)

var args struct {
	Start *StartCmd `arg:"subcommand:start" help:"start toastcloud server"`
	Quiet bool      `arg:"-q" help:"turn off logging"`
	Home  string    `arg:"-h" help:"Default is ~/.toastcloud"`
}

type StartCmd struct {
}

func main() {
	arg.MustParse(&args)

	if args.Home == "" {
		d, err := homedir.Dir()
		if err != nil {
			log.Fatal(err)
		}

		args.Home = filepath.Join(d, ".toastcloud")
		log.Println("Home directory is", args.Home)

		err = os.MkdirAll(args.Home, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	var configFile string

	if utils.FileExists(filepath.Join(args.Home, "config.json")) {
		configFile = filepath.Join(args.Home, "config.json")
	} else if utils.FileExists(filepath.Join(args.Home, "config.yml")) {
		configFile = filepath.Join(args.Home, "config.yml")
	}

	if configFile == "" {
		configFile = filepath.Join(args.Home, "config.json")

		b, err := json.Marshal(config.DefaultConfig())
		if err != nil {
			log.Fatal("could not marshall default configuration", err)
		}

		err = os.WriteFile(configFile, b, 0644)
		if err != nil {
			log.Fatal("could not write default configuration", err)
		}
	}

	err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatal("could not load configuration", err)
	}

	config.Home = args.Home

	if len(config.DNSProvider.ENV) > 0 {
		for k, v := range config.DNSProvider.ENV {
			err = os.Setenv(k, v)
			if err != nil {
				log.Fatal("could not setenv ", k, "=", v, err)
			}
		}
	}

	utils.InitLogging()

	switch {
	case args.Start != nil:
		err = nodes.Init()
		if err != nil {
			log.Fatal("could not initialize nodes", err)
		}

		err = redisdb.Init()
		if err != nil {
			log.Fatal("could not initialize Redis", err)
		}

		if config.IsAPI {
			err = objectdb.Init()
			if err != nil {
				log.Fatal("could not initialize SQL Database", err)
			}
		}

		err = objectstorage.Init()
		if err != nil {
			log.Fatal("could not initialize SQL Database", err)
		}

		err = email.Init()
		if err != nil {
			log.Fatal("could not initialize Email client", err)
		}

		if config.IsRunner {
			err = runner.Init()
			if err != nil {
				log.Fatal("could not start runner", err)
			}
		}

		if config.IsAPI {
			// acme.init is in startServer because we need the http server to be running for HTTP Challenges
			// this needs to be at the end of initialization to take every dynamic routes into account
			_, err = startServer()
			if err != nil {
				log.Fatal("could not start api server", err)
			}
		}

	default:
		log.Fatal("you must provide a command")
	}
}
