package playdate

import "errors"

var (
	ErrMenuItemCreate = errors.New("playdate: create menu item")
	ErrMenuTitle      = errors.New("playdate: menu title must not be empty")
	ErrMenuOptions    = errors.New("playdate: menu options must not be empty")
	ErrMenuValue      = errors.New("playdate: menu value is out of range")
)
