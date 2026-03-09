package ourkafka

import (
	"context"
	"errors"
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/require"
)

type mockReader struct {
	msg      kafka.Message
	readErr  error
	closeErr error
	gotCtx   context.Context
}

func (m *mockReader) ReadMessage(ctx context.Context) (kafka.Message, error) {
	m.gotCtx = ctx
	if m.readErr != nil {
		return kafka.Message{}, m.readErr
	}
	return m.msg, nil
}

func (m *mockReader) Close() error {
	return m.closeErr
}

func TestConsumerReadMessageSuccess(t *testing.T) {
	expected := kafka.Message{Key: []byte("k1"), Value: []byte("v1")}
	mr := &mockReader{msg: expected}
	c := newConsumerWithReader(mr)

	ctx := context.Background()
	msg, err := c.ReadMessage(ctx)

	require.NoError(t, err)
	require.Equal(t, expected, msg)
	require.Equal(t, ctx, mr.gotCtx)
}

func TestConsumerReadMessageError(t *testing.T) {
	mr := &mockReader{readErr: errors.New("read failed")}
	c := newConsumerWithReader(mr)

	_, err := c.ReadMessage(context.Background())
	require.ErrorContains(t, err, "read failed")
}

func TestConsumerCloseSuccess(t *testing.T) {
	mr := &mockReader{}
	c := newConsumerWithReader(mr)

	err := c.Close()
	require.NoError(t, err)
}

func TestConsumerCloseError(t *testing.T) {
	mr := &mockReader{closeErr: errors.New("close failed")}
	c := newConsumerWithReader(mr)

	err := c.Close()
	require.ErrorContains(t, err, "close failed")
}
