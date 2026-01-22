package bondrewards

import "strings"

const (
	Gift7  = 1 << 0
	Gift14 = 1 << 1
	Gift30 = 1 << 2
)

type Unlock struct {
	GiftMask int
	GiftName string
}

func TitleForHighestStreak(highest int) string {
	switch {
	case highest >= 100:
		return "Eternal Moonbound"
	case highest >= 60:
		return "Keeper of the Night Purr"
	case highest >= 30:
		return "Lunar Bonded Soul"
	case highest >= 14:
		return "Starlight Companion"
	case highest >= 7:
		return "Moon-Touched Friend"
	default:
		return "Soft New Moon"
	}
}

func GiftUnlocks(oldHighest, newHighest int) []Unlock {
	var out []Unlock
	if oldHighest < 7 && newHighest >= 7 {
		out = append(out, Unlock{GiftMask: Gift7, GiftName: "Pastel Moon Bell"})
	}
	if oldHighest < 14 && newHighest >= 14 {
		out = append(out, Unlock{GiftMask: Gift14, GiftName: "Starlit Ribbon Collar"})
	}
	if oldHighest < 30 && newHighest >= 30 {
		out = append(out, Unlock{GiftMask: Gift30, GiftName: "Lunar Memory Charm"})
	}
	return out
}

func RenderGifts(mask int) []string {
	var gifts []string
	if mask&Gift7 != 0 {
		gifts = append(gifts, "Pastel Moon Bell")
	}
	if mask&Gift14 != 0 {
		gifts = append(gifts, "Starlit Ribbon Collar")
	}
	if mask&Gift30 != 0 {
		gifts = append(gifts, "Lunar Memory Charm")
	}
	return gifts
}

func JoinGifts(list []string) string {
	if len(list) == 0 {
		return "None yet — Purrito is watching quietly"
	}
	return strings.Join(list, ", ")
}
