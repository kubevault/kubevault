package controller

import "time"

type Options struct {
	ClusterName    string
	VaultAddress   string
	VaultToken     string
	CACertFile     string
	ResyncPeriod   time.Duration
	MaxNumRequeues int
}
