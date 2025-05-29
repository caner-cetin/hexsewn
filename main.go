package main

import (
	"bytes"
	"context"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
	"golang.design/x/clipboard"
	"net/url"
	"os"
	"strings"
)

var ctx = context.Background()

func init() {
	pflag.CountVarP(&cfg.Verbosity, "verbose", "v", "verbosity level (-v info, -vv debug, -vvv (and further) trace, default level warn.")
	pflag.StringVarP(&cfg.ConfigPath, "config", "c", ".hexsewn.toml", "path to .hexsewn.toml")
	pflag.BoolVarP(&cfg.PrettyLogs, "pretty", "p", false, "pretty print logs, if given, logs will be pretty printed with a human-friendly format, with colors. (default: false, json structured logs)")
	pflag.Parse()

	if cfg.PrettyLogs {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	if err := cfg.Read(); err != nil {
		log.Error().Err(err).Msg("failed to read config")
		return
	}
	go cfg.Watch(ctx)
	switch cfg.Verbosity {
	case 0:
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case 1:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case 2:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	}
	if err := clipboard.Init(); err != nil {
		log.Error().Err(err).Msg("failed to initialize clipboard")
		return
	}
}

func main() {
	log.Info().Msg("watching clipboard")
	ch := clipboard.Watch(ctx, clipboard.FmtText)
	var lastWritten []byte
	for dataBytes := range ch {
		if !bytes.Equal(lastWritten, dataBytes) {
			data := string(dataBytes)
			uri, err := url.Parse(data)
			if err != nil {
				if len(data) > 60 {
					data = data[:60]
				}
				log.Debug().Str("url", data).Msg("skipping invalid URL")
				continue
			}
			domain, redirectTo, found := cfg.RedirectConfig.DetectDomain(uri)
			if found {
				lastWritten = []byte(strings.Replace(data, domain, redirectTo, 1))
				clipboard.Write(clipboard.FmtText, lastWritten)
			}
		}
	}
}
