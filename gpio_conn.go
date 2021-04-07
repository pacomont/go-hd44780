package hd44780

// Hitachi HD44780U support library

import (
	"github.com/stianeikeland/go-rpio"
	"strings"
	"sync"
	"time"
)

// GPIO4bit - interface to lcd by GPIO in 4-bit mode
type GPIO4bit struct {
	sync.Mutex

	// Mapping GPIO - Data lines
	// Select register pin
	RSPin int // 7
	// Enable bit/pin
	EPin int // 8
	// Data4 bit/pin
	D4Pin int // 25
	// Data5 bit/pin
	D5Pin int // 24
	// Data6 bit/pin
	D6Pin int // 23
	// Data7 bit/pin
	D7Pin int // 18

	// max lines
	Lines int
	// Memory address for each line
	LinesAddr []byte
	// LCD width (number of character in line)
	Width int

	lcdRS rpio.Pin
	lcdE  rpio.Pin
	lcdD4 rpio.Pin
	lcdD5 rpio.Pin
	lcdD6 rpio.Pin
	lcdD7 rpio.Pin

	lastLines []string
	active    bool
}

// NewGPIO4bit create new GPIO4bit structure with some defaults
func NewGPIO4bit() (h *GPIO4bit) {
	h = &GPIO4bit{
		RSPin:     7,
		EPin:      8,
		D4Pin:     25,
		D5Pin:     24,
		D6Pin:     23,
		D7Pin:     18,
		Lines:     4,
		LinesAddr: []byte{lcdLine1, lcdLine2, lcdLine3, lcdLine4},
		Width:     lcdWidth,
	}
	return
}

// Open / initialize LCD interface
func (h *GPIO4bit) Open() (err error) {
	h.Lock()
	defer h.Unlock()

	if h.active {
		return
	}

	if err := rpio.Open(); err != nil {
		return err
	}

	h.lcdRS = initPin(h.RSPin)
	h.lcdE = initPin(h.EPin)
	h.lcdD4 = initPin(h.D4Pin)
	h.lcdD5 = initPin(h.D5Pin)
	h.lcdD6 = initPin(h.D6Pin)
	h.lcdD7 = initPin(h.D7Pin)
	h.lastLines = make([]string, h.Lines, h.Lines)
	h.reset()
	h.active = true

	return
}

// Active return true when interface is working ok
func (h *GPIO4bit) Active() bool {
	return h.active
}

// Reset interface
func (h *GPIO4bit) Reset() {
	h.Lock()
	defer h.Unlock()
	h.reset()
}

func (h *GPIO4bit) reset() {
	// initialize
	h.write4Bits(0x3, lcdCmd)
	time.Sleep(5 * time.Millisecond)
	h.write4Bits(0x3, lcdCmd)
	time.Sleep(120 * time.Microsecond)
	h.write4Bits(0x3, lcdCmd)
	time.Sleep(120 * time.Microsecond)

	h.write4Bits(0x2, lcdCmd)
	time.Sleep(120 * time.Microsecond)

	h.writeByte(0x28, lcdCmd) // Data length, number of lines, font size
	h.writeByte(0x0C, lcdCmd) // Display On,Cursor Off, Blink Off
	h.writeByte(0x06, lcdCmd) // Cursor move direction

	h.writeByte(0x01, lcdCmd) // Clear display
	time.Sleep(5 * time.Millisecond)
}

// Clear display
func (h *GPIO4bit) Clear() {
	h.Lock()
	defer h.Unlock()

	if !h.active {
		return
	}

	h.writeByte(lcdLine1, lcdCmd)
	for i := 0; i < lcdWidth; i++ {
		h.writeByte(' ', lcdChr)
	}
	h.writeByte(lcdLine2, lcdCmd)
	for i := 0; i < lcdWidth; i++ {
		h.writeByte(' ', lcdChr)
	}
	h.writeByte(lcdLine3, lcdCmd)
	for i := 0; i < lcdWidth; i++ {
		h.writeByte(' ', lcdChr)
	}
	h.writeByte(lcdLine4, lcdCmd)
	for i := 0; i < lcdWidth; i++ {
		h.writeByte(' ', lcdChr)
	}
}

