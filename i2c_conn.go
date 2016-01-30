package hd44780

import (
	"strings"
	"sync"

	"github.com/kidoman/embd"
	"github.com/kidoman/embd/controller/hd44780"

	// load only rpi
	_ "github.com/kidoman/embd/host/rpi"
)

// I2C4bit allow communicate wit HD44780 via I2C in 4bit mode
type I2C4bit struct {
	sync.Mutex

	// max lines
	Lines int
	// LCD width (number of character in line)
	Width int

	// i2c address
	addr      byte
	lastLines []string
	active    bool

	hd        *hd44780.HD44780
	backlight bool
}

// NewI2C4bit create new I2C4bit structure with some defaults
func NewI2C4bit(addr byte) (h *I2C4bit) {
	h = &I2C4bit{
		Lines: 2,
		addr:  addr,
		Width: lcdWidth,
	}
	return
}

// Open / initialize LCD interface
func (h *I2C4bit) Open() (err error) {
	h.Lock()
	defer h.Unlock()

	if h.active {
		return
	}

	if err := embd.InitI2C(); err != nil {
		panic(err)
	}

	bus := embd.NewI2CBus(1)

	h.hd, err = hd44780.NewI2C(
		bus,
		h.addr,
		hd44780.PCF8574PinMap,
		hd44780.RowAddress16Col,
		hd44780.TwoLine,
		hd44780.BlinkOff,
	)
	if err != nil {
		return err
	}

	h.lastLines = make([]string, h.Lines, h.Lines)
	h.reset()

	h.hd.BacklightOn()
	h.backlight = true

	h.active = true

	return
}

// Active return true when interface is working ok
func (h *I2C4bit) Active() bool {
	return h.active
}

// Reset interface
func (h *I2C4bit) Reset() {
	h.Lock()
	defer h.Unlock()
	h.hd.Clear()
}

func (h *I2C4bit) reset() {
	// clear the display
	h.hd.Clear()
}

// Close interface, clear display.
func (h *I2C4bit) Close() {
	h.Lock()
	defer h.Unlock()

	if !h.active {
		return
	}
	h.hd.BacklightOff()
	h.backlight = false
	h.hd.Clear()
	embd.CloseI2C()
	h.active = false
}

// DisplayLines sends one or more lines separated by \n to lcd
func (h *I2C4bit) DisplayLines(msg string) {
	for line, text := range strings.Split(msg, "\n") {
		h.Display(line, text)
	}
}

// Display only one line
func (h *I2C4bit) Display(line int, text string) {
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

	h.hd.SetCursor(0, line)
	for _, c := range text {
		h.hd.WriteChar(byte(c))
	}
}

func (h *I2C4bit) ToggleBacklight() {
	if !h.active {
		return
	}
	if h.backlight {
		h.hd.BacklightOff()
	} else {
		h.hd.BacklightOn()
	}
	h.backlight = !h.backlight

}
