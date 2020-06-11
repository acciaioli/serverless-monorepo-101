package main_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEcho(t *testing.T) {
	require.Equal(t, 1+1, 2)
}
