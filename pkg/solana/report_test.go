package solana

import (
	"encoding/binary"
	"math/big"
	"testing"
	"time"

	"github.com/smartcontractkit/libocr/commontypes"
	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
	"github.com/stretchr/testify/assert"
)

func TestBuildReport(t *testing.T) {
	c := ReportCodec{}
	oo := []median.ParsedAttributedObservation{}

	// expected outputs
	n := 4
	observers := make([]byte, 32)
	v := big.NewInt(0)
	v.SetString("1000000000000000000", 10)

	for i := 0; i < n; i++ {
		oo = append(oo, median.ParsedAttributedObservation{
			Timestamp:       uint32(time.Now().Unix()),
			Value:           big.NewInt(1234567890),
			JuelsPerFeeCoin: v,
			Observer:        commontypes.OracleID(i),
		})

		// create expected outputs
		observers[i] = uint8(i)
	}

	report, err := c.BuildReport(oo)
	assert.NoError(t, err)

	// validate length
	assert.Equal(t, int(ReportLen), len(report), "validate length")

	// validate timestamp
	assert.Equal(t, oo[0].Timestamp, binary.BigEndian.Uint32(report[0:4]), "validate timestamp")

	// validate observer count
	assert.Equal(t, uint8(len(observers)), report[4], "validate observer count")

	// validate observers
	index := 4 + 1
	assert.Equal(t, observers, []byte(report[index:index+32]), "validate observers")

	// validate median observation
	index = 4 + 1 + 32
	assert.Equal(t, oo[0].Value.FillBytes(make([]byte, 16)), []byte(report[index:index+16]), "validate median observation")

	// validate juelsToEth
	assert.Equal(t, v.FillBytes(make([]byte, 8)), []byte(report[ReportLen-8:ReportLen]), "validate juelsToEth")
}

func TestMedianFromReport(t *testing.T) {
	c := ReportCodec{}

	report := types.Report{
		97, 91, 43, 83, // observations_timestamp
		2,                                                                                              // observer_count
		0, 1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // observers
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 73, 150, 2, 210, // observation 2
		13, 224, 182, 179, 167, 100, 0, 0, // juels per luna (1 with 18 decimal places)
	}

	res, err := c.MedianFromReport(report)
	assert.NoError(t, err)
	assert.Equal(t, "1234567890", res.String())
}

func TestHashReport(t *testing.T) {
	var mockDigest = [32]byte{
		0, 3, 94, 221, 213, 66, 228, 80, 239, 231, 7, 96,
		83, 156, 95, 165, 199, 168, 222, 107, 47, 238, 157, 46,
		65, 205, 71, 121, 195, 138, 77, 137,
	}
	var mockReportCtx = types.ReportContext{
		ReportTimestamp: types.ReportTimestamp{
			ConfigDigest: mockDigest,
			Epoch:        1,
			Round:        1,
		},
		ExtraHash: [32]byte{},
	}

	var mockReport = types.Report{
		97, 91, 43, 83, // observations_timestamp
		2,                                                                                              // observer_count
		0, 1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // observers
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 210, // median
		13, 224, 182, 179, 167, 100, 0, 0, // juels per sol (1 with 18 decimal places)
	}

	var mockHash = []byte{
		0x9a, 0xe0, 0xc3, 0x7d, 0x9d, 0x45, 0x58, 0xdc,
		0x1e, 0x8b, 0xbc, 0xf4, 0x7d, 0x6b, 0xc8, 0xb0,
		0x5, 0xbe, 0xbe, 0x5f, 0xd, 0x28, 0x33, 0x3b,
		0x27, 0x11, 0x33, 0x5f, 0xed, 0x43, 0x91, 0x60,
	}

	h, err := HashReport(mockReportCtx, mockReport)
	assert.NoError(t, err)
	assert.Equal(t, mockHash, h)
}
