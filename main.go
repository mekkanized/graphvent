package main

import (
  "log"
)

func fake_data() * EventManager {
  event_manager := NewEventManager()

  teams := []*Team{}
  teams = append(teams, NewTeam("6659", "A", []string{"jimmy"}))
  teams = append(teams, NewTeam("6659", "B", []string{"timmy"}))
  teams = append(teams, NewTeam("6659", "C", []string{"grace"}))
  teams = append(teams, NewTeam("6659", "D", []string{"jeremy"}))
  teams = append(teams, NewTeam("210",  "W", []string{"bobby"}))
  teams = append(teams, NewTeam("210",  "X", []string{"toby"}))
  teams = append(teams, NewTeam("210",  "Y", []string{"jennifer"}))
  teams = append(teams, NewTeam("210",  "Z", []string{"emily"}))
  teams = append(teams, NewTeam("315",  "W", []string{"bobby"}))
  teams = append(teams, NewTeam("315",  "X", []string{"toby"}))
  teams = append(teams, NewTeam("315",  "Y", []string{"jennifer"}))
  teams = append(teams, NewTeam("315",  "Z", []string{"emily"}))

  for _, team := range teams {
    err := event_manager.AddResource(team)
    if err != nil {
      log.Print(err)
    }
  }


  alliances := []Resource{}
  for i, team := range teams[:len(teams)-1] {
    for _, team2 := range teams[i+1:] {
      alliance := NewAlliance(team, team2)
      alliances = append(alliances, alliance)
      err := event_manager.AddResource(alliance)
      if err != nil {
        log.Print(err)
      }
    }
  }

  return event_manager
}

func main() {
  event_manager := fake_data()
  log.Printf("Starting event_manager: %+v", event_manager)
}