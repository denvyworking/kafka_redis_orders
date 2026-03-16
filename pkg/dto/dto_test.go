package dto

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrderJSONRoundTrip(t *testing.T) {
	expected := Order{
		OrderID: "order-200",
		UserID:  "user-7",
		Total:   19.99,
	}

	raw, err := json.Marshal(expected)
	require.NoError(t, err)

	var got Order
	err = json.Unmarshal(raw, &got)
	require.NoError(t, err)
	require.Equal(t, expected, got)
}
