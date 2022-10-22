package main

import (
	"github.com/parasource/hound/hound"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

var configDefaults = map[string]interface{}{
	"address":          "127.0.0.1:51820",
	"shutdown_timeout": 30,
}

func init() {
	rootCmd.Flags().String("address", "127.0.0.1:1723", "address")
	rootCmd.Flags().Int("shutdown_timeout", 30, "shutdown timeout")

	viper.BindPFlag("address", rootCmd.Flags().Lookup("http_host"))
	viper.BindPFlag("shutdown_timeout", rootCmd.Flags().Lookup("shutdown_timeout"))
}

var rootCmd = &cobra.Command{
	Use: "hound",
	Run: func(cmd *cobra.Command, args []string) {
		for k, v := range configDefaults {
			viper.SetDefault(k, v)
		}

		bindEnvs := []string{
			"address", "shutdown_timeout",
		}
		for _, env := range bindEnvs {
			err := viper.BindEnv(env)
			if err != nil {
				panic(err)
			}
		}

		if os.Getenv("GOMAXPROCS") == "" {
			if viper.IsSet("gomaxprocs") && viper.GetInt("gomaxprocs") > 0 {
				runtime.GOMAXPROCS(viper.GetInt("gomaxprocs"))
			} else {
				runtime.GOMAXPROCS(runtime.NumCPU())
			}
		}

		v := viper.GetViper()

		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

		address := v.GetString("address")

		cfg := hound.Config{
			Address:     address,
			MasterToken: "secret",
		}

		h, err := hound.New(cfg)
		if err != nil {
			panic(err)
		}

		log.Info().Str("address", address).Msg("Hound server successfully started")

		shutdownC := make(chan struct{})
		go handleSignals(shutdownC)

		select {
		case <-shutdownC:
			h.Shutdown()
		}
	},
}

func handleSignals(shutdownCh chan<- struct{}) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGHUP, syscall.SIGINT, os.Interrupt, syscall.SIGTERM)
	for {
		sig := <-sigc
		switch sig {
		case syscall.SIGHUP:

		case syscall.SIGINT, os.Interrupt, syscall.SIGTERM:

			log.Info().Msg("shutdown signal received")

			pidFile := viper.GetString("pid_file")
			shutdownTimeout := time.Duration(viper.GetInt("shutdown_timeout")) * time.Second

			close(shutdownCh)

			go time.AfterFunc(shutdownTimeout, func() {
				if pidFile != "" {
					os.Remove(pidFile)
				}
				os.Exit(1)
			})
		}
	}
}
