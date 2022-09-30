package main

import (
	"context"
	"fmt"
	"github.com/enrico-laboratory/google-api-personal-client/cmd/googleclient"
	"github.com/enrico-laboratory/notion-api-personal-client/cmd/notionclient"
	"github.com/rs/zerolog"
	"log"
	"os"
	"time"
)

const (
	calendarName = "Notion"
)

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
	now := time.Now().Format("2006-01-02")
	var updatedEvent int
	var insertedEvent int
	var deletedEvent int

	f, err := os.OpenFile(fmt.Sprintf("logfile.%v", now), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	_, err = f.WriteString(fmt.Sprintf("===============================\n+ Log Session: %v +\n===============================\n", time.Now().Format("06-04-02T15:04")))
	if err != nil {
		log.Fatalf("error writing into file: %v", err)
	}

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	// Get Notion Client
	logger.Info().Msg("Initiate Notion Client")
	notionC, err := notionclient.NewClient()
	if err != nil {
		logger.Fatal().Err(err)
	}
	// Get Google Client
	logger.Info().Msg("Initiate Google Client")
	googleC, err := googleclient.NewClient("credentials.json", context.Background())
	if err != nil {
		logger.Fatal().Err(err)
	}

	var app application

	app.notionClient = notionC
	app.googleClient = googleC

	// Query Notion Database for tasks scheduled after time.Now()
	logger.Info().Msg("Querying Notion schedule database")
	schedule, err := app.queryScheduleDatabase()
	if err != nil {
		logger.Fatal().Err(err)
	}

	// Query Notion Database for music project list
	logger.Info().Msg("Querying Notion music projects database")
	musicProjects, err := app.queryMusicProjectDatabase()
	if err != nil {
		logger.Fatal().Err(err)
	}

	logger.Info().Msg("Getting google calendars list")
	calendarList, err := googleC.GCalendar.List()
	if err != nil {
		logger.Fatal().Err(err)
	}

	logger.Info().Msg("Creating Google calendar if does not exist")
	calendarID, err := app.InsertCalendarIfDoesNotExist(calendarName, calendarList, googleC)
	if err != nil {
		logger.Fatal().Err(err)
	}

	logger.Info().Msg(fmt.Sprintf("Getting the events with end date greater than %v", now))
	events, err := googleC.GEvent.ListByTimeMin(calendarID, time.Now())
	if err != nil {
		logger.Fatal().Err(err)
	}

	logger.Info().Msg("Scanning Notion tasks for new or newer tasks")
	for _, task := range schedule {
		matchObject, err := app.getTaskIdFromEvent(task.Id, events)
		if err != nil {
			logger.Fatal().Err(err)
		}
		if matchObject.eventTaskId == "" {
			var event googleclient.GEventModel
			err = app.buildEventObject(task, musicProjects, &event)
			eventId, err := googleC.GEvent.Insert(calendarID, &event)
			if err != nil {
				logger.Fatal().Err(err)
			}
			logger.Info().Msg(fmt.Sprintf("Inserted new event in calendar with summary %v and id %v", event.Summary, eventId))
			insertedEvent++
		} else {
			eventTaskLastEditedParsed, err := time.Parse(time.RFC3339, matchObject.eventTaskLastEdited)
			if err != nil {
				logger.Fatal().Err(err)
			}
			if eventTaskLastEditedParsed.Equal(task.LastEditedTime) {
				// do nothing
			} else {
				logger.Info().Msg("Updating newer event in calendar")
				// delete and insert event
				err = googleC.GEvent.Delete(calendarID, matchObject.eventId)
				// Build the event to update
				var event googleclient.GEventModel
				err = app.buildEventObject(task, musicProjects, &event)
				eventId, err := googleC.GEvent.Insert(calendarID, &event)
				if err != nil {
					logger.Fatal().Err(err)
				}
				logger.Info().Msg(fmt.Sprintf("Updated  event in calendar with summary %v and id %v", event.Summary, eventId))
				updatedEvent++
			}
		}
	}
	logger.Info().Msg("Deleting event not present in Notion anymore")
	// delete events that exist in calendar but not in Notion
	for _, event := range events {
		if !app.doesEventExistInNotion(event, schedule) {
			err = googleC.GEvent.Delete(calendarID, event.EventID)
			if err != nil {
				logger.Fatal().Err(err)
			}
			logger.Info().Msg(fmt.Sprintf("Deleted  event in calendar with summary %v and id %v", event.Summary, event.EventID))
			deletedEvent++
		}
	}
	logger.Info().Msg(fmt.Sprintf("%v event(s) inserted, %v event(s) updated, %v event(s) deleted", insertedEvent, updatedEvent, deletedEvent))

}
