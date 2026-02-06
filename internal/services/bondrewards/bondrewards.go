package bondrewards

import "strings"

const (
	Gift7   = 1 << 0
	Gift14  = 1 << 1
	Gift21  = 1 << 2
	Gift30  = 1 << 3
	Gift45  = 1 << 4
	Gift100 = 1 << 5 // ✅ Secret Gift (Forever Human)
)

type Unlock struct {
	GiftMask int
	GiftName string
}

func TitleForHighestStreak(highest int) string {
	switch {
	case highest >= 100:
		return "\x0304Purrito’s Forever Human 🐾❤️\x0F"

	case highest >= 60:
		return "\x0306Purrito’s Trusted Companion 🐱\x0F"

	case highest >= 30:
		return "\x0303Deeply Bonded Friend 😽\x0F"

	case highest >= 14:
		return "\x0308Warm Purr Companion 🐾\x0F"

	case highest >= 7:
		return "\x0311Getting Purrito’s Trust 🐱\x0F"

	default:
		return "\x0309Just Met Purrito 🐾\x0F"
	}
}

func GiftUnlocks(oldHighest, newHighest int) []Unlock {
	var out []Unlock

	if oldHighest < 7 && newHighest >= 7 {
		out = append(out, Unlock{
			GiftMask: Gift7,
			GiftName: "🐹 Tiny Guinea Pig",
		})
	}

	if oldHighest < 14 && newHighest >= 14 {
		out = append(out, Unlock{
			GiftMask: Gift14,
			GiftName: "🐍 Cute Python",
		})
	}

	if oldHighest < 21 && newHighest >= 21 {
		out = append(out, Unlock{
			GiftMask: Gift21,
			GiftName: "🦜 Noisy Parrot",
		})
	}

	if oldHighest < 30 && newHighest >= 30 {
		out = append(out, Unlock{
			GiftMask: Gift30,
			GiftName: "🐠 Colorful Fish",
		})
	}

	if oldHighest < 45 && newHighest >= 45 {
		out = append(out, Unlock{
			GiftMask: Gift45,
			GiftName: "🐱 Friendly Kitten",
		})
	}

	return out
}

func JoinGifts(list []string) string {
	if len(list) == 0 {
		return "None"
	}
	return strings.Join(list, ", ")
}
