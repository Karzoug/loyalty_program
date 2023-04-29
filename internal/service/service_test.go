package service

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"io"
	"math"
	mathrand "math/rand"
	"testing"

	"github.com/Karzoug/loyalty_program/internal/model/order"
	"github.com/Karzoug/loyalty_program/internal/repository/processor"
	pmock "github.com/Karzoug/loyalty_program/internal/repository/processor/mock"
	smock "github.com/Karzoug/loyalty_program/internal/repository/storage/mock"
	"github.com/Karzoug/loyalty_program/pkg/luhn"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var rnd = func() *mathrand.Rand {
	buf := make([]byte, 8)
	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		panic(err)
	}
	src := mathrand.NewSource(int64(binary.LittleEndian.Uint64(buf)))
	return mathrand.New(src)
}()

func newMockServiceWithEmptyProcessor(ctx context.Context, t *testing.T) *Service {
	t.Helper()

	config := zap.NewDevelopmentConfig()
	logger, _ := config.Build()

	storages, err := smock.NewStorages(ctx)
	require.NoError(t, err)

	proc := pmock.NewOrder()
	proc.SetResult(nil, processor.ErrServerNotRespond)

	return New(storages, proc, logger)
}

func generateOrderNumber(t *testing.T) order.Number {
	t.Helper()

	ds := digitSeq(5, 15)
	cd := luhnCheckDigit(ds)
	return order.Number(ds*10 + cd)
}

func generateInvalidOrderNumber(t *testing.T) order.Number {
	t.Helper()

	ds := digitSeq(5, 15)
	cd := luhnCheckDigit(ds)
	if cd < 5 {
		return order.Number(ds*10 + cd + 1)
	}
	return order.Number(ds*10 + cd - 1)
}

func luhnCheckDigit(number int64) int64 {
	checkNumber := luhn.Checksum(number)

	if checkNumber == 0 {
		return 0
	}
	return 10 - checkNumber
}

func digitSeq(minLen, maxLen int) int64 {
	slen := rnd.Intn(maxLen-minLen) + minLen
	var res int64 = 0
	for i := 0; i < slen; i++ {
		idx := rnd.Int63n(9)
		if i == 0 && idx == 0 {
			continue
		}
		res += idx * int64(math.Pow10(i))
	}

	return res
}
