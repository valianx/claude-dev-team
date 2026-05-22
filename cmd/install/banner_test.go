package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Tests: plainBannerLines
// ---------------------------------------------------------------------------

// TestPlainBannerLines_ContainsWordmark verifies the banner includes the
// "team-harness" wordmark.
func TestPlainBannerLines_ContainsWordmark(t *testing.T) {
	found := false
	for _, l := range plainBannerLines() {
		if strings.Contains(l, "team-harness") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected plainBannerLines to contain 'team-harness' wordmark")
	}
}

// TestPlainBannerLines_ContainsTagline verifies the banner includes the tagline.
func TestPlainBannerLines_ContainsTagline(t *testing.T) {
	found := false
	for _, l := range plainBannerLines() {
		if strings.Contains(l, "subagents") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected plainBannerLines to contain tagline with 'subagents'")
	}
}

// TestPlainBannerLines_NoEscapeSequences verifies no ANSI escape sequences
// appear in the plain variant.
func TestPlainBannerLines_NoEscapeSequences(t *testing.T) {
	for i, l := range plainBannerLines() {
		if strings.Contains(l, "\033[") {
			t.Errorf("line %d contains ANSI escape sequence in plain banner: %q", i, l)
		}
	}
}

// TestPlainBannerLines_WidthBound verifies every line fits within 80 columns.
func TestPlainBannerLines_WidthBound(t *testing.T) {
	for i, l := range plainBannerLines() {
		if len(l) > 80 {
			t.Errorf("line %d exceeds 80 cols (%d): %q", i, len(l), l)
		}
	}
}

// TestPlainBannerLines_HeightBound verifies the banner fits within 15 lines.
func TestPlainBannerLines_HeightBound(t *testing.T) {
	lines := plainBannerLines()
	if len(lines) > 15 {
		t.Errorf("banner has %d lines, exceeds 15-line cap", len(lines))
	}
}

// ---------------------------------------------------------------------------
// Tests: printWelcomeBanner — output written to stdout
// ---------------------------------------------------------------------------

// TestPrintWelcomeBanner_ProducesOutput verifies that printWelcomeBanner writes
// at least some bytes to stdout regardless of whether the terminal supports ANSI.
// We redirect stdout to a pipe to capture what is printed.
func TestPrintWelcomeBanner_ProducesOutput(t *testing.T) {
	// Capture stdout by temporarily replacing os.Stdout.
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w

	printWelcomeBanner()

	w.Close()
	os.Stdout = origStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("reading captured output: %v", err)
	}
	r.Close()

	if buf.Len() == 0 {
		t.Error("expected printWelcomeBanner to write output, got empty buffer")
	}
	if !strings.Contains(buf.String(), "team-harness") {
		t.Error("expected output to contain 'team-harness'")
	}
}

// ---------------------------------------------------------------------------
// Tests: pressEnterToExit — non-interactive path is a no-op
// ---------------------------------------------------------------------------

// TestPressEnterToExit_NonInteractiveReturnsImmediately verifies that
// pressEnterToExit does not block or write anything when stdin is not a
// terminal. In the test runner stdin is always non-interactive (piped), so
// this test exercises the guard directly without needing a subprocess.
func TestPressEnterToExit_NonInteractiveReturnsImmediately(t *testing.T) {
	// Verify precondition: stdin is non-interactive in the test runner.
	// If this fails the test environment is unusual (TTY attached to test
	// process stdin) and the test is skipped rather than producing a false failure.
	if isTerminal() {
		t.Skip("stdin appears to be a TTY in this test environment; skipping non-interactive guard test")
	}

	// Capture stdout to assert nothing is printed.
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w

	pressEnterToExit() // must return immediately (no blocking Read)

	w.Close()
	os.Stdout = origStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("reading captured output: %v", err)
	}
	r.Close()

	if buf.Len() != 0 {
		t.Errorf("pressEnterToExit wrote output in non-interactive mode: %q", buf.String())
	}
}
