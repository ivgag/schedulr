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

// Provider returns the provider of the calendar service.
func (c *GoogleCalendarService) Provider() model.Provider {
	return model.ProviderGoogle
}

// CreateEvent creates a new calendar event using the provided token and event data.
func (c *GoogleCalendarService) CreateEvent(userID int, event *model.Event) (model.ScheduledEvent, error) {
	client, err := c.tokenService.ClientForUser(userID)
	if err != nil {
		return model.ScheduledEvent{}, err
	}

	ctx := context.Background()

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return model.ScheduledEvent{}, err
	}

	calEvent, err := toGoogleCalendarEvent(event)
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

func toGoogleCalendarEvent(event *model.Event) (*calendar.Event, error) {
	startTime, err := toLocalTime(event.Start.Timestamp, event.Start.TimeZone)
	endTime, err := toLocalTime(event.End.Timestamp, event.End.TimeZone)
	if err != nil {
		return nil, err
	}

	return &calendar.Event{
		Summary:     event.Title,
		Location:    event.Location,
		Description: event.Description,
		Start: &calendar.EventDateTime{
			DateTime: startTime.Format(time.RFC3339),
			TimeZone: event.Start.TimeZone,
		},
		End: &calendar.EventDateTime{
			DateTime: endTime.Format(time.RFC3339),
			TimeZone: event.End.TimeZone,
		},
	}, nil
}

func toLocalTime(
	t time.Time,
	timezone string,
) (time.Time, error) {

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		log.Error().
			Str("timezone", timezone).
			Err(err).
			Msg("Failed to load timezone")

		return t, err
	}

	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc), nil
}
