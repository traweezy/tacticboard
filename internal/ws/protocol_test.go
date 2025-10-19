package ws

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/traweezy/tacticboard/internal/model"
)

func TestDecodeClientMessage_Hello(t *testing.T) {
	data := []byte(`{"type":"hello","roomId":"abc","cap":"edit","since":1,"token":"tok"}`)
	env, err := DecodeClientMessage(data)
	require.NoError(t, err)
	require.NotNil(t, env.Hello)
	require.Equal(t, "abc", env.Hello.RoomID)
	require.Nil(t, env.Op)
}

func TestDecodeClientMessage_Invalid(t *testing.T) {
	_, err := DecodeClientMessage([]byte(`{"type":"unknown"}`))
	require.Error(t, err)
}

func TestEncodeDelta(t *testing.T) {
	op := model.Operation{
		RoomID: "room-1",
		Seq:    5,
		Ops:    []json.RawMessage{json.RawMessage(`{"k":"move"}`)},
	}

	payload, err := EncodeDelta(op)
	require.NoError(t, err)

	var decoded DeltaPayload
	require.NoError(t, json.Unmarshal(payload, &decoded))
	require.Equal(t, TypeDelta, decoded.Type)
	require.EqualValues(t, 4, decoded.From)
	require.EqualValues(t, 5, decoded.To)
	require.Len(t, decoded.Ops, 1)
}

func TestEncodeError(t *testing.T) {
	payload := EncodeError(ErrorInvalid, "bad")
	var decoded ErrorPayload
	require.NoError(t, json.Unmarshal(payload, &decoded))
	require.Equal(t, ErrorInvalid, decoded.Code)
	require.Equal(t, "bad", decoded.Msg)
}

func TestEncodeSnapshot(t *testing.T) {
	snapshot := model.Snapshot{RoomID: "room-1", Seq: 2, State: json.RawMessage(`{"nodes":[]}`)}
	payload, err := EncodeSnapshot("room-1", snapshot)
	require.NoError(t, err)

	var decoded SnapshotPayload
	require.NoError(t, json.Unmarshal(payload, &decoded))
	require.EqualValues(t, 2, decoded.Seq)
	require.JSONEq(t, `{"nodes":[]}`, string(decoded.State))
}
