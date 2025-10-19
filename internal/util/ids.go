package util

import (
	crand "crypto/rand"
	"encoding/binary"
	"math/rand"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
	"go.uber.org/fx"
)

// IDGenerator produces monotonic ULIDs safe for concurrent use.
type IDGenerator struct {
	mu      sync.Mutex
	entropy *ulid.MonotonicEntropy
}

// NewIDGenerator constructs a generator seeded with crypto/rand.
func NewIDGenerator() (*IDGenerator, error) {
	var seed int64
	if err := binary.Read(crand.Reader, binary.LittleEndian, &seed); err != nil {
		return nil, err
	}

	if seed == 0 {
		seed = time.Now().UnixNano()
	} else if seed < 0 {
		seed = -seed
	}

	source := rand.New(rand.NewSource(seed))
	return &IDGenerator{
		entropy: ulid.Monotonic(source, 0),
	}, nil
}

// Module wires the IDGenerator for Fx consumers.
var Module = fx.Module(
	"id-generator",
	fx.Provide(NewIDGenerator),
)

// New generates a new ULID string.
func (g *IDGenerator) New() string {
	g.mu.Lock()
	defer g.mu.Unlock()

	return ulid.MustNew(ulid.Timestamp(time.Now()), g.entropy).String()
}
