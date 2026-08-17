package playdate

// EpochTime is an instant measured from midnight, January 1, 2000. Milliseconds
// is in the range 0..999.
type EpochTime struct {
	Seconds      uint32
	Milliseconds uint32
}

// DateTime is an owned calendar value. Weekday uses 1 for Monday through 7 for
// Sunday. DateTimeToEpoch derives the instant from the other fields and ignores
// Weekday.
type DateTime struct {
	Year    uint16
	Month   uint8
	Day     uint8
	Weekday uint8
	Hour    uint8
	Minute  uint8
	Second  uint8
}

// SystemInfo is a copied snapshot of runtime and game compatibility metadata.
// Version values use the official numeric encoding; for example, 20705 is
// version 2.7.5.
type SystemInfo struct {
	OSVersion  uint32
	Language   Language
	PDXVersion uint32
}

// SystemEnvironment is the optional offline clock, calendar, and device-info
// capability. Epoch values use the Playdate epoch beginning January 1, 2000.
type SystemEnvironment interface {
	CurrentEpochTime() EpochTime
	EpochToDateTime(epoch uint32) DateTime
	DateTimeToEpoch(dateTime DateTime) (uint32, error)
	ResetElapsedTime()
	ElapsedTime() float32
	SystemInfo() SystemInfo
}
