package peer2peer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPtransport(t *testing.T) {
	opts := TCPtransportOps{
		ListenAddr: ":3000",
	}
	tr := NewTCPtransport(opts)
	assert.Equal(t, tr.ListenAddr, opts.ListenAddr) //not sure about it , prev for ref: assert.Equal(t, tr.ListenAddress, listenAddr)

	assert.Nil(t, tr.ListenAndAccept())

}