// Close interface, clear display.
func (h *GPIO4bit) Close() {
	h.Lock()
	defer h.Unlock()

	if !h.active {
		return
	}

	h.writeByte(lcdLine1, lcdCmd)
	for i := 0; i < lcdWidth; i++ {
		h.writeByte(' ', lcdChr)
	}
	h.writeByte(lcdLine2, lcdCmd)
	for i := 0; i < lcdWidth; i++ {
		h.writeByte(' ', lcdChr)
	}
	h.writeByte(lcdLine3, lcdCmd)
	for i := 0; i < lcdWidth; i++ {
		h.writeByte(' ', lcdChr)
	}
	h.writeByte(lcdLine4, lcdCmd)
	for i := 0; i < lcdWidth; i++ {
		h.writeByte(' ', lcdChr)
	}

	h.writeByte(0x01, lcdCmd) // 000001 Clear display
	time.Sleep(5 * time.Millisecond)

	h.writeByte(0x0C, lcdCmd) // 001000 Display Off

	h.lcdRS.Low()
	h.lcdE.Low()
	h.lcdD4.Low()
	h.lcdD5.Low()
	h.lcdD6.Low()
	h.lcdD7.Low()
	rpio.Close()

	h.active = false
}

// writeByte send byte to lcd
func (h *GPIO4bit) writeByte(bits byte, characterMode bool) {
	if characterMode {
		h.lcdRS.High()
	} else {
		h.lcdRS.Low()
	}

	// High bits
	h.push4bits(bits >> 4)

	// Low bits
	h.push4bits(bits)

	time.Sleep(eDelay)
}

// write4Bits send (lower) 4bits  to lcd
func (h *GPIO4bit) write4Bits(bits byte, characterMode bool) {
	if characterMode {
		h.lcdRS.High()
	} else {
		h.lcdRS.Low()
	}

	h.push4bits(bits)

	time.Sleep(eDelay)
}

// push4bits push 4 bites on data lines
func (h *GPIO4bit) push4bits(bits byte) {
	if bits&0x01 == 0x01 {
		h.lcdD4.High()
	} else {
		h.lcdD4.Low()
	}
	if bits&0x02 == 0x02 {
		h.lcdD5.High()
	} else {
		h.lcdD5.Low()
	}
	if bits&0x04 == 0x04 {
		h.lcdD6.High()
	} else {
		h.lcdD6.Low()
	}
	if bits&0x08 == 0x08 {
		h.lcdD7.High()
	} else {
		h.lcdD7.Low()
	}
	// Toggle 'Enable' pin
	time.Sleep(ePulse)
	h.lcdE.High()
	time.Sleep(ePulse)
	h.lcdE.Low()
	time.Sleep(ePulse)
}

// DisplayLines sends one or more lines separated by \n to lcd
func (h *GPIO4bit) DisplayLines(msg string) {
	for line, text := range strings.Split(msg, "\n") {
		h.Display(line, text)
	}
}

// Display only one line
func (h *GPIO4bit) Display(line int, text string) {
	h.Lock()
	defer h.Unlock()

	if !h.active {
		return
	}

	if line >= h.Lines {
		return
	}

	if len(text) < lcdWidth {
		text = text + strings.Repeat(" ", h.Width-len(text))
	} else {
		text = text[:h.Width]
	}

	// skip not changed lines
	if h.lastLines[line] == text {
		return
	}

	h.lastLines[line] = text

	h.writeByte(h.LinesAddr[line], lcdCmd)

	for c := 0; c < h.Width; c++ {
		h.writeByte(byte(text[c]), lcdChr)
	}
}

func (h *GPIO4bit) SetChar(pos byte, def []byte) {
	if len(def) != 8 {
		panic("invalid def - req 8 bytes")
	}
	h.writeByte(0x40+pos*8, lcdCmd)
	for _, d := range def {
		h.writeByte(d, lcdChr)
	}
}

func (h *GPIO4bit) ToggleBacklight() {
}
