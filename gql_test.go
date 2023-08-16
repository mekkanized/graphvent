package graphvent

import (
  "testing"
  "time"
  "fmt"
  "encoding/json"
  "encoding/base64"
  "io"
  "net/http"
  "net"
  "crypto/tls"
  "crypto/x509"
  "crypto/rand"
  "crypto/ed25519"
  "bytes"
  "golang.org/x/net/websocket"
  "github.com/google/uuid"
)

func TestGQLServer(t *testing.T) {
  ctx := logTestContext(t, []string{"test", "gql", "policy", "pending"})

  TestNodeType := NodeType("TEST")
  err := ctx.RegisterNodeType(TestNodeType, []ExtType{LockableExtType})
  fatalErr(t, err)

  pub, gql_key, err := ed25519.GenerateKey(rand.Reader)
  fatalErr(t, err)
  gql_id := KeyID(pub)

  group_policy_1 := NewAllNodesPolicy(Tree{
    ReadSignalType.String(): Tree{
      GroupExtType.String(): Tree{
        "members": Tree{},
      },
    },
    ReadResultSignalType.String(): nil,
    ErrorSignalType.String(): nil,
  })

  group_policy_2 := NewMemberOfPolicy(map[NodeID]Tree{
    gql_id: Tree{
      LinkSignalType.String(): nil,
      LinkStartSignalType.String(): nil,
      LockSignalType.String(): nil,
      StatusSignalType.String(): nil,
      ReadSignalType.String(): nil,
    },
  })

  user_policy_1 := NewAllNodesPolicy(Tree{
    ReadResultSignalType.String(): nil,
    ErrorSignalType.String(): nil,
  })

  user_policy_2 := NewMemberOfPolicy(map[NodeID]Tree{
    gql_id: Tree{
      LinkSignalType.String(): nil,
      ReadSignalType.String(): nil,
    },
  })

  gql_ext, err := NewGQLExt(ctx, ":0", nil, nil)
  fatalErr(t, err)

  listener_ext := NewListenerExt(10)
  n1 := NewNode(ctx, nil, TestNodeType, 10, map[PolicyType]Policy{
    MemberOfPolicyType: &user_policy_2,
    AllNodesPolicyType: &user_policy_1,
  }, NewLockableExt(nil))

  gql := NewNode(ctx, gql_key, GQLNodeType, 10, map[PolicyType]Policy{
    MemberOfPolicyType: &group_policy_2,
    AllNodesPolicyType: &group_policy_1,
  }, NewLockableExt([]NodeID{n1.ID}), gql_ext, NewGroupExt(map[NodeID]string{
    n1.ID: "user",
    gql_id: "self",
  }), listener_ext)

  ctx.Log.Logf("test", "GQL:  %s", gql.ID)
  ctx.Log.Logf("test", "NODE: %s", n1.ID)

  _, err = WaitForSignal(listener_ext.Chan, 100*time.Millisecond, StatusSignalType, func(sig *IDStringSignal) bool {
    return sig.Str == "server_started"
  })
  fatalErr(t, err)

  skipVerifyTransport := &http.Transport{
    TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
  }
  client := &http.Client{Transport: skipVerifyTransport}
  port := gql_ext.tcp_listener.Addr().(*net.TCPAddr).Port
  url := fmt.Sprintf("https://localhost:%d/gql", port)
  ws_url := fmt.Sprintf("wss://127.0.0.1:%d/gqlws", port)

  req_1 := GQLPayload{
    Query: "query Node($id:String) { Node(id:$id) { ID, TypeHash } }",
    Variables: map[string]interface{}{
      "id": n1.ID.String(),
    },
  }

  req_2 := GQLPayload{
    Query: "query Node($id:String) { Node(id:$id) { ID, TypeHash, ... on GQLServer { Members { ID } , Listen, Requirements { ID, TypeHash Owner { ID } } } } }",
    Variables: map[string]interface{}{
      "id": gql.ID.String(),
    },
  }

  auth_username := base64.StdEncoding.EncodeToString(n1.ID.Serialize())
  key_bytes, err := x509.MarshalPKCS8PrivateKey(n1.Key)
  fatalErr(t, err)
  auth_password := base64.StdEncoding.EncodeToString(key_bytes)
  auth_b64 := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", auth_username, auth_password)))

  SendGQL := func(payload GQLPayload) []byte {
    ser, err := json.MarshalIndent(&payload, "", "  ")
    fatalErr(t, err)

    req_data := bytes.NewBuffer(ser)
    req, err := http.NewRequest("GET", url, req_data)
    fatalErr(t, err)

    req.SetBasicAuth(auth_username, auth_password)
    resp, err := client.Do(req)
    fatalErr(t, err)

    body, err := io.ReadAll(resp.Body)
    fatalErr(t, err)

    resp.Body.Close()
    return body
  }

  resp_1 := SendGQL(req_1)
  ctx.Log.Logf("test", "RESP_1: %s", resp_1)
  resp_2 := SendGQL(req_2)
  ctx.Log.Logf("test", "RESP_2: %s", resp_2)

  sub_1 := GQLPayload{
    Query: "subscription { Self }",
  }

  SubGQL := func(payload GQLPayload) {
    config, err := websocket.NewConfig(ws_url, url)
    fatalErr(t, err)
    config.Protocol = append(config.Protocol, "graphql-ws")
    config.TlsConfig = &tls.Config{InsecureSkipVerify: true}
    config.Header.Add("Authorization", fmt.Sprintf("Basic %s", auth_b64))

    ws, err := websocket.DialConfig(config)

    fatalErr(t, err)

    init := GQLWSMsg{
      ID: uuid.New().String(),
      Type: "connection_init",
    }

    ser, err := json.Marshal(&init)
    fatalErr(t, err)

    _, err = ws.Write(ser)
    fatalErr(t, err)

    resp := make([]byte, 1024)
    n, err := ws.Read(resp)

    var init_resp GQLWSMsg
    err = json.Unmarshal(resp[:n], &init_resp)
    fatalErr(t, err)

    if init_resp.Type != "connection_ack" {
      t.Fatal("Didn't receive connection_ack")
    }

    sub := GQLWSMsg{
      ID: uuid.New().String(),
      Type: "subscribe",
      Payload: sub_1,
    }

    ser, err = json.Marshal(&sub)
    fatalErr(t, err)
    _, err = ws.Write(ser)
    fatalErr(t, err)

    for i := 0; i < 10; i++ {
      n, err = ws.Read(resp)
      fatalErr(t, err)

      ctx.Log.Logf("test", "SUB_%d: %s", i, resp[:n])
    }
  }

  SubGQL(sub_1)

  msgs := Messages{}
  msgs = msgs.Add(gql.ID, gql.Key, &StopSignal, gql.ID)
  err = ctx.Send(msgs)
  fatalErr(t, err)
  _, err = WaitForSignal(listener_ext.Chan, 100*time.Millisecond, StatusSignalType, func(sig *IDStringSignal) bool {
    return sig.Str == "stopped"
  })
  fatalErr(t, err)
}

