package util

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type PortForwardOptions struct {
	Local     int
	Remote    int
	Namespace string
	PodName   string
	Out       io.Writer
	stopChan  chan struct{}
	readyChan chan struct{}
	config    *rest.Config
	client    rest.Interface
}

func NewPortForwardOptions(client rest.Interface, config *rest.Config, namespace, podName string, remote int) *PortForwardOptions {
	return &PortForwardOptions{
		config:    config,
		client:    client,
		Namespace: namespace,
		PodName:   podName,
		Remote:    remote,
		stopChan:  make(chan struct{}, 1),
		readyChan: make(chan struct{}, 1),
		Out:       ioutil.Discard,
	}
}

func (p *PortForwardOptions) ForwardPort() error {
	u := p.client.Post().
		Resource("pods").
		Namespace(p.Namespace).
		Name(p.PodName).
		SubResource("portforward").URL()

	transport, upgrader, err := spdy.RoundTripperFor(p.config)
	if err != nil {
		return err
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", u)

	local, err := getAvailablePort()
	if err != nil {
		return fmt.Errorf("could not find an available port: %s", err)
	}
	p.Local = local

	ports := []string{fmt.Sprintf("%d:%d", p.Local, p.Remote)}

	pf, err := portforward.New(dialer, ports, p.stopChan, p.readyChan, p.Out, p.Out)
	if err != nil {
		return err
	}

	errChan := make(chan error)
	go func() {
		errChan <- pf.ForwardPorts()
	}()

	select {
	case err = <-errChan:
		return fmt.Errorf("forwarding ports: %v", err)
	case <-pf.Ready:
		return nil
	}
}

// stop port forward
func (p *PortForwardOptions) Stop() {
	close(p.stopChan)
}

func getAvailablePort() (int, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer l.Close()

	_, p, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		return 0, err
	}
	port, err := strconv.Atoi(p)
	if err != nil {
		return 0, err
	}
	return port, err
}
