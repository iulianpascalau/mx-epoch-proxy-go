package api

import (
	"fmt"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAPIEngine(t *testing.T) {
	t.Parallel()

	t.Run("nil handler should error", func(t *testing.T) {
		t.Parallel()

		engine, err := NewAPIEngine(":0", nil)
		assert.Nil(t, engine)
		assert.Equal(t, errNilHandler, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		engine, err := NewAPIEngine(":0", &testscommon.HttpHandlerStub{})
		assert.NotNil(t, engine)
		assert.Nil(t, err)

		fmt.Println(engine.Address())
	})
	t.Run("should error if can not bind on port", func(t *testing.T) {
		t.Parallel()

		firstEngine, err := NewAPIEngine(":0", &testscommon.HttpHandlerStub{})
		require.Nil(t, err)

		// Get the actual address/port chosen by the OS
		addr := firstEngine.Address()

		secondEngine, err := NewAPIEngine(addr, &testscommon.HttpHandlerStub{})
		assert.Nil(t, secondEngine)
		assert.NotNil(t, err)

		_ = firstEngine.Close()
	})
}

func TestApiEngine_Close(t *testing.T) {
	t.Parallel()

	engine, _ := NewAPIEngine(":0", &testscommon.HttpHandlerStub{})

	err := engine.Close()
	require.Nil(t, err)
}
