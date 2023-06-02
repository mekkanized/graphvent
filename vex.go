package main

import (
  "fmt"
  "log"
  "time"
  "errors"
)

type Member struct {
  BaseResource
}

func NewMember(name string) * Member {
  member := &Member{
    BaseResource: NewBaseResource(name, "A Team Member", []Resource{}),
  }

  return member
}

type Team struct {
  BaseResource
  Org string
  Team string
}

func (team * Team) Members() []*Member {
  ret := make([]*Member, len(team.children))
  for idx, member := range(team.children) {
    ret[idx] = member.(*Member)
  }

  return ret
}

func NewTeam(org string, team string, members []*Member) * Team {
  name := fmt.Sprintf("%s%s", org, team)
  description := fmt.Sprintf("Team %s", name)
  resource := &Team{
    BaseResource: NewBaseResource(name, description, make([]Resource, len(members))),
    Org: org,
    Team: team,
  }

  for idx, member := range(members) {
    resource.children[idx] = member
  }

  return resource
}

type Alliance struct {
  BaseResource
}

func NewAlliance(team0 * Team, team1 * Team) * Alliance {
  name := fmt.Sprintf("Alliance %s/%s", team0.Name(), team1.Name())
  description := ""

  resource := &Alliance{
    BaseResource: NewBaseResource(name, description, []Resource{team0, team1}),
  }
  return resource
}

type Match struct {
  BaseEvent
  state string
  control string
  control_start time.Time
}

type Arena struct {
  BaseResource
  connected bool
}

func NewVirtualArena(name string) * Arena {
  arena := &Arena{
    BaseResource: NewBaseResource(name, "A virtual vex arena", []Resource{}),
    connected: false,
  }

  return arena
}

func (arena * Arena) Lock(event Event) error {
  if arena.connected == false {
    log.Printf("ARENA NOT CONNECTED: %s", arena.Name())
    error_str := fmt.Sprintf("%s is not connected, cannot lock", arena.Name())
    return errors.New(error_str)
  }
  return arena.lock(event)
}

func (arena * Arena) Update(signal GraphSignal) error {
  log.Printf("UPDATE Arena %s: %+v", arena.Name(), signal)

  arena.BaseResource.Update(signal)

  if arena.connected == true {
    arena.signal <- signal
  }

  return nil
}

func (arena * Arena) Connect(abort chan error) bool {
  log.Printf("Connecting %s", arena.Name())
  go func(arena * Arena, abort chan error) {
    owner := arena.Owner()
    arena.connected = true
    update_str := fmt.Sprintf("VIRTUAL_ARENA connected: %s", arena.Name())
    signal := NewSignal(arena, "arena_connected")
    signal.description = update_str
    arena.Update(signal)
    log.Printf("VIRTUAL_ARENA goroutine starting: %s", arena.Name())
    for true {
      select {
      case <- abort:
        log.Printf("Virtual arena %s aborting", arena.Name())
        break
      case update := <- arena.signal:
        log.Printf("%s update: %s", arena.Name(), update)
        new_owner := arena.Owner()
        if new_owner != owner {
          log.Printf("NEW_OWNER for %s", arena.Name())
          if new_owner != nil {
            log.Printf("new: %s", new_owner.Name())
          } else {
            log.Printf("new: nil")
          }

          if owner != nil {
            log.Printf("old: %s", owner.Name())
          } else {
            log.Printf("old: nil")
          }

          owner = new_owner
          if owner != nil {
          } else {
          }
        }
      }
    }
  }(arena, abort)
  return true
}

const start_slack = 3000 * time.Millisecond

func NewMatch(alliance0 * Alliance, alliance1 * Alliance, arena * Arena) * Match {
  name := fmt.Sprintf("Match: %s vs. %s on %s", alliance0.Name(), alliance1.Name(), arena.Name())
  description := "A vex match"

  match := &Match{
    BaseEvent: NewBaseEvent(name, description, []Resource{alliance0, alliance1, arena}),
    state: "init",
    control: "init",
    control_start: time.UnixMilli(0),
  }
  match.LockDone()

  match.actions["start"] = func() (string, error) {
    log.Printf("STARTING_MATCH %s", match.Name())
    match.control = "none"
    match.state = "scheduled"
    return "wait", nil
  }

  match.actions["queue_autonomous"] = func() (string, error) {
    match.control = "none"
    match.state = "autonomous_queued"
    match.control_start = time.Now().Add(start_slack)
    return "wait", nil
  }

  match.actions["start_autonomous"] = func() (string, error) {
    match.control = "autonomous"
    match.state = "autonomous_running"
    return "wait", nil
  }

  return match
}
