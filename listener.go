package graphvent

import (
  "encoding/json"
)

// A Listener extension provides a channel that can receive signals on a different thread
type ListenerExt struct {
  Buffer int
  Chan chan Signal
}

// Create a new listener extension with a given buffer size
func NewListenerExt(buffer int) *ListenerExt {
  return &ListenerExt{
    Buffer: buffer,
    Chan: make(chan Signal, buffer),
  }
}

func (ext *ListenerExt) Field(name string) interface{} {
  return ResolveFields(ext, name, map[string]func(*ListenerExt)interface{}{
    "buffer": func(ext *ListenerExt) interface{} {
      return ext.Buffer
    },
    "chan": func(ext *ListenerExt) interface{} {
      return ext.Chan
    },
  })
}

// Simple load function, unmarshal the buffer int from json
func (ext *ListenerExt) Deserialize(ctx *Context, data []byte) error {
  err := json.Unmarshal(data, &ext.Buffer)
  ext.Chan = make(chan Signal, ext.Buffer)
  return err
}

func (listener *ListenerExt) Type() ExtType {
  return ListenerExtType
}

// Send the signal to the channel, logging an overflow if it occurs
func (ext *ListenerExt) Process(ctx *Context, node *Node, source NodeID, signal Signal) Messages {
  ctx.Log.Logf("listener", "LISTENER_PROCESS: %s - %+v", node.ID, signal)
  select {
  case ext.Chan <- signal:
  default:
    ctx.Log.Logf("listener", "LISTENER_OVERFLOW: %s", node.ID)
  }
  return nil
}

func (ext *ListenerExt) Serialize() ([]byte, error) {
  return json.Marshal(ext.Buffer)
}
