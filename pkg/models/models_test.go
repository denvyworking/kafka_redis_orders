package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrderJSONRoundTrip(t *testing.T) {
	expected := Order{
		OrderID: "order-100",
		UserID:  "user-42",
		Total:   55.75,
		Status:  "processed",
	}

	raw, err := json.Marshal(expected)
	require.NoError(t, err)

	var got Order
	err = json.Unmarshal(raw, &got)
	require.NoError(t, err)
	require.Equal(t, expected, got)
}

func TestOrderStatusJSONRoundTrip(t *testing.T) {
	expected := OrderStatus{
		OrderID: "order-100",
		Status:  "processed",
	}

	raw, err := json.Marshal(expected)
	require.NoError(t, err)

	var got OrderStatus
	err = json.Unmarshal(raw, &got)
	require.NoError(t, err)
	require.Equal(t, expected, got)
}
