package ourkafka

import (
	"context"
	"errors"
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/require"
)

type mockWriter struct {
	writeErr error
	closeErr error
	gotCtx   context.Context
	gotMsgs  []kafka.Message
}

func (m *mockWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	m.gotCtx = ctx
	m.gotMsgs = append([]kafka.Message(nil), msgs...)
	return m.writeErr
}

func (m *mockWriter) Close() error {
	return m.closeErr
}

func TestMessage(t *testing.T) {
	msg := Message("user-1", `{"order_id":"1"}`)
	require.Equal(t, []byte("user-1"), msg.Key)
	require.Equal(t, []byte(`{"order_id":"1"}`), msg.Value)
}

func TestProducerWriteMessagesSuccess(t *testing.T) {
	mw := &mockWriter{}
	p := newProducerWithWriter(mw)

	ctx := context.Background()
	m1 := kafka.Message{Key: []byte("k1"), Value: []byte("v1")}
	m2 := kafka.Message{Key: []byte("k2"), Value: []byte("v2")}

	err := p.WriteMessages(ctx, m1, m2)
	require.NoError(t, err)
	require.Equal(t, ctx, mw.gotCtx)
	require.Equal(t, []kafka.Message{m1, m2}, mw.gotMsgs)
}

func TestProducerWriteMessagesError(t *testing.T) {
	mw := &mockWriter{writeErr: errors.New("write failed")}
	p := newProducerWithWriter(mw)

	err := p.WriteMessages(context.Background(), kafka.Message{})
	require.ErrorContains(t, err, "write failed")
}

func TestProducerCloseSuccess(t *testing.T) {
	mw := &mockWriter{}
	p := newProducerWithWriter(mw)

	err := p.Close()
	require.NoError(t, err)
}

func TestProducerCloseError(t *testing.T) {
	mw := &mockWriter{closeErr: errors.New("close failed")}
	p := newProducerWithWriter(mw)

	err := p.Close()
	require.ErrorContains(t, err, "close failed")
}
