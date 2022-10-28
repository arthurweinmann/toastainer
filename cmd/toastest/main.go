package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/alexflint/go-arg"
	"github.com/toastate/toastainer/internal/config"
	"github.com/toastate/toastainer/internal/utils"
	"github.com/toastate/toastainer/test/library"
)

var args struct {
	Run  *RunCmd `arg:"subcommand:run" help:"run tests"`
	Home string  `arg:"-h" help:"Default is ~/.toastainer"`
}

type RunCmd struct {
}

func main() {
	arg.MustParse(&args)

	if args.Home == "" {
		log.Fatal("you must provide toastainer binary home folder")
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

	config.LogLevel = "all"
	utils.InitLogging(config.LogLevel)

	switch {
	case args.Run != nil:
		fat, err := library.NewFullAPITest(func() *http.Client {
			return &http.Client{}
		}, "https://"+config.APIDomain, config.APIDomain, config.ToasterDomain, config.DashboardDomain)
		if err != nil {
			log.Fatal("could not setup full api test", err)
		}

		err = fat.Run()
		if err != nil {
			log.Fatal("full api test error:", err)
		}

		log.Println("All tests passed successfully")

	default:
		log.Fatal("no valid command provided")
	}
}
