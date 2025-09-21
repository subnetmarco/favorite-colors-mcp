package storage

import (
	"fmt"
	"strings"
	"testing"
)

func TestColorStorage_AddColor(t *testing.T) {
	cs := NewColorStorage()

	// Test adding a new color
	message, added := cs.AddColor("blue")
	if !added {
		t.Error("Expected color to be added")
	}
	if !strings.Contains(message, "Successfully added") {
		t.Errorf("Expected success message, got: %s", message)
	}

	// Test adding duplicate color
	message, added = cs.AddColor("blue")
	if added {
		t.Error("Expected duplicate color not to be added")
	}
	if !strings.Contains(message, "already in your favorites") {
		t.Errorf("Expected duplicate message, got: %s", message)
	}

	// Verify count
	if cs.Count() != 1 {
		t.Errorf("Expected 1 color, got %d", cs.Count())
	}
}

func TestColorStorage_GetColors(t *testing.T) {
	cs := NewColorStorage()

	// Test empty storage
	colors, text := cs.GetColors()
	if len(colors) != 0 {
		t.Errorf("Expected empty colors list, got %v", colors)
	}
	if !strings.Contains(text, "no favorite colors yet") {
		t.Errorf("Expected empty message, got: %s", text)
	}

	// Add some colors
	cs.AddColor("red")
	cs.AddColor("blue")

	// Test non-empty storage
	colors, text = cs.GetColors()
	if len(colors) != 2 {
		t.Errorf("Expected 2 colors, got %d", len(colors))
	}
	if !strings.Contains(text, "2 total") {
		t.Errorf("Expected count in message, got: %s", text)
	}
	if !strings.Contains(text, "red") || !strings.Contains(text, "blue") {
		t.Errorf("Expected colors in message, got: %s", text)
	}
}

func TestColorStorage_RemoveColor(t *testing.T) {
	cs := NewColorStorage()

	// Test removing from empty storage
	message, removed := cs.RemoveColor("nonexistent")
	if removed {
		t.Error("Expected color not to be removed from empty storage")
	}
	if !strings.Contains(message, "was not found") {
		t.Errorf("Expected not found message, got: %s", message)
	}

	// Add a color and remove it
	cs.AddColor("green")
	message, removed = cs.RemoveColor("green")
	if !removed {
		t.Error("Expected color to be removed")
	}
	if !strings.Contains(message, "Successfully removed") {
		t.Errorf("Expected success message, got: %s", message)
	}

	// Verify it's gone
	if cs.Count() != 0 {
		t.Errorf("Expected 0 colors after removal, got %d", cs.Count())
	}
}

func TestColorStorage_ClearColors(t *testing.T) {
	cs := NewColorStorage()

	// Test clearing empty storage
	message, count := cs.ClearColors()
	if count != 0 {
		t.Errorf("Expected 0 cleared from empty storage, got %d", count)
	}

	// Add colors and clear
	cs.AddColor("red")
	cs.AddColor("blue")
	cs.AddColor("green")

	message, count = cs.ClearColors()
	if count != 3 {
		t.Errorf("Expected 3 colors cleared, got %d", count)
	}
	if !strings.Contains(message, "Successfully cleared 3") {
		t.Errorf("Expected clear message with count, got: %s", message)
	}

	// Verify all gone
	if cs.Count() != 0 {
		t.Errorf("Expected 0 colors after clear, got %d", cs.Count())
	}
}

func TestColorStorage_Concurrency(t *testing.T) {
	cs := NewColorStorage()

	// Test concurrent access
	done := make(chan bool, 10)

	// Add colors concurrently
	for i := 0; i < 10; i++ {
		go func(i int) {
			cs.AddColor(fmt.Sprintf("color-%d", i))
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have 10 unique colors
	if cs.Count() != 10 {
		t.Errorf("Expected 10 colors after concurrent adds, got %d", cs.Count())
	}
}

func BenchmarkColorStorage_AddColor(b *testing.B) {
	cs := NewColorStorage()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cs.AddColor(fmt.Sprintf("color-%d", i))
	}
}

func BenchmarkColorStorage_GetColors(b *testing.B) {
	cs := NewColorStorage()

	// Pre-populate with colors
	for i := 0; i < 100; i++ {
		cs.AddColor(fmt.Sprintf("color-%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cs.GetColors()
	}
}
