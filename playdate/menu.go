package playdate

// Language identifies a language supported by the Playdate localization API.
type Language int

const (
	LanguageEnglish Language = iota
	LanguageJapanese
	LanguageSystem
)

// SystemMenu is the optional capability for game-owned System Menu items.
// A game may own at most three items. Remove is safe to call more than once.
type SystemMenu interface {
	AddActionMenuItem(title string, callback func()) (MenuItem, error)
	AddCheckmarkMenuItem(title string, value bool, callback func()) (CheckmarkMenuItem, error)
	AddOptionsMenuItem(title string, options []string, callback func()) (OptionsMenuItem, error)
}

// MenuItem is an owned System Menu item.
type MenuItem interface {
	Title() string
	SetTitle(string) error
	Remove()
}

// CheckmarkMenuItem is a menu item with a boolean value.
type CheckmarkMenuItem interface {
	MenuItem
	Value() bool
	SetValue(bool)
}

// OptionsMenuItem is a menu item whose value is an option index.
type OptionsMenuItem interface {
	MenuItem
	Value() int
	SetValue(int) error
}

// Localization is the optional capability for Playdate system localization.
// Missing keys are reported by the boolean result; fallback text is game policy.
type Localization interface {
	Language() Language
	LocalizedText(key string, language Language) (string, bool)
}
