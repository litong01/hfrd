package utilities

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestCreatChannelTxEnvelope(t *testing.T)  {
	_, err := CreatChannelTxEnvelope("mychannel", "SampleConsortium", "Org1MSP", "Org2MSP")
	assert.NoError(t, err)
	_, err = CreatChannelTxEnvelope("mychannel", "SampleConsortium", "Org1MSP", "Org2MSP", "Org3MSP")
	assert.NoError(t, err)
	_, err = CreatChannelTxEnvelope("mychannel", "SampleConsortium")
	assert.NoError(t, err)
}
