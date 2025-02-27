package tgbot_test

import (
	"testing"

	"github.com/go-telegram/bot/models"
	"github.com/ivgag/schedulr/tgbot"
)

func TestFormatTextWithEntities(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		entities []models.MessageEntity
		want     string
	}{
		{
			name: "No entities",
			text: "Hello world",
			// nil slice also works
			entities: nil,
			want:     "Hello world",
		},
		{
			name: "Single entity covering whole text",
			text: "Hello world",
			entities: []models.MessageEntity{
				{Offset: 0, Length: 11, URL: "http://example.com"},
			},
			want: "Hello world[http://example.com]",
		},
		{
			name: "Single entity in the middle",
			text: "Say hello to my friend",
			entities: []models.MessageEntity{
				{Offset: 4, Length: 5, URL: "http://hi.com"},
			},
			// "Say hello" is formed by taking text[0:9] then appending the URL,
			// and the remainder " to my friend" is appended.
			want: "Say hello[http://hi.com] to my friend",
		},
		{
			name: "Multiple entities",
			text: "Check out google and facebook",
			entities: []models.MessageEntity{
				{Offset: 21, Length: 8, URL: "https://facebook.com"},
				{Offset: 10, Length: 6, URL: "https://google.com"},
			},
			// After sorting, first entity: text[0:16] = "Check out google"
			// becomes "Check out google[https://google.com]",
			// then from index 16 to (21+8)=29: " and facebook" becomes " and facebook[https://facebook.com]"
			want: "Check out google[https://google.com] and facebook[https://facebook.com]",
		},
		{
			name: "Multibyte characters",
			text: "Привет, мир!",
			entities: []models.MessageEntity{
				{Offset: 8, Length: 3, URL: "http://russia.com"},
			},
			// rune conversion ensures correct handling of multibyte characters.
			// substring from index 0 to 8+3=11 yields "Привет, мир" then appended "[http://russia.com]" and finally "!".
			want: "Привет, мир[http://russia.com]!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tgbot.FormatTextWithEntities(tt.text, tt.entities)
			if got != tt.want {
				t.Errorf("formatMessageText() = %q, want %q", got, tt.want)
			}
		})
	}
}
