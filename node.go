package graphvent

import (
  "sync"
  "github.com/google/uuid"
  badger "github.com/dgraph-io/badger/v3"
  "fmt"
  "encoding/binary"
  "crypto/sha256"
)

// IDs are how nodes are uniquely identified, and can be serialized for the database
type NodeID string

func (id NodeID) Serialize() []byte {
  return []byte(id)
}

// Types are how nodes are associated with structs at runtime(and from the DB)
type NodeType string
func (node_type NodeType) Hash() uint64 {
  hash := sha256.New()
  hash.Write([]byte(node_type))
  bytes := hash.Sum(nil)

  return binary.BigEndian.Uint64(bytes[(len(bytes)-9):(len(bytes)-1)])
}

// Generate a random id
func RandID() NodeID {
  uuid_str := uuid.New().String()
  return NodeID(uuid_str)
}

// A Node represents data that can be read by multiple goroutines and written to by one, with a unique ID attached, and a method to process updates(including propagating them to connected nodes)
// RegisterChannel and UnregisterChannel are used to connect arbitrary listeners to the node
type Node interface {
  sync.Locker
  RLock()
  RUnlock()
  Serialize() ([]byte, error)
  ID() NodeID
  Type() NodeType
  Signal(ctx *Context, signal GraphSignal, nodes NodeMap) error
  RegisterChannel(id NodeID, listener chan GraphSignal)
  UnregisterChannel(id NodeID)
}

// A GraphNode is an implementation of a Node that can be embedded into more complex structures
type GraphNode struct {
  sync.RWMutex
  listeners_lock sync.Mutex

  id NodeID
  listeners map[NodeID]chan GraphSignal
}

// GraphNode doesn't serialize any additional information by default
func (node * GraphNode) Serialize() ([]byte, error) {
  return nil, nil
}

func LoadGraphNode(ctx * Context, id NodeID, data []byte, nodes NodeMap)(Node, error) {
  if len(data) > 0 {
    return nil, fmt.Errorf("Attempted to load a graph_node with data %+v, should have been 0 length", string(data))
  }
  node := NewGraphNode(id)
  return &node, nil
}

func (node * GraphNode) ID() NodeID {
  return node.id
}

func (node * GraphNode) Type() NodeType {
  return NodeType("graph_node")
}

func (node * GraphNode) Signal(ctx *Context, signal GraphSignal, nodes NodeMap) error {
  ctx.Log.Logf("signal", "SIGNAL: %s - %s", node.ID(), signal.String())
  node.listeners_lock.Lock()
  defer node.listeners_lock.Unlock()
  closed := []NodeID{}

  for id, listener := range node.listeners {
    ctx.Log.Logf("signal", "UPDATE_LISTENER %s: %p", node.ID(), listener)
    select {
    case listener <- signal:
    default:
      ctx.Log.Logf("signal", "CLOSED_LISTENER %s: %p", node.ID(), listener)
      go func(node Node, listener chan GraphSignal) {
        listener <- NewDirectSignal(node, "listener_closed")
        close(listener)
      }(node, listener)
      closed = append(closed, id)
    }
  }

  for _, id := range(closed) {
    delete(node.listeners, id)
  }
  return nil
}

func (node * GraphNode) RegisterChannel(id NodeID, listener chan GraphSignal) {
  node.listeners_lock.Lock()
  _, exists := node.listeners[id]
  if exists == false {
    node.listeners[id] = listener
  }
  node.listeners_lock.Unlock()
}

func (node * GraphNode) UnregisterChannel(id NodeID) {
  node.listeners_lock.Lock()
  _, exists := node.listeners[id]
  if exists == false {
    panic("Attempting to unregister non-registered listener")
  } else {
    delete(node.listeners, id)
  }
  node.listeners_lock.Unlock()
}

func NewGraphNode(id NodeID) GraphNode {
  return GraphNode{
    id: id,
    listeners: map[NodeID]chan GraphSignal{},
  }
}

