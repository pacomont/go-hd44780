package hd44780

import "time"

const (
	// Timing constants
	ePulse = 1 * time.Microsecond
	eDelay = 70 * time.Microsecond

	// Some defaults
	lcdWidth = 20 // Maximum characters per line
	lcdChr   = true
	lcdCmd   = false

	lcdLine1 = 0x80+0x00 // LCD RAM address for the 1st line
	lcdLine2 = 0x80+0x40 // LCD RAM address for the 2nd line
	lcdLine3 = 0x80+0x14 // LCD RAM address for the 3nd line
	lcdLine4 = 0x80+0x54 // LCD RAM address for the 4nd line
)
