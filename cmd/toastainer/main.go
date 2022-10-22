package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/alexflint/go-arg"
	"github.com/mitchellh/go-homedir"
	"github.com/toastate/toastainer/internal/config"
	"github.com/toastate/toastainer/internal/supervisor"
	"github.com/toastate/toastainer/internal/utils"
)

var args struct {
	Start            *StartCmd            `arg:"subcommand:start" help:"start toastainer server"`
	GenConfigExample *GenConfigExampleCmd `arg:"subcommand:configexpl" help:"Generate a configuration file example"`

	Quiet bool   `arg:"-q" help:"turn off logging"`
	Home  string `arg:"-h" help:"Default is ~/.toastainer"`
}

type StartCmd struct {
}

type GenConfigExampleCmd struct {
	Path string `arg:"-p" help:"Either a JSON or YAML filepath"`
}

func main() {
	arg.MustParse(&args)

	switch {
	case args.GenConfigExample != nil:
		if args.GenConfigExample.Path == "" {
			log.Fatal("you must provide a path for your configuration example file")
		}

		b, err := json.Marshal(config.DefaultConfig())
		if err != nil {
			log.Fatal("could not marshall configuration example", err)
		}

		err = os.WriteFile(args.GenConfigExample.Path, b, 0644)
		if err != nil {
			log.Fatal("could not write configuration example file", err)
		}

		return
	}

	if args.Home == "" {
		d, err := homedir.Dir()
		if err != nil {
			log.Fatal(err)
		}

		args.Home = filepath.Join(d, ".toastainer")
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
		log.Fatal("configuration:", err)
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

	if args.Quiet {
		config.LogLevel = "quiet"
	}
	utils.InitLogging(config.LogLevel)

	switch {
	case args.Start != nil:
		wat, err := supervisor.Start()
		if err != nil {
			log.Fatal("could not start supervisor: ", err)
		}
		wat.WaitForShutdown()

	default:
		log.Fatal("you must provide a command")
	}
}
