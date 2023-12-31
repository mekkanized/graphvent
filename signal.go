package graphvent

import (
  "fmt"
  "time"

 "github.com/google/uuid"
)

type SignalDirection uint8
const (
  Up SignalDirection = iota
  Down
  Direct
)

type TimeoutSignal struct {
  ResponseHeader
}

func NewTimeoutSignal(req_id uuid.UUID) *TimeoutSignal {
  return &TimeoutSignal{
    NewResponseHeader(req_id, Direct),
  }
}

func (signal TimeoutSignal) String() string {
  return fmt.Sprintf("TimeoutSignal(%s)", &signal.ResponseHeader)
}

// Timeouts are internal only, no permission allows sending them
func (signal TimeoutSignal) Permission() Tree {
  return nil
}

type SignalHeader struct {
  Id uuid.UUID `gv:"id"`
  Dir SignalDirection `gv:"direction"`
}

func (signal SignalHeader) ID() uuid.UUID {
  return signal.Id
}

func (signal SignalHeader) Direction() SignalDirection {
  return signal.Dir
}

func (header SignalHeader) String() string {
   return fmt.Sprintf("SignalHeader(%d, %s)", header.Dir, header.Id)
}

type ResponseSignal interface {
  Signal
  ResponseID() uuid.UUID
}

type ResponseHeader struct {
  SignalHeader
  ReqID uuid.UUID `gv:"req_id"`
}

func (header ResponseHeader) ResponseID() uuid.UUID {
  return header.ReqID
}

func (header ResponseHeader) String() string {
   return fmt.Sprintf("ResponseHeader(%d, %s->%s)", header.Dir, header.Id, header.ReqID)
}

type Signal interface {
  fmt.Stringer
  ID() uuid.UUID
  Direction() SignalDirection
  Permission() Tree
}

func WaitForResponse(listener chan Signal, timeout time.Duration, req_id uuid.UUID) (ResponseSignal, []Signal, error) {
  signals := []Signal{}
  var timeout_channel <- chan time.Time
  if timeout > 0 {
    timeout_channel = time.After(timeout)
  }

  for true {
    select {
    case signal := <- listener:
      if signal == nil {
        return nil, signals, fmt.Errorf("LISTENER_CLOSED")
      }

      resp_signal, ok := signal.(ResponseSignal)
      if ok == true && resp_signal.ResponseID() == req_id {
        return resp_signal, signals, nil
      } else {
        signals = append(signals, signal)
      }

    case <-timeout_channel:
      return nil, signals, fmt.Errorf("LISTENER_TIMEOUT")
    }
  }
  return nil, signals, fmt.Errorf("UNREACHABLE")
}

//TODO: Add []Signal return as well for other signals
func WaitForSignal[S Signal](listener chan Signal, timeout time.Duration, check func(S)bool) (S, error) {
  var zero S
  var timeout_channel <- chan time.Time
  if timeout > 0 {
    timeout_channel = time.After(timeout)
  }
  for true {
    select {
    case signal := <- listener:
      if signal == nil {
        return zero, fmt.Errorf("LISTENER_CLOSED")
      }
      sig, ok := signal.(S)
      if ok == true {
        if check(sig) == true {
          return sig, nil
        }
      }
    case <-timeout_channel:
      return zero, fmt.Errorf("LISTENER_TIMEOUT")
    }
  }
  return zero, fmt.Errorf("LOOP_ENDED")
}

func NewSignalHeader(direction SignalDirection) SignalHeader {
  return SignalHeader{
    uuid.New(),
    direction,
  }
}

func NewResponseHeader(req_id uuid.UUID, direction SignalDirection) ResponseHeader {
  return ResponseHeader{
    NewSignalHeader(direction),
    req_id,
  }
}

type CreateSignal struct {
  SignalHeader
}

func (signal CreateSignal) Permission() Tree {
  return Tree{
    SerializedType(CreateSignalType): nil,
  }
}

func NewCreateSignal() *CreateSignal {
  return &CreateSignal{
    NewSignalHeader(Direct),
  }
}

type StartSignal struct {
  SignalHeader
}
func (signal StartSignal) Permission() Tree {
  return Tree{
    SerializedType(StartSignalType): nil,
  }
}
func NewStartSignal() *StartSignal {
  return &StartSignal{
    NewSignalHeader(Direct),
  }
}

type StoppedSignal struct {
  ResponseHeader
  Source NodeID
}
func (signal StoppedSignal) Permission() Tree {
  return Tree{
    ResponseType: nil,
  }
}
func NewStoppedSignal(sig *StopSignal, source NodeID) *StoppedSignal {
  return &StoppedSignal{
    NewResponseHeader(sig.ID(), Up),
    source,
  }
}

type StopSignal struct {
  SignalHeader
}
func (signal StopSignal) Permission() Tree {
  return Tree{
    SerializedType(StopSignalType): nil,
  }
}
func NewStopSignal() *StopSignal {
  return &StopSignal{
    NewSignalHeader(Direct),
  }
}

