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
	49: true, // Free-kick post
	57: true, // Goalie save
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
