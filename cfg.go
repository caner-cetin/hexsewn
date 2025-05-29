package main

import (
	"context"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/pelletier/go-toml/v2"
	"log/slog"
	"net/url"
	"os"
	"reflect"
	"strings"
)

var cfg RootCfg

type RootCfg struct {
	RedirectConfig RedirectCfg `toml:"Redirect"`
	ConfigPath     string      `toml:"-"`
	Verbosity      int         `toml:"-"`
	PrettyLogs     bool        `toml:"-"`
}

type RedirectCfg struct {
	Twitter Redirect `toml:"Twitter"`
	Reddit  Redirect `toml:"Reddit"`
}

type Redirect struct {
	From []string `toml:"From"`
	To   string   `toml:"To"`
}

func (r *Redirect) FindDomain(uri *url.URL) (domain string, redirectTo string, found bool) {
	for _, h := range r.From {
		if strings.Contains(uri.Host, h) {
			return domain, r.To, found
		}
	}
	return "", "", false
}

func (r RedirectCfg) DetectDomain(uri *url.URL) (domain string, redirectTo string, found bool) {
	redirectCfg := reflect.ValueOf(r)
	for i := 0; i < redirectCfg.NumField(); i++ {
		field := redirectCfg.Field(i)
		if field.CanInterface() {
			if redirect, ok := field.Interface().(Redirect); ok {
				if domain, redirectTo, found = redirect.FindDomain(uri); found {
					return domain, redirectTo, true
				}
			}
		}
	}
	return "", "", false
}

func (c RootCfg) Read() error {
	cfgContents, err := os.ReadFile(c.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}
	if err = toml.Unmarshal(cfgContents, &cfg); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	return nil
}

func (c RootCfg) Watch(ctx context.Context) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		slog.Error("failed to create fs watcher", "error", err)
		return
	}
	defer func(watcher *fsnotify.Watcher) {
		err = watcher.Close()
		if err != nil {
			slog.Error("could not close fsnotify watcher", "error", err)
		}
	}(watcher)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Remove) {
					slog.Warn("config file is deleted! disabling config watcher...", "error", err)
					return
				}
				if event.Has(fsnotify.Write) {
					slog.Info("config file changed, reading...")
					if err = c.Read(); err != nil {
						slog.Error("cannot load modified file", "error", err)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				slog.Error("error from config watcher", "error", err)
			}
		}
	}()
	err = watcher.Add(c.ConfigPath)
	if err != nil {
		slog.Error("failed to watch config", "error", err)
		return
	}
	<-make(chan struct{})
}
