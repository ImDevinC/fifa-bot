package fifa

import (
	go_fifa "github.com/ImDevinC/go-fifa"
)

var eventsToSkip = map[go_fifa.MatchEvent]bool{
	9:  true, // Ref paused
	10: true, // Ref resumed
	12: true, // Goal attempt
	15: true, // Offside
	16: true, // Corner kick
	17: true, // Blocked shot
	18: true, // Foul
	33: true, // Crossbar
	57: true, // Goalie save
}
