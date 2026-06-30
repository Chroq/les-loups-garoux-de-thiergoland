package domain

type DisplayMode uint8

const (
	DisplayModeTerminal DisplayMode = iota
	DisplayModeWebsocket

	DisplayModeLabelTerminal  = "terminal"
	DisplayModeLabelWebsocket = "websocket"
)

func (d DisplayMode) String() string {
	switch d {
	case DisplayModeTerminal:
		return DisplayModeLabelTerminal
	case DisplayModeWebsocket:
		return DisplayModeLabelWebsocket
	default:
		return "unknown"
	}
}
