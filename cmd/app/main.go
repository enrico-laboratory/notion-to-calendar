package main

import (
	"context"
	"fmt"
	"github.com/enrico-laboratory/google-api-personal-client/cmd/googleclient"
	"github.com/enrico-laboratory/notion-api-personal-client/cmd/notionclient"
	"github.com/go-co-op/gocron"
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

var logFilePath string

func main() {
	now := time.Now().Format("2006-01-02")
	logFilePath = fmt.Sprintf("%v.%v", os.Getenv("LOGS_FILE_PATH"), now)

	var err error
	logFile, err = os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	logger = zerolog.New(logFile).With().Timestamp().Caller().Logger()
	err = notionToCalendar()
	if err != nil {
		logger.Error().Err(err).Send()
	}

	time.Sleep(1000 * time.Second)
}

func runCronJobs() error {
	s := gocron.NewScheduler(time.UTC)

	job, err := s.Every(15).Minutes().Do(func() {
		err := notionToCalendar()
		if err != nil {
			logger.Error().Err(err).Send()
		}
	})
	if err != nil {
		return err
	}
	lastRun := job.LastRun().Format("06-01-02T15:04")
	logger.Info().Msg(fmt.Sprintf("LAST RUN: %v", lastRun))
	s.StartBlocking()
	return nil
}

func notionToCalendar() error {
	var err error
	calendarName := os.Getenv("NOTION_CALENDAR_NAME")
	now := time.Now().Format("2006-01-02")
	logFile, err = os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	var updatedEvent int
	var insertedEvent int
	var deletedEvent int

	// Starting Session
	logger.Info().Str("SESSION", fmt.Sprintf("%v", time.Now().Format("06-01-02T15:04"))).Send()

	// Get Notion Client
	logger.Info().Msg("Initiate Notion Client")
	notionC, err := notionclient.NewClient()
	if err != nil {
		return err
	}
	// Get Google Client
	logger.Info().Msg("Initiate Google Client")
	googleC, err := googleclient.NewClient(context.Background(), os.Getenv("GOOGLE_KEY"))
	if err != nil {
		return err
	}

	var app application

	app.notionClient = notionC
	app.googleClient = googleC

	logger.Info().Msg("Querying Notion schedule database")
	schedule, err := app.queryScheduleDatabase()
	if err != nil {
		return err
	}

	logger.Info().Msg("Querying Notion music projects database")
	musicProjects, err := app.queryMusicProjectDatabase()
	if err != nil {
		return err
	}

	logger.Info().Msg(fmt.Sprintf("Getting the events with end date greater than %v", now))
	events, err := googleC.GEvent.ListByTimeMin(calendarName, time.Now())
	if err != nil {
		return err
	}

	logger.Info().Msg("Scanning Notion tasks for new or newer tasks")
	for _, task := range schedule {
		matchObject, err := app.getTaskIdFromEvent(task.Id, events)
		if err != nil {
			return err
		}
		//log.Println(fmt.Sprintf("\nEventID: %v\nTaskID: %v\nEventTaskID: %v\nEventLastEdited: %v", matchObject.eventId, task.Id, matchObject.eventTaskId, matchObject.eventTaskLastEdited))
		if matchObject.eventTaskId == "" {
			var event googleclient.GEventModel
			err = app.buildEventObject(task, musicProjects, &event)
			eventId, err := googleC.GEvent.Insert(calendarName, &event)
			if err != nil {
				return err
			}
			logger.Info().Msg(fmt.Sprintf("Inserted new event in calendar with summary %v and id %v", event.Summary, eventId))
			insertedEvent++
		} else {
			eventTaskLastEditedParsed, err := time.Parse(time.RFC3339, matchObject.eventTaskLastEdited)
			if err != nil {
				return err
			}
			if eventTaskLastEditedParsed.Equal(task.LastEditedTime) {
				// do nothing
			} else {
				logger.Info().Msg("Updating newer event in calendar")
				// delete and insert event
				err = googleC.GEvent.Delete(calendarName, matchObject.eventId)
				// Build the event to update
				var event googleclient.GEventModel
				err = app.buildEventObject(task, musicProjects, &event)
				eventId, err := googleC.GEvent.Insert(calendarName, &event)
				if err != nil {
					return err
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
			err = googleC.GEvent.Delete(calendarName, event.EventID)
			if err != nil {
				return err
			}
			logger.Info().Msg(fmt.Sprintf("Deleted  event in calendar with summary %v and id %v", event.Summary, event.EventID))
			deletedEvent++
		}
	}
	logger.Info().Msg(fmt.Sprintf("%v event(s) inserted, %v event(s) updated, %v event(s) deleted", insertedEvent, updatedEvent, deletedEvent))
	return nil
}
