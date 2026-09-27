package config

// Bards may cast while singing. Below BardSingCastFreeTier the spell's
// damage or healing is cut by BardSingCastPenalty; from that tier on the
// song costs the spell nothing.
var (
	BardSingCastPenalty  = .2
	BardSingCastFreeTier = 7
)

// BardSingCastMod is the multiplier applied to a cast spell's damage or
// healing. It is below 1 only for a low-tier bard who is singing.
func BardSingCastMod(class int, tier int, singing bool) float64 {
	if class != BARD || !singing || tier >= BardSingCastFreeTier {
		return 1
	}
	return 1 - BardSingCastPenalty
}
