package tcp

// we receive events and dispatch commands; see below

// these flow from neovim to go
const (
	EventTypeState = "state"
	EventTypeHello = "hello"
)

// these flow from go to neovim
const (
	CommandGoodbye = "goodbye"
)

type Line struct {
	Text  string `json:"text"`
	Width int    `json:"width"`
}

type RenderConfig struct {
	GutterLeft int `json:"gutter_left"`
}

type Event struct {
	Type       string       `json:"type"`
	Pid        int          `json:"pid,omitempty"`
	Top        int          `json:"top"`
	Bot        int          `json:"bot"`
	WinWidth   int          `json:"win_width"`
	WinHeight  int          `json:"win_height"`
	ScreenCols int          `json:"screen_cols"`
	ScreenRows int          `json:"screen_rows"`
	LeftCol    int          `json:"leftcol"`
	Lines      []Line       `json:"lines"`
	Cursor     [2]int       `json:"cursor"` // [row, col] to match nvim_win_get_cursor
	Config     RenderConfig `json:"config"`
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