const NODE_DB_MAGIC = 0x2491df14
const NODE_DB_HEADER_LEN = 12
type DBHeader struct {
  Magic uint32
  TypeHash uint64
}

func (header DBHeader) Serialize() []byte {
  if header.Magic != NODE_DB_MAGIC {
    panic(fmt.Sprintf("Serializing header with invalid magic %0x", header.Magic))
  }

  ret := make([]byte, NODE_DB_HEADER_LEN)
  binary.BigEndian.PutUint32(ret[0:4], header.Magic)
  binary.BigEndian.PutUint64(ret[4:12], header.TypeHash)
  return ret
}

func NewDBHeader(node_type NodeType) DBHeader {
  return DBHeader{
    Magic: NODE_DB_MAGIC,
    TypeHash: node_type.Hash(),
  }
}

func getNodeBytes(ctx * Context, node Node) ([]byte, error) {
  if node == nil {
    return nil, fmt.Errorf("DB_SERIALIZE_ERROR: cannot serialize nil node")
  }
  ser, err := node.Serialize()
  if err != nil {
    return nil, fmt.Errorf("DB_SERIALIZE_ERROR: %e", err)
  }

  header := NewDBHeader(node.Type())

  db_data := append(header.Serialize(), ser...)

  return db_data, nil
}

// Write a node to the database
func WriteNode(ctx * Context, node Node) error {
  ctx.Log.Logf("db", "DB_WRITE: %+v", node)

  node_bytes, err := getNodeBytes(ctx, node)
  if err != nil {
    return err
  }

  id_ser := node.ID().Serialize()

  err = ctx.DB.Update(func(txn *badger.Txn) error {
    err := txn.Set(id_ser, node_bytes)
    return err
  })

  return err
}

// Write multiple nodes to the database in a single transaction
func WriteNodes(ctx * Context, nodes NodeMap) error {
  ctx.Log.Logf("db", "DB_WRITES: %d", len(nodes))
  if nodes == nil {
    return fmt.Errorf("Cannot write nil map to DB")
  }

  serialized_bytes := make([][]byte, len(nodes))
  serialized_ids := make([][]byte, len(nodes))
  i := 0
  for _, node := range(nodes) {
    node_bytes, err := getNodeBytes(ctx, node)
    if err != nil {
      return err
    }

    id_ser := node.ID().Serialize()

    serialized_bytes[i] = node_bytes
    serialized_ids[i] = id_ser

    i++
  }

  err := ctx.DB.Update(func(txn *badger.Txn) error {
    for i, id := range(serialized_ids) {
      err := txn.Set(id, serialized_bytes[i])
      if err != nil {
        return err
      }
    }
    return nil
  })

  return err
}

// Get the bytes associates with `id` in the database, or error
func readNodeBytes(ctx * Context, id NodeID) (uint64, []byte, error) {
  var bytes []byte
  err := ctx.DB.View(func(txn *badger.Txn) error {
    item, err := txn.Get(id.Serialize())
    if err != nil {
      return err
    }

    return item.Value(func(val []byte) error {
      bytes = append([]byte{}, val...)
      return nil
    })
  })

  if err != nil {
    ctx.Log.Logf("db", "DB_READ_ERR: %s - %e", id, err)
    return 0, nil, err
  }

  if len(bytes) < NODE_DB_HEADER_LEN {
    return 0, nil, fmt.Errorf("header for %s is %d/%d bytes", id, len(bytes), NODE_DB_HEADER_LEN)
  }

  header := DBHeader{}
  header.Magic = binary.BigEndian.Uint32(bytes[0:4])
  header.TypeHash = binary.BigEndian.Uint64(bytes[4:12])

  if header.Magic != NODE_DB_MAGIC {
    return 0, nil, fmt.Errorf("header for %s, invalid magic 0x%x", id, header.Magic)
  }

  node_bytes := make([]byte, len(bytes) - NODE_DB_HEADER_LEN)
  copy(node_bytes, bytes[NODE_DB_HEADER_LEN:])

  ctx.Log.Logf("db", "DB_READ: %s - %s", id, string(bytes))

  return header.TypeHash, node_bytes, nil
}

