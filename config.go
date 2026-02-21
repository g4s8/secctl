package main

import "flag"

type Config struct {
	EditorPath string
	KubeConfig string

	showVersion bool
}

func (c *Config) Parse() {
	flag.StringVar(&c.EditorPath, "editor", "", "Path to the text editor (default: $EDITOR)")
	flag.StringVar(&c.KubeConfig, "kubeconfig", "", "Path to the kubeconfig file (default: $KUBECONFIG or ~/.kube/config)")
	flag.BoolVar(&c.showVersion, "version", false, "Show version information and exit")
	flag.Parse()
}
