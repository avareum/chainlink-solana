package solana

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/big"
	"sort"

	"github.com/smartcontractkit/libocr/offchainreporting2/chains/evmutil"
	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

var _ median.ReportCodec = (*ReportCodec)(nil)

type ReportCodec struct{}

func (c ReportCodec) BuildReport(oo []median.ParsedAttributedObservation) (types.Report, error) {
	n := len(oo)
	if n == 0 {
		return nil, fmt.Errorf("cannot build report from empty attributed observations")
	}

	// copy so we can safely re-order subsequently
	oo = append([]median.ParsedAttributedObservation{}, oo...)

	// get median timestamp
	sort.Slice(oo, func(i, j int) bool {
		return oo[i].Timestamp < oo[j].Timestamp
	})
	timestamp := oo[n/2].Timestamp

	// get median juelsPerFeeCoin
	sort.Slice(oo, func(i, j int) bool {
		return oo[i].JuelsPerFeeCoin.Cmp(oo[j].JuelsPerFeeCoin) < 0
	})
	juelsPerFeeCoin := oo[n/2].JuelsPerFeeCoin

	// get median by value
	// solana program size tx execution limit prevents reporting all observations
	// reporting only median value
	sort.Slice(oo, func(i, j int) bool {
		return oo[i].Value.Cmp(oo[j].Value) < 0
	})
	median := oo[n/2].Value

	observers := [32]byte{}
	for i, o := range oo {
		observers[i] = byte(o.Observer)
	}

	// report encoding
	report := []byte{}

	time := make([]byte, 4)
	binary.BigEndian.PutUint32(time, timestamp)
	report = append(report, time[:]...)

	observersCount := uint8(len(observers))
	report = append(report, observersCount)

	report = append(report, observers[:]...)

	mBytes := make([]byte, MedianLen)
	report = append(report, median.FillBytes(mBytes)[:]...)

	jBytes := make([]byte, JuelsLen)
	report = append(report, juelsPerFeeCoin.FillBytes(jBytes)[:]...)

	return types.Report(report), nil
}

func (c ReportCodec) MedianFromReport(report types.Report) (*big.Int, error) {
	// report should contain timestamp + observers + median + juels per eth
	if len(report) != int(ReportLen) {
		return nil, fmt.Errorf("report length missmatch: %d (received), %d (expected)", len(report), ReportLen)
	}

	// unpack median observation
	start := 4 + 1 + 32
	end := start + int(MedianLen)
	median := report[start:end]
	return big.NewInt(0).SetBytes(median), nil
}

// Create report digest using SHA256 hash fn
func HashReport(ctx types.ReportContext, r types.Report) ([]byte, error) {
	rawCtx := RawReportContext(ctx)
	buf := sha256.New()
	for _, v := range [][]byte{r[:], rawCtx[0][:], rawCtx[1][:], rawCtx[2][:]} {
		if _, err := buf.Write(v); err != nil {
			return []byte{}, err
		}
	}

	return buf.Sum(nil), nil
}

func RawReportContext(ctx types.ReportContext) [3][32]byte {
	return evmutil.RawReportContext(ctx)
}
