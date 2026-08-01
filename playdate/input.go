package playdate

// Buttons is a bit set of Playdate controls.
type Buttons uint8

const (
	ButtonLeft Buttons = 1 << iota
	ButtonRight
	ButtonUp
	ButtonDown
	ButtonB
	ButtonA
)

// Has reports whether every requested button is present in the set.
func (buttons Buttons) Has(requested Buttons) bool { return buttons&requested == requested }

// Input is the immutable input snapshot for the current frame.
type Input struct {
	Buttons              Buttons
	Pressed              Buttons
	Released             Buttons
	Held                 Buttons
	CrankAngle           float32
	CrankDelta           float32
	CrankDocked          bool
	CrankDockedThisFrame bool
	CrankUndocked        bool
	DeltaSeconds         float32
}

// InputReader exposes the snapshot captured at the start of the current frame.
type InputReader interface {
	Input() Input
}
