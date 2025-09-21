// Copyright 2025 Favorite Colors MCP Server
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package storage

import (
	"fmt"
	"sync"
)

// ColorStorage manages the favorite colors storage
type ColorStorage struct {
	colors []string
	mutex  sync.RWMutex
}

// NewColorStorage creates a new color storage instance
func NewColorStorage() *ColorStorage {
	return &ColorStorage{
		colors: make([]string, 0),
	}
}

// AddColor adds a color to the favorites list
func (cs *ColorStorage) AddColor(color string) (string, bool) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	// Check if color already exists
	for _, existingColor := range cs.colors {
		if existingColor == color {
			return fmt.Sprintf("Color '%s' is already in your favorites", color), false
		}
	}

	cs.colors = append(cs.colors, color)
	return fmt.Sprintf("Successfully added '%s' to your favorite colors!", color), true
}

// GetColors returns all favorite colors
func (cs *ColorStorage) GetColors() ([]string, string) {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	colors := make([]string, len(cs.colors))
	copy(colors, cs.colors)

	var text string
	if len(colors) == 0 {
		text = "You have no favorite colors yet."
	} else {
		text = fmt.Sprintf("Your favorite colors (%d total):\n", len(colors))
		for i, color := range colors {
			text += fmt.Sprintf("%d. %s\n", i+1, color)
		}
	}

	return colors, text
}

// RemoveColor removes a color from the favorites list
func (cs *ColorStorage) RemoveColor(color string) (string, bool) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	for i, existingColor := range cs.colors {
		if existingColor == color {
			cs.colors = append(cs.colors[:i], cs.colors[i+1:]...)
			return fmt.Sprintf("Successfully removed '%s' from your favorite colors!", color), true
		}
	}

	return fmt.Sprintf("Color '%s' was not found in your favorites", color), false
}

// ClearColors removes all colors from the favorites list
func (cs *ColorStorage) ClearColors() (string, int) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	clearedCount := len(cs.colors)
	cs.colors = []string{}

	return fmt.Sprintf("Successfully cleared %d favorite colors!", clearedCount), clearedCount
}

// Count returns the number of favorite colors
func (cs *ColorStorage) Count() int {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()
	return len(cs.colors)
}
