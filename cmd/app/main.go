package main

import (
	"context"
	"fmt"
	"github.com/enrico-laboratory/google-api-personal-client/cmd/googleclient"
	"github.com/enrico-laboratory/notion-api-personal-client/cmd/notionclient"
	"github.com/rs/zerolog"
	"os"
	"time"
)

var logger zerolog.Logger
var logFile *os.File

type application struct {
	notionClient *notionclient.NotionApiClient
	googleClient *googleclient.GClient
}

type match struct {
	eventId             string
	eventTaskId         string
	eventTaskLastEdited string
}

func main() {

	logger = zerolog.New(os.Stdout).With().Timestamp().Stack().Caller().Logger()

	for {
		err := notionToCalendar()
		if err != nil {
			logger.Error().Err(err).Send()
		}

		waitingTime := 1 * time.Hour
		logger.Info().Msg(fmt.Sprintf("Waiting for %v seconds...", waitingTime))
		time.Sleep(waitingTime)
	}
}

func notionToCalendar() error {
	calendarName := os.Getenv("GOOGLE_CALENDAR_ID")
	now := time.Now()

	//beginningOfCurrentYear := time.Date(currentYear, time.Month(1), 1, 0, 0, 0, 0, &time.Location{})
	//endOfCurrentYear := time.Date(currentYear, time.Month(12), 31, 23, 59, 59, 0, &time.Location{})

	logger.Info().Msg("==== NEW SESSION ====")

	logger.Info().Msg("Initiate Notion Client")
	notionC, err := notionclient.NewClient()
	if err != nil {
		return err
	}

	logger.Info().Msg("Initiate Google Client")
	googleC, err := googleclient.NewClient(context.Background(), os.Getenv("GOOGLE_KEY"))
	if err != nil {
		return err
	}

	var app application

	app.notionClient = notionC
	app.googleClient = googleC

	logger.Info().Msg("Querying Notion schedule for rehearsals, concerts and meeting type tasks...")
	schedule, err := app.queryNotionScheduleForFutureTasks(now.Format(time.RFC3339))
	if err != nil {
		return err
	}
	logger.Info().Msg(fmt.Sprintf("%v tasks retrieved.", len(schedule)))

	logger.Info().Msg("Querying Notion music projects")
	musicProjects, err := app.queryMusicProjects()
	if err != nil {
		return err
	}

	logger.Info().Msg(fmt.Sprintf("Getting future events..."))
	events, err := googleC.GEvent.ListByTimeMin(calendarName, now)
	if err != nil {
		return err
	}
	logger.Info().Msg(fmt.Sprintf("%v events retrieved.", len(events)))

	logger.Info().Msg("Deleting all events retrieved...")
	eventsDeleted := 0
	for _, event := range events {
		if isANotionScheduleEvent(event.Description) {
			err = googleC.GEvent.Delete(calendarName, event.EventID)
			if err != nil {
				return err
			}
			eventsDeleted++
		}
	}
	logger.Info().Msg(fmt.Sprintf("%v events deleted", eventsDeleted))

	logger.Info().Msg(fmt.Sprintf("Inserting %v task into GCalendar...", len(schedule)))
	eventsCreated := 0
	for _, task := range schedule {
		var event googleclient.GEventModel
		err = app.buildEventObject(task, musicProjects, &event)
		_, err := googleC.GEvent.Insert(calendarName, &event)
		if err != nil {
			return err
		}
		eventsCreated++
	}
	logger.Info().Msg(fmt.Sprintf("%v events inserted.", eventsCreated))

	return nil
}
