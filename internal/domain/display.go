package domain

type DisplayMode uint8

const (
	DisplayModeTerminal DisplayMode = iota
	DisplayModeWebsocket
)
