package main

import (
	"bytes"
	"context"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/pflag"
	"golang.design/x/clipboard"
	"log/slog"
	"net/url"
	"os"
	"strings"
)

var Verbosity int
var ConfigPath string
var cfg RootCfg

func init() {
	pflag.CountVarP(&Verbosity, "verbose", "v", "verbosity level (-v info -vv debug), default level warn.")
	pflag.StringVarP(&ConfigPath, "config", "c", ".hexsewn.toml", "path to .hexsewn.toml")
	pflag.Parse()

	cfgContents, err := os.ReadFile(ConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Error("config path does not exist", "path", ConfigPath)
		} else {
			slog.Error("cannot read config", "path", ConfigPath, "error", err)
		}
		os.Exit(1)
	}
	if err = toml.Unmarshal(cfgContents, &cfg); err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	switch Verbosity {
	case 0:
		slog.SetLogLoggerLevel(slog.LevelWarn)
	case 1:
		slog.SetLogLoggerLevel(slog.LevelInfo)
	case 2:
		slog.SetLogLoggerLevel(slog.LevelDebug)
	default:
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
	if err = clipboard.Init(); err != nil {
		slog.Error("failed to initialize clipboard", "error", err)
		os.Exit(1)
	}
}

func main() {
	slog.Info("watching clipboard...")
	ch := clipboard.Watch(context.Background(), clipboard.FmtText)
	var lastWritten []byte
	for dataBytes := range ch {
		if !bytes.Equal(lastWritten, dataBytes) {
			data := string(dataBytes)
			uri, err := url.Parse(data)
			if err != nil {
				if len(data) > 60 {
					data = data[:60]
				}
				slog.Debug("skipping invalid URL", "url", data)
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
