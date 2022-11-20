package fifa

import (
	go_fifa "github.com/ImDevinC/go-fifa"
)

var eventsToSkip = map[go_fifa.MatchEvent]bool{
	go_fifa.Substitution: true, // Substitution
	go_fifa.MatchPaused:  true, // Ref paused
	go_fifa.MatchResumed: true, // Ref resumed
	go_fifa.GoalAttempt:  true, // Goal attempt
	go_fifa.Offside:      true, // Offside
	go_fifa.CornerKick:   true, // Corner kick
	go_fifa.ShotBlocked:  true, // Blocked shot
	go_fifa.Foul:         true, // Foul
	go_fifa.CoinToss:     true, // Coin BlockedShottoss
	go_fifa.Unknown3:     true, // Unknown
	go_fifa.ThrowIn:      true, // Throw in
	go_fifa.Clearance:    true, // Clearance
	go_fifa.Unknown2:     true, // No idea
	go_fifa.Crossbar:     true, // Crossbar
	go_fifa.FreeKickPost: true, // Free-kick post
	go_fifa.GoalieSaved:  true, // Goalie save
	go_fifa.Unknown:      true, // Placeholder?
}

var flagEmojis = map[string]string{
	"ARG": ":flag-ar:",
	"AUS": ":flag-au:",
	"BEL": ":flag-be:",
	"BRA": ":flag-br:",
	"CAN": ":flag-ca:",
	"CHI": ":flag-cl:",
	"CHN": ":flag-cn:",
	"CMR": ":flag-cm:",
	"COL": ":flag-co:",
	"CRC": ":flag-cr:",
	"CRO": ":flag-hr:",
	"DEN": ":flag-dk:",
	"EGY": ":flag-eg:",
	"ENG": ":flag-england:",
	"ESP": ":flag-es:",
	"FRA": ":flag-fr:",
	"GER": ":flag-de:",
	"IRN": ":flag-ir:",
	"ISL": ":flag-is:",
	"ITA": ":flag-it:",
	"JAM": ":flag-jm:",
	"JPN": ":flag-jp:",
	"KOR": ":flag-kr:",
	"KSA": ":flag-sa:",
	"MAR": ":flag-ma:",
	"MEX": ":flag-mx:",
	"NED": ":flag-nl:",
	"NGA": ":flag-ng:",
	"NOR": ":flag-no:",
	"NZL": ":flag-nz:",
	"PAN": ":flag-pa:",
	"PER": ":flag-pe:",
	"POL": ":flag-pl:",
	"POR": ":flag-pt:",
	"RSA": ":flag-za:",
	"RUS": ":flag-ru:",
	"SCO": ":flag-scotland:",
	"SEN": ":flag-sn:",
	"SRB": ":flag-rs:",
	"SUI": ":flag-ch:",
	"SWE": ":flag-se:",
	"THA": ":flag-th:",
	"TUN": ":flag-tn:",
	"URU": ":flag-uy:",
	"ZAF": ":flag-za:",
}
