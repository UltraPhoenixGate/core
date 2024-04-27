package hub_test

import (
	"testing"
	"ultraphx-core/internal/hub"
)

func TestGetTopicPermission(t *testing.T) {
	t.Run("empty topic", func(t *testing.T) {
		if got, want := hub.GetTopicPermission(""), ""; got != want {
			t.Errorf("GetTopicPermission() = %v, want %v", got, want)
		}
	})

	t.Run("topic with wildcard", func(t *testing.T) {
		if got, want := hub.GetTopicPermission("a::#"), "a"; got != want {
			t.Errorf("GetTopicPermission() = %v, want %v", got, want)
		}
	})

	t.Run("topic with multiple parts", func(t *testing.T) {
		if got, want := hub.GetTopicPermission("a::b::c"), "a"; got != want {
			t.Errorf("GetTopicPermission() = %v, want %v", got, want)
		}
	})
}
