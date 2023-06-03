package main

import (
  "log"
  "runtime/pprof"
  "time"
  "os"
)

func fake_team(org string, id string, names []string) (*Team, []*Member) {
  members := []*Member{}
  for _, name := range(names) {
    members = append(members, NewMember(name))
  }
  team := NewTeam(org, id, members)
  return team, members
}

func fake_data() * EventManager {
  resources := []Resource{}

  teams := []*Team{}
  t1,  m1  := fake_team("6659", "A", []string{"jimmy"})
  t2,  m2  := fake_team("6659", "B", []string{"timmy"})
  t3,  m3  := fake_team("6659", "C", []string{"grace"})
  t4,  m4  := fake_team("6659", "D", []string{"jeremy"})
  t5,  m5  := fake_team("210",  "W", []string{"bobby"})
  t6,  m6  := fake_team("210",  "X", []string{"toby"})
  t7,  m7  := fake_team("210",  "Y", []string{"jennifer"})
  t8,  m8  := fake_team("210",  "Z", []string{"emily"})
  //t9,  m9  := fake_team("666",  "A", []string{"jimmy"})
  //t10, m10 := fake_team("666",  "B", []string{"timmy"})
  //t11, m11 := fake_team("666",  "C", []string{"grace"})
  //t12, m12 := fake_team("666",  "D", []string{"jeremy"})
  //t13, m13 := fake_team("315",  "W", []string{"bobby"})
  //t14, m14 := fake_team("315",  "X", []string{"toby"})
  //t15, m15 := fake_team("315",  "Y", []string{"jennifer"})
  //t16, m16 := fake_team("315",  "Z", []string{"emily"})

  teams = append(teams, t1)
  teams = append(teams, t2)
  teams = append(teams, t3)
  teams = append(teams, t4)
  teams = append(teams, t5)
  teams = append(teams, t6)
  teams = append(teams, t7)
  teams = append(teams, t8)
  //teams = append(teams, t9)
  //teams = append(teams, t10)
  //teams = append(teams, t11)
  //teams = append(teams, t12)
  //teams = append(teams, t13)
  //teams = append(teams, t14)
  //teams = append(teams, t15)
  //teams = append(teams, t16)

  resources = append(resources, m1[0])
  resources = append(resources, m2[0])
  resources = append(resources, m3[0])
  resources = append(resources, m4[0])
  resources = append(resources, m5[0])
  resources = append(resources, m6[0])
  resources = append(resources, m7[0])
  resources = append(resources, m8[0])
  //resources = append(resources, m9[0])
  //resources = append(resources, m10[0])
  //resources = append(resources, m11[0])
  //resources = append(resources, m12[0])
  //resources = append(resources, m13[0])
  //resources = append(resources, m14[0])
  //resources = append(resources, m15[0])
  //resources = append(resources, m16[0])

  arenas := []*Arena{}
  arenas = append(arenas, NewVirtualArena("Arena 1"))
  arenas = append(arenas, NewVirtualArena("Arena 2"))

  for _, arena := range arenas {
    resources = append(resources, arena)
  }

  for _, team := range teams {
    resources = append(resources, team)
  }

  alliances := []*Alliance{}
  alliances = append(alliances, NewAlliance(t1, t2))
  alliances = append(alliances, NewAlliance(t3, t4))
  alliances = append(alliances, NewAlliance(t5, t6))
  alliances = append(alliances, NewAlliance(t7, t8))

  for _, alliance := range alliances {
    resources = append(resources, alliance)
  }


  root_event := NewEventQueue("root_event", "", []Resource{})
  stay_resource := NewResource("stay_resource", "", []Resource{})
  resources = append(resources, stay_resource)
  stay_event := NewEvent("stay_event", "", []Resource{stay_resource})
  LockResource(stay_resource, stay_event)
  event_manager := NewEventManager(root_event, resources)
  event_manager.AddEvent(root_event, stay_event, NewEventQueueInfo(1))

  go func(alliances []*Alliance, arenas []*Arena, event_manager * EventManager) {
    for i, alliance := range(alliances) {
      for j, alliance2 := range(alliances) {
        if j != i {
          if alliance.Children()[0] == alliance2.Children()[0] || alliance.Children()[0] == alliance2.Children()[1] || alliance.Children()[1] == alliance2.Children()[0] || alliance.Children()[1] == alliance2.Children()[1] {
          } else {
            for arena_idx := 0; arena_idx < len(arenas); arena_idx++ {
              match := NewMatch(alliance, alliance2, arenas[arena_idx])
              log.Printf("Adding %s", match.Name())
              err := event_manager.AddEvent(root_event, match, NewEventQueueInfo(i))
              if err != nil {
                log.Printf("Error adding %s: %s", match.Name(), err)
              }
            }
          }
        }
      }
    }
  }(alliances, arenas, event_manager)

  return event_manager
}

func main() {
  go func() {
    time.Sleep(5 * time.Second)
    pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
  }()

  event_manager := fake_data()
  log.Printf("Starting event_manager")
  err := event_manager.Run()
  if err != nil {
    log.Printf("Error running event_manager: %s", err)
  } else {
    log.Printf("Finished event_manager")
  }
}
