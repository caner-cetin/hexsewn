package main

import (
	"github.com/samber/lo"
	"net/url"
	"reflect"
	"strings"
)

type RootCfg struct {
	RedirectConfig RedirectCfg `toml:"Redirect"`
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
	domain, found = lo.Find(r.From, func(domain string) bool {
		return strings.Contains(uri.Host, domain)
	})
	return domain, r.To, found
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
