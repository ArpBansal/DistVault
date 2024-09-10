package peer2peer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPtransport(t *testing.T) {
	listenAddr := ":4000"
	tr := NewTCPtransport(listenAddr)
	assert.Equal(t, tr.ListenAddress, listenAddr)

	assert.Nil(t, tr.ListenAndAccept())
	select {}
}
