package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dpemmons/mcprinter/cmd"
)

func TestResolveArgs_LiteralText(t *testing.T) {
	items, err := cmd.ResolveArgs([]string{"Hello, world!"})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Type != cmd.ItemText {
		t.Errorf("expected text, got %v", items[0].Type)
	}
	if items[0].Data != "Hello, world!" {
		t.Errorf("got %q", items[0].Data)
	}
}

func TestResolveArgs_TextFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "receipt.txt")
	os.WriteFile(path, []byte("Order #1234"), 0644)

	items, err := cmd.ResolveArgs([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Type != cmd.ItemText {
		t.Fatalf("expected 1 text item, got %d items", len(items))
	}
	if items[0].Data != "Order #1234" {
		t.Errorf("got %q", items[0].Data)
	}
}

func TestResolveArgs_ImageFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "logo.png")
	os.WriteFile(path, []byte("fake png"), 0644)

	items, err := cmd.ResolveArgs([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Type != cmd.ItemImage {
		t.Fatalf("expected 1 image item, got %+v", items)
	}
	if items[0].Path != path {
		t.Errorf("got path %q", items[0].Path)
	}
}

func TestResolveArgs_ImageThenText(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "logo.png")
	os.WriteFile(imgPath, []byte("fake"), 0644)

	items, err := cmd.ResolveArgs([]string{imgPath, "Thank you!"})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	// Images should sort before text
	if items[0].Type != cmd.ItemImage {
		t.Error("expected image first")
	}
	if items[1].Type != cmd.ItemText {
		t.Error("expected text second")
	}
}

func TestResolveArgs_NoArgs(t *testing.T) {
	_, err := cmd.ResolveArgs([]string{})
	if err == nil {
		t.Fatal("expected error for no args")
	}
}
