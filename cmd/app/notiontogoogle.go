package main

import (
	"errors"
	"fmt"
	"github.com/enrico-laboratory/google-api-personal-client/cmd/googleclient"
	"github.com/enrico-laboratory/notion-api-personal-client/cmd/notionclient/models/parsedmodels"
	"time"
)

func (app *application) buildEventObject(task parsedmodels.Task, musicProjects []parsedmodels.MusicProject, event *googleclient.GEventModel) error {

	choir := app.getChoir(musicProjects, &task)
	app.setDescription(event, &task)
	app.setSummary(choir, event, &task)
	err := app.setEventDates(event, &task)
	if err != nil {
		return err
	}
	//		log.Println(fmt.Sprintf(
	//			`
	//+++++++++++++++
	//Title:
	//%v
	//Description:
	//%v
	//Start Date Time: %v
	//End Date Time: %v
	//		`, event.Summary, event.Description, event.StartDateTime, event.EndDateTime))
	return nil
}

func (app *application) getChoir(musicProjects []parsedmodels.MusicProject, task *parsedmodels.Task) string {
	for _, project := range musicProjects {
		if len(task.MusicProject) > 0 {
			if project.Id == task.MusicProject[0] {
				return project.ChoirRollup
			}
		}
	}
	return ""
}

func (app *application) setDescription(event *googleclient.GEventModel, task *parsedmodels.Task) {
	if task.Type != "" {
		event.Description = fmt.Sprintf(
			"%v\n\nTaskID\n%v\nLastEdited\n%v", task.Title, task.Id, task.LastEditedTime.Format(time.RFC3339))
	} else if task.Priority == "Scheduled" {
		event.Description = fmt.Sprintf(
			"TaskID\n%v\nLastEdited\n%v", task.Id, task.LastEditedTime.Format(time.RFC3339))
	} else {
		event.Description = "NO SCHEDULED, CHECK!!"
	}
}

func (app *application) setSummary(choir string, event *googleclient.GEventModel, task *parsedmodels.Task) {
	if task.Type != "" {
		event.Summary = fmt.Sprintf(
			`%v: %v`, choir, task.Type)
	} else if task.Priority == "Scheduled" {
		event.Summary = task.Title
	} else {
		event.Summary = "NO SCHEDULED, CHECK!!"
	}
}

func (app *application) setEventDates(event *googleclient.GEventModel, task *parsedmodels.Task) error {
	var eventEndDate googleclient.GEventDateTime
	var eventStartDate googleclient.GEventDateTime
	taskEndDate := task.EndDateAndTime
	taskStartDate := task.StartDateAndTime
	err := app.parseAndSetDate(&eventStartDate, taskStartDate)
	if err != nil {
		return err
	}

	if len(taskEndDate) == 0 {
		if len(taskStartDate) == 10 {
			eventEndDate.Date = eventStartDate.Date
		} else if len(taskStartDate) > 10 {
			eventEndDate.DateTime = eventStartDate.DateTime.Add(30 * time.Minute)
		}
	} else {
		err = app.parseAndSetDate(&eventEndDate, taskEndDate)
		if err != nil {
			return err
		}
	}

	event.StartDateTime = eventStartDate
	event.EndDateTime = eventEndDate
	return nil
}

func (app *application) parseAndSetDate(eventDate *googleclient.GEventDateTime, taskDate string) error {
	if len(taskDate) == 10 {
		parsedDate, err := time.Parse("2006-01-02", taskDate)
		if err != nil {
			return err
		}
		eventDate.Date = parsedDate
	}
	if len(taskDate) < 10 {
		return errors.New("invalid date length")
	}
	if len(taskDate) > 10 {
		parsedDate, err := time.Parse(time.RFC3339, taskDate)
		if err != nil {
			return err
		}
		eventDate.DateTime = parsedDate
	}
	return nil
}
