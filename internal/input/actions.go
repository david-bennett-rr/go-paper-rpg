package input

import (
	einput "github.com/quasilyte/ebitengine-input"
)

const (
	ActionMoveUp einput.Action = iota
	ActionMoveDown
	ActionMoveLeft
	ActionMoveRight
	ActionConfirm
	ActionCancel
	ActionMenu
	ActionPartner
)
