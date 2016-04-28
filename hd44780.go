package hd44780

// Hitachi HD44780U support library

type HD44780 interface {
	Open()
	Reset()
	Close()
	Clear()
	Display(text string)
	Active() bool
	SetChar(pos byte, def []byte)
}