func TestGQLDB(t *testing.T) {
  ctx := logTestContext(t, []string{"test"})

  TestUserNodeType := NodeType("TEST_USER")
  err := ctx.RegisterNodeType(TestUserNodeType, []ExtType{})
  fatalErr(t, err)
  u1 := NewNode(ctx, nil, TestUserNodeType, 10, nil)

  ctx.Log.Logf("test", "U1_ID: %s", u1.ID)

  gql_ext, err := NewGQLExt(ctx, ":0", nil, nil)
  fatalErr(t, err)
  listener_ext := NewListenerExt(10)
  gql := NewNode(ctx, nil, GQLNodeType, 10, nil,
                 gql_ext,
                 listener_ext,
                 NewGroupExt(nil))
  ctx.Log.Logf("test", "GQL_ID: %s", gql.ID)

  msgs := Messages{}
  msgs = msgs.Add(gql.ID, gql.Key, &StopSignal, gql.ID)
  err = ctx.Send(msgs)
  fatalErr(t, err)
  _, err = WaitForSignal(listener_ext.Chan, 100*time.Millisecond, StatusSignalType, func(sig *IDStringSignal) bool {
    return sig.Str == "stopped" && sig.NodeID == gql.ID
  })
  fatalErr(t, err)

  ser1, err := gql.Serialize()
  ser2, err := u1.Serialize()
  ser3, err := StopSignal.Serialize()
  ctx.Log.Logf("test", "SER_1: \n%s\n\n", ser1)
  ctx.Log.Logf("test", "SER_2: \n%s\n\n", ser2)
  ctx.Log.Logf("test", "SER_3: \n%s\n\n", ser3)

  // Clear all loaded nodes from the context so it loads them from the database
  ctx.Nodes = map[NodeID]*Node{}
  gql_loaded, err := LoadNode(ctx, gql.ID)
  fatalErr(t, err)
  listener_ext, err = GetExt[*ListenerExt](gql_loaded)
  fatalErr(t, err)
  msgs = Messages{}
  msgs = msgs.Add(gql_loaded.ID, gql_loaded.Key, &StopSignal, gql_loaded.ID)
  err = ctx.Send(msgs)
  fatalErr(t, err)
  _, err = WaitForSignal(listener_ext.Chan, 100*time.Millisecond, StatusSignalType, func(sig *IDStringSignal) bool {
    return sig.Str == "stopped" && sig.NodeID == gql_loaded.ID
  })
  fatalErr(t, err)
}

