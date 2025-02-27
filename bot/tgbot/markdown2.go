package tgbot

import (
	"fmt"
	"strings"
	"time"

	"github.com/ivgag/schedulr/model"
)

var replacer = strings.NewReplacer(
	"_", "\\_",
	"*", "\\*",
	"[", "\\[",
	"]", "\\]",
	"(", "\\(",
	")", "\\)",
	"~", "\\~",
	"`", "\\`",
	">", "\\>",
	"#", "\\#",
	"+", "\\+",
	"-", "\\-",
	".", "\\.",
	",", "\\,",
	"!", "\\!",
)

func FormatScheduledEvent(scheduledEvent *model.ScheduledEvent) string {
	event := scheduledEvent.Event

	// Helper to escape reserved characters for MarkdownV2.
	escape := func(text string) string {

		return replacer.Replace(text)
	}

	message := fmt.Sprintf("*%s*\n", escape(event.Title))

	if event.Description != "" {
		message += fmt.Sprintf("%s\n", escape(event.Description))
	}

	message += fmt.Sprintf("*When:* %s \\(%s\\) \\- %s \\(%s\\)\n",
		escape(event.Start.Timestamp.Format(time.DateTime)),
		escape(event.Start.TimeZone),
		escape(event.End.Timestamp.Format(time.DateTime)),
		escape(event.End.TimeZone),
	)

	if event.Location != "" {
		message += fmt.Sprintf("*Where:* %s\n", escape(event.Location))
	}

	if scheduledEvent.Link != "" {
		message += fmt.Sprintf("[%s](%s)\n", escape("More details"), scheduledEvent.Link)
	} else {
		message += fmt.Sprintf("[%s](%s)\n", escape("Add event to Calendar"), event.DeepLink)
	}
	return message
}

func escape(str string) string {
	return replacer.Replace(str)
}
