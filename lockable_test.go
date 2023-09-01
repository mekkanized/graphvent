package graphvent

import (
  "testing"
  "time"
  "crypto/ed25519"
  "crypto/rand"
)

var TestLockableType = NewNodeType("TEST_LOCKABLE")
func lockableTestContext(t *testing.T, logs []string) *Context {
  ctx := logTestContext(t, logs)

  err := ctx.RegisterNodeType(TestLockableType, []ExtType{LockableExtType})
  fatalErr(t, err)

  return ctx
}

func TestLink(t *testing.T) {
  ctx := lockableTestContext(t, []string{"lockable"})

  l1_pub, l1_key, err := ed25519.GenerateKey(rand.Reader)
  fatalErr(t, err)
  l1_id := KeyID(l1_pub)
  policy := NewPerNodePolicy(map[NodeID]Tree{
    l1_id: nil,
  })

  l2_listener := NewListenerExt(10)
  l2 := NewNode(ctx, nil, TestLockableType, 10,
                map[PolicyType]Policy{
                  PerNodePolicyType: &policy,
                },
                  l2_listener,
                  NewLockableExt(nil),
                )

  l1_listener := NewListenerExt(10)
  l1 := NewNode(ctx, l1_key, TestLockableType, 10, nil,
                 l1_listener,
                 NewLockableExt(nil),
               )

  msgs := Messages{}
  msgs = msgs.Add(ctx, l1.ID, l1.Key, NewLinkSignal("add", l2.ID), l1.ID)
  err = ctx.Send(msgs)
  fatalErr(t, err)

  _, err = WaitForSignal(l1_listener.Chan, time.Millisecond*10, func(sig *ErrorSignal) bool {
    return sig.Error == "req_added"
  })
  fatalErr(t, err)

  msgs = Messages{}
  msgs = msgs.Add(ctx, l1.ID, l1.Key, NewLinkSignal("remove", l2.ID), l1.ID)
  err = ctx.Send(msgs)
  fatalErr(t, err)

  _, err = WaitForSignal(l1_listener.Chan, time.Millisecond*10, func(sig *ErrorSignal) bool {
    return sig.Error == "req_removed"
  })
  fatalErr(t, err)
}

func Test10KLink(t *testing.T) {
  ctx := lockableTestContext(t, []string{"test"})

  l_pub, listener_key, err := ed25519.GenerateKey(rand.Reader)
  fatalErr(t, err)
  listener_id := KeyID(l_pub)
  child_policy := NewPerNodePolicy(map[NodeID]Tree{
    listener_id: Tree{
      uint64(LockSignalType): nil,
    },
  })
  NewLockable := func()(*Node) {
    l := NewNode(ctx, nil, TestLockableType, 10,
                  map[PolicyType]Policy{
                    PerNodePolicyType: &child_policy,
                  },
                  NewLockableExt(nil),
                )
    return l
  }

  reqs := make([]NodeID, 1000)
  for i, _ := range(reqs) {
    new_lockable := NewLockable()
    reqs[i] = new_lockable.ID
  }
  ctx.Log.Logf("test", "CREATED_10K")

  l_policy := NewAllNodesPolicy(Tree{
    uint64(LockSignalType): nil,
  })
  listener := NewListenerExt(100000)
  node := NewNode(ctx, listener_key, TestLockableType, 10000,
                map[PolicyType]Policy{
                  AllNodesPolicyType: &l_policy,
                },
                listener,
                NewLockableExt(reqs),
              )
  ctx.Log.Logf("test", "CREATED_LISTENER")

  _, err = LockLockable(ctx, node, node.ID)
  fatalErr(t, err)

  _, err = WaitForSignal(listener.Chan, time.Millisecond*1000, func(sig *LockSignal) bool {
    return sig.State == "locked"
  })
  fatalErr(t, err)

  for _, _ = range(reqs) {
    _, err := WaitForSignal(listener.Chan, time.Millisecond*100, func(sig *LockSignal) bool {
      return sig.State == "locked"
    })
    fatalErr(t, err)
  }
  ctx.Log.Logf("test", "LOCKED_10K")
}

func TestLock(t *testing.T) {
  ctx := lockableTestContext(t, []string{"lockable"})

  policy := NewAllNodesPolicy(nil)

  NewLockable := func(reqs []NodeID)(*Node, *ListenerExt) {
    listener := NewListenerExt(100)
    l := NewNode(ctx, nil, TestLockableType, 10,
                  map[PolicyType]Policy{
                    AllNodesPolicyType: &policy,
                  },
                  listener,
                  NewLockableExt(reqs),
                )
    return l, listener
  }

  l2, _ := NewLockable(nil)
  l3, _ := NewLockable(nil)
  l4, _ := NewLockable(nil)
  l5, _ := NewLockable(nil)
  l0, l0_listener := NewLockable([]NodeID{l2.ID, l3.ID, l4.ID, l5.ID})
  l1, l1_listener := NewLockable([]NodeID{l2.ID, l3.ID, l4.ID, l5.ID})

  locked := func(sig *LockSignal) bool {
    return sig.State == "locked"
  }

  unlocked := func(sig *LockSignal) bool {
    return sig.State == "unlocked"
  }

  _, err := LockLockable(ctx, l0, l5.ID)
  fatalErr(t, err)
  _, err = WaitForSignal(l0_listener.Chan, time.Millisecond*10, locked)
  fatalErr(t, err)

  id, err := LockLockable(ctx, l1, l1.ID)
  fatalErr(t, err)
  _, err = WaitForSignal(l1_listener.Chan, time.Millisecond*10, func(sig *ErrorSignal) bool {
    return sig.Error == "not_unlocked" && sig.Header().ReqID == id
  })
  fatalErr(t, err)

  _, err = UnlockLockable(ctx, l0, l5.ID)
  fatalErr(t, err)
  _, err = WaitForSignal(l0_listener.Chan, time.Millisecond*10, unlocked)
  fatalErr(t, err)

  _, err = LockLockable(ctx, l1, l1.ID)
  fatalErr(t, err)
  for i := 0; i < 4; i++ {
    _, err = WaitForSignal(l1_listener.Chan, time.Millisecond*10, func(sig *LockSignal) bool {
      return sig.State == "locked"
    })
    fatalErr(t, err)
  }
}
