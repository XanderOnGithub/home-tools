package discord

// Poll packs a fresh install starts with (decision #47). They're only
// defaults: once in config.json, they're edited (or removed) in the UI
// like any other pack. Written with Xander for the group.

// DefaultPollPacks returns fresh copies (callers may change them).
func DefaultPollPacks() []PollPack {
	return []PollPack{
		{Name: "Overwatch", Enabled: true, Questions: overwatchPack()},
	}
}

func overwatchPack() []PollQuestion {
	q := func(question string, answers ...string) PollQuestion {
		return PollQuestion{Question: question, Answers: answers}
	}
	const (
		roleSwap = "Which Role Swap Would You Rather See?"
		ult      = "Which Ultimate Change Would You Rather See?"
		rework   = "Which Hero Should Get Fully Reworked?"
	)
	return []PollQuestion{
		q(roleSwap, "Mei becomes a Tank", "Zarya becomes a DPS"),
		q(rework, "Tracer", "Genji"),
		q(ult, "Deadeye becomes a rapid-fire dual revolver ult", "RIP-Tire becomes a giant bomb Junkrat rides"),
		q(roleSwap, "Mercy becomes a DPS", "Soldier: 76 becomes a Support"),
		q("Which Tank Should Lose Their Barrier?", "Reinhardt loses his shield", "Winston loses his bubble"),
		q(ult, "Dragonstrike becomes a one-shot giant arrow", "Dragonblade becomes a dragon that fights with Genji"),
		q(roleSwap, "Bastion becomes a Tank", "Orisa becomes a DPS"),
		q(ult, "Sound Barrier becomes a 3x speed boost", "Transcendence makes the team invisible"),
		q(roleSwap, "Widowmaker becomes a healing sniper Support", "Ana becomes a DPS with no healing"),
		q("Which Old Kit Should Come Back?", "Roadhog’s one-shot combo", "Doomfist’s DPS kit"),
		q(roleSwap, "Symmetra becomes a Tank", "Torbjörn becomes a Support"),
		q(ult, "D.Va’s mech turns into a turret for her ult", "Minefield becomes a giant rolling hamster ball"),
		q(rework, "Lúcio", "Ana"),
		q(roleSwap, "Reaper becomes a Tank", "Junker Queen becomes a DPS"),
		q(ult, "Kitsune Rush becomes a fox that bites enemies", "Brigitte’s Rally becomes a horse charge"),
		q("Which Ability Change Would You Rather See?", "Roadhog’s hook pulls him to enemies", "Hammond gets a De-Mech state like D.Va"),
		q(roleSwap, "Cassidy becomes a Support", "Baptiste becomes a DPS"),
		q(ult, "Death Blossom becomes a draining wraith form", "Barrage becomes one giant nuke rocket"),
		q(roleSwap, "Lifeweaver becomes a Tank", "Sigma becomes a Support"),
		q(rework, "Mercy", "Reinhardt"),
		q("Which New Hero Should Be Removed Forever?", "Doctrine", "D.Mon", "Shion", "Sierra", "Jetpack Cat"),
	}
}
