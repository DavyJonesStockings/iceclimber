package rpc

const (
	EventTypeState = "state"
)

type Line struct {
	Text  string `json:"text"`
	Width int    `json:"width"`
}

type Event struct {
	Type      string `json:"type"`
	Top       int    `json:"top"`
	Bot       int    `json:"bot"`
	WinWidth  int    `json:"win_width"`
	WinHeight int    `json:"win_height"`
	LeftCol   int    `json:"leftcol"`
	Lines     []Line `json:"lines"`
	Cursor    [2]int `json:"cursor"` // [row, col] to match nvim_win_get_cursor
}

const (
	CommandScrollLeft  = "scroll_left"
	CommandScrollRight = "scroll_right"
	CommandCursorMove  = "cursor_move"
)

type Command struct {
	Type  string `json:"type"`
	Count int    `json:"count,omitempty"`
	Line  int    `json:"line,omitempty"`
	Col   int    `json:"col,omitempty"`
}
