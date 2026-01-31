package escpos_test

import (
	"bytes"
	"testing"

	"github.com/dpemmons/mcprinter/internal/escpos"
)

func TestEncodeText(t *testing.T) {
	data := escpos.EncodeText("Hello")

	// Must start with ESC @ (initialize)
	if !bytes.HasPrefix(data, []byte{0x1B, 0x40}) {
		t.Error("missing ESC @ init command")
	}

	// Must contain the text
	if !bytes.Contains(data, []byte("Hello")) {
		t.Error("missing text content")
	}

	// Must end with feed + cut (GS V \x00 = full cut)
	if !bytes.HasSuffix(data, []byte{0x1D, 0x56, 0x00}) {
		t.Error("missing cut command at end")
	}
}

func TestEncodeText_MultipleLines(t *testing.T) {
	data := escpos.EncodeText("Line1\nLine2")

	if !bytes.Contains(data, []byte("Line1\nLine2")) {
		t.Error("missing multi-line text content")
	}
}
