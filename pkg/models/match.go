package models

import (
	"encoding/json"
	"fmt"
)

type Match struct {
	Id             string   `json:"id,omitempty" redis:"id,omitempty"`
	Events         []string `json:"events,omitempty" redis:"events,omitempty"`
	Expiration     int      `json:"expiration,omitempty" redis:"expiration,omitempty"`
	CompetitionId  string   `json:"competition_id" redis:"competition_id"`
	SeasonId       string   `json:"season_id" redis:"season_id"`
	StageId        string   `json:"stage_id" redis:"stage_id"`
	MatchId        string   `json:"match_id" redis:"match_id"`
	LastEvent      string   `json:"last_event,omitempty" redis:"last_event,omitempty"`
	HomeTeamName   string   `json:"home_team_name,omitempty" redis:"home_team_name,omitempty"`
	AwayTeamName   string   `json:"away_team_name,omitempty" redis:"away_team_name,omitempty"`
	HomeTeamAbbrev string   `json:"home_team_abbrev,omitempty" redis:"home_team_abbrev,omitempty"`
	AwayTeamAbbrev string   `json:"away_team_abbrev,omitempty" redis:"away_team_abbrev,omitempty"`
}

func (m *Match) GetMap() (map[string]any, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return map[string]any{}, fmt.Errorf("failed to marshal match into bytes. %w", err)
	}
	result := map[string]any{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return map[string]any{}, fmt.Errorf("failed to unmarshal match to map. %w", err)
	}
	eventsJSON, err := json.Marshal(m.Events)
	if err != nil {
		return map[string]any{}, fmt.Errorf("failed to marshal events. %w", err)
	}
	result["events"] = string(eventsJSON)
	return result, nil
}

func MatchFromRedis(data map[string]string) (Match, error) {
	match := Match{}
	if val, exists := data["id"]; exists {
		match.Id = val
	}
	if val, exists := data["competition_id"]; exists {
		match.CompetitionId = val
	}
	if val, exists := data["season_id"]; exists {
		match.SeasonId = val
	}
	if val, exists := data["stage_id"]; exists {
		match.StageId = val
	}
	if val, exists := data["match_id"]; exists {
		match.MatchId = val
	}
	if val, exists := data["last_event"]; exists {
		match.LastEvent = val
	}
	if val, exists := data["home_team_name"]; exists {
		match.HomeTeamName = val
	}
	if val, exists := data["home_team_abbrev"]; exists {
		match.HomeTeamAbbrev = val
	}
	if val, exists := data["away_team_name"]; exists {
		match.AwayTeamName = val
	}
	if val, exists := data["away_team_abbrev"]; exists {
		match.AwayTeamAbbrev = val
	}
	if val, exists := data["events"]; exists {
		err := json.Unmarshal([]byte(val), &match.Events)
		if err != nil {
			return Match{}, fmt.Errorf("failed to unmarshal events. %w", err)
		}
	}
	return match, nil
}
