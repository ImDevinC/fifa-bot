package models

import "encoding/json"

type Match struct {
	Id             string   `redis:"id,omitempty"`
	Events         []string `redis:"events,omitempty"`
	Expiration     int      `redis:"expiration,omitempty"`
	CompetitionId  string   `redis:"competition_id"`
	SeasonId       string   `redis:"season_id"`
	StageId        string   `redis:"stage_id"`
	MatchId        string   `redis:"match_id"`
	LastEvent      string   `redis:"last_event,omitempty"`
	HomeTeamName   string   `redis:"home_team_name,omitempty"`
	AwayTeamName   string   `redis:"away_team_name,omitempty"`
	HomeTeamAbbrev string   `redis:"home_team_abbrev,omitempty"`
	AwayTeamAbbrev string   `redis:"away_team_abbrev,omitempty"`
}

func (m Match) MarshalBinary() ([]byte, error) {
	return json.Marshal(m)
}
