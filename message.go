package graphvent

import (
  "time"
  "crypto/ed25519"
  "crypto/rand"
  "crypto"
)

type AuthInfo struct {
  // The Node that issued the authorization
  Identity ed25519.PublicKey

  // Time the authorization was generated
  Start time.Time

  // Signature of Start + Principal with Identity private key
  Signature []byte
}

type AuthorizationToken struct {
  AuthInfo

  // The private key generated by the client, encrypted with the servers public key
  KeyEncrypted []byte
}

type ClientAuthorization struct {
  AuthInfo

  // The private key generated by the client
  Key ed25519.PrivateKey
}

// Authorization structs can be passed in a message that originated from a different node than the sender
type Authorization struct {
  AuthInfo

  // The public key generated for this authorization
  Key ed25519.PublicKey
}

type Message struct {
  Dest NodeID
  Source ed25519.PublicKey

  Authorization *Authorization

  Signal Signal
  Signature []byte
}

type Messages []*Message
func (msgs Messages) Add(ctx *Context, dest NodeID, source *Node, authorization *ClientAuthorization, signal Signal) Messages {
  msg, err := NewMessage(ctx, dest, source, authorization, signal)
  if err != nil {
    panic(err)
  } else {
    msgs = append(msgs, msg)
  }
  return msgs
}

func NewMessages(ctx *Context, dest NodeID, source *Node, authorization *ClientAuthorization, signals... Signal) Messages {
  messages := Messages{}
  for _, signal := range(signals) {
    messages = messages.Add(ctx, dest, source, authorization, signal)
  }
  return messages
}

func NewMessage(ctx *Context, dest NodeID, source *Node, authorization *ClientAuthorization, signal Signal) (*Message, error) {
  signal_ser, err := SerializeAny(ctx, signal)
  if err != nil {
    return nil, err
  }

  signal_chunks, err := signal_ser.Chunks()
  if err != nil {
    return nil, err
  }

  dest_ser, err := dest.MarshalBinary()
  if err != nil {
    return nil, err
  }
  source_ser, err := source.ID.MarshalBinary()
  if err != nil {
    return nil, err
  }
  sig_data := append(dest_ser, source_ser...)
  sig_data = append(sig_data, signal_chunks.Slice()...)
  var message_auth *Authorization = nil
  if authorization != nil {
    sig_data = append(sig_data, authorization.Signature...)
    message_auth = &Authorization{
      authorization.AuthInfo,
      authorization.Key.Public().(ed25519.PublicKey),
    }
  }

  sig, err := source.Key.Sign(rand.Reader, sig_data, crypto.Hash(0))
  if err != nil {
    return nil, err
  }

  return &Message{
    Dest: dest,
    Source: source.Key.Public().(ed25519.PublicKey),
    Authorization: message_auth,
    Signal: signal,
    Signature: sig,
  }, nil
}
