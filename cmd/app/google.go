package main

import (
	"errors"
	"fmt"
	"github.com/enrico-laboratory/google-api-personal-client/cmd/googleclient"
	"github.com/enrico-laboratory/notion-api-personal-client/cmd/notionclient/models/parsedmodels"
)

func (app *application) isCalendar(calendarSummary string, calendarList []googleclient.GCalendarModel) bool {

	for _, calendar := range calendarList {
		if calendar.Summary == calendarSummary {
			return true
		}
	}
	return false
}

func (app *application) getCalendarById(calendarSummary string, calendarList []googleclient.GCalendarModel) (string, error) {
	for _, calendar := range calendarList {
		if calendar.Summary == calendarSummary {
			return calendar.Id, nil
		}
	}
	return "", errors.New(fmt.Sprintf("no calendar found with summary: %v", calendarSummary))
}

func (app *application) InsertCalendarIfDoesNotExist(calendarName string, calendarList []googleclient.GCalendarModel, googleC *googleclient.GClient) (string, error) {
	var calendarID string
	var err error
	if !app.isCalendar(calendarName, calendarList) {
		calendarID, err = googleC.GCalendar.Insert(calendarName)
		if err != nil {
			return "", err
		}
	} else {
		calendarID, err = app.getCalendarById(calendarName, calendarList)
		if err != nil {
			return "", err
		}
	}
	return calendarID, nil
}

func (app *application) getTaskIdFromEvent(taskId string, eventList []googleclient.GEventModel) (match, error) {
	for _, event := range eventList {
		eventTaskId, err := getNotionTaskIdFromDescription(event.Description)
		if err != nil {
			return match{}, err
		}
		eventTaskLastEdited, err := getNotionLastEditedFromDescription(event.Description)
		if err != nil {
			return match{}, err
		}
		if taskId == eventTaskId {
			return match{
				eventId:             event.EventID,
				eventTaskId:         eventTaskId,
				eventTaskLastEdited: eventTaskLastEdited,
			}, nil
		}
	}

	return match{}, nil
}

func (app *application) doesEventExistInNotion(event googleclient.GEventModel, schedule []parsedmodels.Task) bool {
	eventTaskId, _ := getNotionTaskIdFromDescription(event.Description)

	for _, task := range schedule {
		if task.Id == eventTaskId {
			return true
		}

	}
	return false
}
