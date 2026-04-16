package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSeedCars_UsesStableFixtureTimestamps(t *testing.T) {
	t.Parallel()

	cars := seedCars()

	require.Len(t, cars, 5)

	for _, car := range cars {
		assert.Equal(t, seedCreatedAt, car.CreatedAt)
		assert.Equal(t, seedUpdatedAt, car.UpdatedAt)
	}
}

func TestSeedCars_IsDeterministicAcrossCalls(t *testing.T) {
	t.Parallel()

	first := seedCars()
	time.Sleep(10 * time.Millisecond)
	second := seedCars()

	require.Len(t, first, len(second))

	for i := range first {
		assert.Equal(t, first[i], second[i])
	}
}
