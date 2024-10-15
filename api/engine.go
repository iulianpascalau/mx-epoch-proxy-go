package api

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
)

const timeToWaitForInitialError = time.Second * 2
const interfaceType = "tcp4"

type apiEngine struct {
	server   *http.Server
	chErrors chan error
	chClosed chan struct{}
	address  string
}

// NewAPIEngine creates and opens a new API engine
func NewAPIEngine(address string, generalHandler http.Handler) (*apiEngine, error) {
	if check.IfNilReflect(generalHandler) {
		return nil, errNilHandler
	}

	ln, err := net.Listen(interfaceType, address)
	if err != nil {
		return nil, err
	}

	serv := &http.Server{
		Handler: generalHandler,
	}

	engine := &apiEngine{
		server:   serv,
		chErrors: make(chan error, 1),
		chClosed: make(chan struct{}),
		address:  ln.Addr().String(),
	}

	go func() {
		errSeve := serv.Serve(ln)
		if errors.Is(errSeve, http.ErrServerClosed) {
			close(engine.chClosed)
			return
		}

		engine.chErrors <- errSeve
	}()

	select {
	case err = <-engine.chErrors:
		return nil, err
	case <-time.After(timeToWaitForInitialError):
	}

	return engine, nil
}

// Address returns the bound address
func (engine *apiEngine) Address() string {
	return engine.address
}

// Close will close the API engine
func (engine *apiEngine) Close() error {
	err := engine.server.Shutdown(context.Background())
	if err != nil {
		return err
	}

	select {
	case err = <-engine.chErrors:
		return err
	case <-engine.chClosed:
		return nil
	}
}
