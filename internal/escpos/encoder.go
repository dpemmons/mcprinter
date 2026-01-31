package escpos

// ESC/POS command bytes
var (
	CmdInit    = []byte{0x1B, 0x40}       // ESC @ - initialize printer
	CmdFeed    = []byte{0x1B, 0x64, 0x04} // ESC d 4 - feed 4 lines
	CmdFullCut = []byte{0x1D, 0x56, 0x00} // GS V 0 - full cut
)

// EncodeText converts a text string into ESC/POS bytes with init and cut.
func EncodeText(text string) []byte {
	var buf []byte
	buf = append(buf, CmdInit...)
	buf = append(buf, []byte(text)...)
	buf = append(buf, '\n')
	buf = append(buf, CmdFeed...)
	buf = append(buf, CmdFullCut...)
	return buf
}