func LoadNode(ctx * Context, id NodeID) (Node, error) {
  nodes := NodeMap{}
  return LoadNodeRecurse(ctx, id, nodes)
}

func LoadNodeRecurse(ctx * Context, id NodeID, nodes NodeMap) (Node, error) {
  node, exists := nodes[id]
  if exists == false {
    type_hash, bytes, err := readNodeBytes(ctx, id)
    if err != nil {
      return nil, err
    }

    node_type, exists := ctx.Types[type_hash]
    if exists == false {
      return nil, fmt.Errorf("0x%x is not a known node type: %+s", type_hash, bytes)
    }

    if node_type.Load == nil {
      return nil, fmt.Errorf("0x%x is an invalid node type, nil Load", type_hash)
    }

    node, err = node_type.Load(ctx, id, bytes, nodes)
    if err != nil {
      return nil, err
    }

    ctx.Log.Logf("db", "DB_NODE_LOADED: %s", id)
  }
  return node, nil
}

func checkForDuplicate(nodes []Node) error {
  found := map[NodeID]bool{}
  for _, node := range(nodes) {
    if node == nil {
      return fmt.Errorf("Cannot get state of nil node")
    }

    _, exists := found[node.ID()]
    if exists == true {
      return fmt.Errorf("Attempted to get state of %s twice", node.ID())
    }
    found[node.ID()] = true
  }
  return nil
}

func NodeList[K Node](list []K) []Node {
  nodes := make([]Node, len(list))
  for i, node := range(list) {
    nodes[i] = node
  }
  return nodes
}

type NodeMap map[NodeID]Node
type NodesFn func(nodes NodeMap) error
func UseStates(ctx * Context, init_nodes []Node, nodes_fn NodesFn) error {
  nodes := NodeMap{}
  return UseMoreStates(ctx, init_nodes, nodes, nodes_fn)
}
func UseMoreStates(ctx * Context, new_nodes []Node, nodes NodeMap, nodes_fn NodesFn) error {
  err := checkForDuplicate(new_nodes)
  if err != nil {
    return err
  }

  locked_nodes := []Node{}
  for _, node := range(new_nodes) {
    _, locked := nodes[node.ID()]
    if locked == false {
      node.RLock()
      nodes[node.ID()] = node
      locked_nodes = append(locked_nodes, node)
    }
  }

  err = nodes_fn(nodes)

  for _, node := range(locked_nodes) {
    delete(nodes, node.ID())
    node.RUnlock()
  }

  return err
}

func UpdateStates(ctx * Context, nodes []Node, nodes_fn NodesFn) error {
  locked_nodes := NodeMap{}
  err := UpdateMoreStates(ctx, nodes, locked_nodes, nodes_fn)
  if err == nil {
    err = WriteNodes(ctx, locked_nodes)
  }

  for _, node := range(locked_nodes) {
    node.Unlock()
  }
  return err
}
func UpdateMoreStates(ctx * Context, nodes []Node, locked_nodes NodeMap, nodes_fn NodesFn) error {
  for _, node := range(nodes) {
    _, locked := locked_nodes[node.ID()]
    if locked == false {
      node.Lock()
      locked_nodes[node.ID()] = node
    }
  }

  return nodes_fn(locked_nodes)
}

func UpdateChannel(node Node, buffer int, id NodeID) chan GraphSignal {
  if node == nil {
    panic("Cannot get an update channel to nil")
  }
  new_listener := make(chan GraphSignal, buffer)
  node.RegisterChannel(id, new_listener)
  return new_listener
}
