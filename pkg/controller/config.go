package controller

import (
	"strings"
	"time"
)

type Options struct {
	ClusterName      string
	VaultAddress     string
	VaultToken       string
	CACertFile       string
	ResyncPeriod     time.Duration
	TokenRenewPeriod time.Duration
	MaxNumRequeues   int
}

func (opt Options) SecretBackend() string {
	return strings.ToLower(opt.ClusterName) + "-secrets/"
}

func (opt Options) AuthBackend() string {
	return strings.ToLower(opt.ClusterName) + "-service-accounts/"
}
