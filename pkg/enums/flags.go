package enums

// Flag bytes for packets.
const (
	FlagPing byte = 0b00000000
	FlagPong byte = 0b00000001
	FlagEmpt byte = 0b00000010
	FlagData byte = 0b00000011
)