type SuccessSignal struct {
  ResponseHeader
}

func (signal SuccessSignal) String() string {
  return fmt.Sprintf("SuccessSignal(%s)", signal.ResponseHeader)
}

func (signal SuccessSignal) Permission() Tree {
  return Tree{
    ResponseType: {
      SerializedType(SuccessSignalType): nil,
    },
  }
}
func NewSuccessSignal(req_id uuid.UUID) *SuccessSignal {
  return &SuccessSignal{
    NewResponseHeader(req_id, Direct),
  }
}

type ErrorSignal struct {
  ResponseHeader
  Error string
}
func (signal ErrorSignal) String() string {
  return fmt.Sprintf("ErrorSignal(%s, %s)", signal.ResponseHeader, signal.Error)
}
func (signal ErrorSignal) Permission() Tree {
  return Tree{
    ResponseType: {
      SerializedType(ErrorSignalType): nil,
    },
  }
}
func NewErrorSignal(req_id uuid.UUID, fmt_string string, args ...interface{}) *ErrorSignal {
  return &ErrorSignal{
    NewResponseHeader(req_id, Direct),
    fmt.Sprintf(fmt_string, args...),
  }
}

type ACLTimeoutSignal struct {
  ResponseHeader
}
func (signal ACLTimeoutSignal) Permission() Tree {
  return Tree{
    SerializedType(ACLTimeoutSignalType): nil,
  }
}
func NewACLTimeoutSignal(req_id uuid.UUID) *ACLTimeoutSignal {
  sig := &ACLTimeoutSignal{
    NewResponseHeader(req_id, Direct),
  }
  return sig
}

type StatusSignal struct {
  SignalHeader
  Source NodeID `gv:"source"`
  Changes Changes `gv:"changes"`
}
func (signal StatusSignal) Permission() Tree {
  return Tree{
    StatusType: nil,
  }
}
func (signal StatusSignal) String() string {
  return fmt.Sprintf("StatusSignal(%s, %+v)", signal.SignalHeader, signal.Changes)
}
func NewStatusSignal(source NodeID, changes Changes) *StatusSignal {
  return &StatusSignal{
    NewSignalHeader(Up),
    source,
    changes,
  }
}

type LinkSignal struct {
  SignalHeader
  NodeID NodeID
  Action string
}

const (
  LinkActionBase = "LINK_ACTION"
  LinkActionAdd = "ADD"
)

func (signal LinkSignal) Permission() Tree {
  return Tree{
    SerializedType(LinkSignalType): Tree{
      Hash(LinkActionBase, signal.Action): nil,
    },
  }
}

func NewLinkSignal(action string, id NodeID) Signal {
  return &LinkSignal{
    NewSignalHeader(Direct),
    id,
    action,
  }
}

type LockSignal struct {
  SignalHeader
  State string
}
func (signal LockSignal) String() string {
  return fmt.Sprintf("LockSignal(%s, %s)", signal.SignalHeader, signal.State)
}

const (
  LockStateBase = "LOCK_STATE"
)

func (signal LockSignal) Permission() Tree {
  return Tree{
    SerializedType(LockSignalType): Tree{
      Hash(LockStateBase, signal.State): nil,
    },
  }
}

func NewLockSignal(state string) *LockSignal {
  return &LockSignal{
    NewSignalHeader(Direct),
    state,
  }
}

type ReadSignal struct {
  SignalHeader
  Extensions map[ExtType][]string `json:"extensions"`
}

func (signal ReadSignal) String() string {
  return fmt.Sprintf("ReadSignal(%s, %+v)", signal.SignalHeader, signal.Extensions)
}

func (signal ReadSignal) Permission() Tree {
  ret := Tree{}
  for ext, fields := range(signal.Extensions) {
    field_tree := Tree{}
    for _, field := range(fields) {
      field_tree[Hash(FieldNameBase, field)]  = nil
    }
    ret[SerializedType(ext)] = field_tree
  }
  return Tree{SerializedType(ReadSignalType): ret}
}

func NewReadSignal(exts map[ExtType][]string) *ReadSignal {
  return &ReadSignal{
    NewSignalHeader(Direct),
    exts,
  }
}

type ReadResultSignal struct {
  ResponseHeader
  NodeID NodeID
  NodeType NodeType
  Extensions map[ExtType]map[string]SerializedValue
}

func (signal ReadResultSignal) String() string {
  return fmt.Sprintf("ReadResultSignal(%s, %s, %+v, %+v)", signal.ResponseHeader, signal.NodeID, signal.NodeType, signal.Extensions)
}

func (signal ReadResultSignal) Permission() Tree {
  return Tree{
    ResponseType: {
      SerializedType(ReadResultSignalType): nil,
    },
  }
}
func NewReadResultSignal(req_id uuid.UUID, node_id NodeID, node_type NodeType, exts map[ExtType]map[string]SerializedValue) *ReadResultSignal {
  return &ReadResultSignal{
    NewResponseHeader(req_id, Direct),
    node_id,
    node_type,
    exts,
  }
}

