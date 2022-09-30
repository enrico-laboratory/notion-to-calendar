package main

import (
	"context"
	"fmt"
	"github.com/enrico-laboratory/google-api-personal-client/cmd/googleclient"
	"github.com/enrico-laboratory/notion-api-personal-client/cmd/notionclient"
	"github.com/enrico-laboratory/notion-api-personal-client/cmd/notionclient/models/parsedmodels"
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
	chNotion := make(chan *notionclient.NotionApiClient)
	go func() {
		logger.Info().Msg("Initiate Notion Client")
		notionC, err := notionclient.NewClient()
		if err != nil {
			logger.Fatal().Err(err)
		}
		chNotion <- notionC
	}()

	// Get Google Client
	chGoogle := make(chan *googleclient.GClient)
	go func() {
		logger.Info().Msg("Initiate Google Client")
		googleC, err := googleclient.NewClient("credentials.json", context.Background())
		if err != nil {
			logger.Fatal().Err(err)
		}
		chGoogle <- googleC
	}()
	var app application
	app.notionClient = <-chNotion
	app.googleClient = <-chGoogle

	chSchedule := make(chan []parsedmodels.Task)
	// Query Notion Database for tasks scheduled after time.Now()
	go func() {
		logger.Info().Msg("Querying Notion schedule database")
		schedule, err := app.queryScheduleDatabase()
		if err != nil {
			logger.Fatal().Err(err)
		}
		chSchedule <- schedule
	}()

	// Query Notion Database for music project list
	chMusicProjects := make(chan []parsedmodels.MusicProject)
	go func() {
		logger.Info().Msg("Querying Notion music projects database")
		musicProjects, err := app.queryMusicProjectDatabase()
		if err != nil {
			logger.Fatal().Err(err)
		}
		chMusicProjects <- musicProjects
	}()

	chCalendarList := make(chan []googleclient.GCalendarModel)
	go func() {
		logger.Info().Msg("Getting google calendars list")
		calendarList, err := app.googleClient.GCalendar.List()
		if err != nil {
			logger.Fatal().Err(err)
		}
		chCalendarList <- calendarList
	}()
	schedule := <-chSchedule
	musicProjects := <-chMusicProjects
	calendarList := <-chCalendarList

	logger.Info().Msg("Creating Google calendar if does not exist")
	calendarID, err := app.InsertCalendarIfDoesNotExist(calendarName, calendarList, app.googleClient)
	if err != nil {
		logger.Fatal().Err(err)
	}

	logger.Info().Msg(fmt.Sprintf("Getting the events with end date greater than %v", now))
	events, err := app.googleClient.GEvent.ListByTimeMin(calendarID, time.Now())
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
			eventId, err := app.googleClient.GEvent.Insert(calendarID, &event)
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
				err = app.googleClient.GEvent.Delete(calendarID, matchObject.eventId)
				// Build the event to update
				var event googleclient.GEventModel
				err = app.buildEventObject(task, musicProjects, &event)
				eventId, err := app.googleClient.GEvent.Insert(calendarID, &event)
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
			err = app.googleClient.GEvent.Delete(calendarID, event.EventID)
			if err != nil {
				logger.Fatal().Err(err)
			}
			logger.Info().Msg(fmt.Sprintf("Deleted  event in calendar with summary %v and id %v", event.Summary, event.EventID))
			deletedEvent++
		}
	}
	logger.Info().Msg(fmt.Sprintf("%v event(s) inserted, %v event(s) updated, %v event(s) deleted", insertedEvent, updatedEvent, deletedEvent))

}
