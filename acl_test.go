package graphvent

import (
  "testing"
  "time"
  "reflect"
  "runtime/debug"
)

func checkSignal[S Signal](t *testing.T, signal Signal, check func(S)){
  response_casted, cast_ok := signal.(S)
  if cast_ok == false {
    error_signal, is_error := signal.(*ErrorSignal)
    if is_error {
      t.Log(string(debug.Stack()))
      t.Fatal(error_signal.Error)
    }
    t.Fatalf("Response of wrong type %s", reflect.TypeOf(signal))
  }

  check(response_casted)
}

func testSendACL[S Signal](t *testing.T, ctx *Context, listener *Node, action Tree, policies []Policy, check func(S)){
  acl_node, err := NewNode(ctx, nil, BaseNodeType, 100, []Policy{DefaultACLPolicy}, NewACLExt(policies))
  fatalErr(t, err)

  acl_signal := NewACLSignal(listener.ID, action)
  response, _ := testSend(t, ctx, acl_signal, listener, acl_node)

  checkSignal(t, response, check)
}

func testErrorSignal(t *testing.T, error_string string) func(*ErrorSignal){
  return func(response *ErrorSignal) {
    if response.Error != error_string {
      t.Fatalf("Wrong error: %s", response.Error)
    }
  }
}

func testSuccess(*SuccessSignal){}

func testSend(t *testing.T, ctx *Context, signal Signal, source, destination *Node) (ResponseSignal, []Signal) {
  source_listener, err := GetExt[*ListenerExt](source, ListenerExtType)
  fatalErr(t, err)

  messages := Messages{}
  messages = messages.Add(ctx, destination.ID, source, nil, signal)
  fatalErr(t, ctx.Send(messages))

  response, signals, err := WaitForResponse(source_listener.Chan, time.Millisecond*10, signal.ID())
  fatalErr(t, err)

  return response, signals
}

func TestACLBasic(t *testing.T) {
  ctx := logTestContext(t, []string{"serialize_types", "deserialize_types", "test", "listener_debug", "group", "acl", "policy"})

  listener, err := NewNode(ctx, nil, BaseNodeType, 100, nil, NewListenerExt(100))
  fatalErr(t, err)

  testSendACL(t, ctx, listener, nil, nil, testErrorSignal(t, "acl_denied"))

  testSendACL(t, ctx, listener, nil, []Policy{NewAllNodesPolicy(nil)}, testSuccess)

  group, err := NewNode(ctx, nil, GroupNodeType, 100, []Policy{
    DefaultGroupPolicy,
    NewPerNodePolicy(map[NodeID]Tree{
      listener.ID: {
        SerializedType(AddMemberSignalType): nil,
        SerializedType(AddSubGroupSignalType): nil,
      },
    }),
  }, NewGroupExt(nil))
  fatalErr(t, err)

  testSendACL(t, ctx, listener, nil, []Policy{
    NewMemberOfPolicy(map[NodeID]map[string]Tree{
      group.ID: {
        "test_group": nil,
      },
    }),
  }, testErrorSignal(t, "acl_denied"))

  add_subgroup_signal := NewAddSubGroupSignal("test_group")
  add_subgroup_response, _ := testSend(t, ctx, add_subgroup_signal, listener, group)
  checkSignal(t, add_subgroup_response, testSuccess)

  add_member_signal := NewAddMemberSignal("test_group", listener.ID)
  add_member_response, _ := testSend(t, ctx, add_member_signal, listener, group)
  checkSignal(t, add_member_response, testSuccess)

  testSendACL(t, ctx, listener, nil, []Policy{
    NewMemberOfPolicy(map[NodeID]map[string]Tree{
      group.ID: {
        "test_group": nil,
      },
    }),
  }, testSuccess)

  testSendACL(t, ctx, listener, nil, []Policy{
    NewACLProxyPolicy(nil),
  }, testErrorSignal(t, "acl_denied"))

  acl_proxy_1, err := NewNode(ctx, nil, BaseNodeType, 100, []Policy{DefaultACLPolicy}, NewACLExt(nil))
  fatalErr(t, err)

  testSendACL(t, ctx, listener, nil, []Policy{
    NewACLProxyPolicy([]NodeID{acl_proxy_1.ID}),
  }, testErrorSignal(t, "acl_denied"))

  acl_proxy_2, err := NewNode(ctx, nil, BaseNodeType, 100, []Policy{DefaultACLPolicy}, NewACLExt([]Policy{NewAllNodesPolicy(nil)}))
  fatalErr(t, err)

  testSendACL(t, ctx, listener, nil, []Policy{
    NewACLProxyPolicy([]NodeID{acl_proxy_2.ID}),
  }, testSuccess)

  acl_proxy_3, err := NewNode(ctx, nil, BaseNodeType, 100, []Policy{DefaultACLPolicy},
    NewACLExt([]Policy{
      NewMemberOfPolicy(map[NodeID]map[string]Tree{
        group.ID: {
          "test_group": nil,
        },
      }),
    }),
  )
  fatalErr(t, err)

  testSendACL(t, ctx, listener, nil, []Policy{
    NewACLProxyPolicy([]NodeID{acl_proxy_3.ID}),
  }, testSuccess)
}
