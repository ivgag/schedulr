package service

import (
	"context"
	"time"
	_ "time/tzdata"

	"github.com/cenkalti/backoff/v5"
	"github.com/ivgag/schedulr/model"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// CalendarService handles calendar-related operations.
type GoogleCalendarService struct {
	tokenService *GoogleTokenService
}

// NewGoogleCalendarService creates a new GoogleCalendarService.
func NewGoogleCalendarService(tokenService *GoogleTokenService) *GoogleCalendarService {
	return &GoogleCalendarService{tokenService: tokenService}
}

// CreateEvent creates a new calendar event using the provided token and event data.
func (c *GoogleCalendarService) CreateEvent(userID int, timeZone string, event *model.Event) (model.ScheduledEvent, error) {
	client, err := c.tokenService.ClientForUser(userID)
	if err != nil {
		return model.ScheduledEvent{}, err
	}

	ctx := context.Background()

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return model.ScheduledEvent{}, err
	}

	cal, err := srv.Calendars.Get("primary").Do()
	if err != nil {
		return model.ScheduledEvent{}, err
	}

	if timeZone == "" {
		timeZone = cal.TimeZone
	}

	calEvent, err := toGoogleCalendarEvent(event, timeZone)
	if err != nil {
		return model.ScheduledEvent{}, err
	}

	link, err := c.insertEventWithRetries(srv, calEvent)
	if err != nil {
		log.Error().
			Interface("event", calEvent).
			Err(err).
			Msg("Failed to create event")

		return model.ScheduledEvent{}, err
	}

	return model.ScheduledEvent{
		Event: *event,
		Link:  link,
	}, nil

}

func (c *GoogleCalendarService) insertEventWithRetries(
	srv *calendar.Service,
	event *calendar.Event,
) (string, error) {
	operation := func() (string, error) {
		createdEvent, err := srv.Events.Insert("primary", event).Do()
		if err != nil {
			return "", err
		}

		return createdEvent.HtmlLink, nil
	}

	return backoff.Retry(
		context.Background(),
		operation,
		backoff.WithBackOff(backoff.NewExponentialBackOff()),
		backoff.WithMaxTries(3),
	)
}

func toGoogleCalendarEvent(event *model.Event, timezone string) (*calendar.Event, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		log.Error().
			Str("timezone", timezone).
			Err(err).
			Msg("Failed to load timezone")

		return nil, err
	}

	startTime := toLocalTime(event.Start, loc)
	endTime := toLocalTime(event.End, loc)

	return &calendar.Event{
		Summary:     event.Title,
		Location:    event.Location,
		Description: event.Description,
		Start: &calendar.EventDateTime{
			DateTime: startTime.Format(time.RFC3339),
			TimeZone: timezone,
		},
		End: &calendar.EventDateTime{
			DateTime: endTime.Format(time.RFC3339),
			TimeZone: timezone,
		},
	}, nil
}

func toLocalTime(t time.Time, loc *time.Location) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)
}
