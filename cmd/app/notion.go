package main

import (
	"fmt"
	"github.com/enrico-laboratory/notion-api-personal-client/cmd/notionclient/models/parsedmodels"
	"time"
)

func (app *application) queryScheduleDatabase() ([]parsedmodels.Task, error) {

	query := fmt.Sprintf(`{ 
				"filter": {
		              "property": "Do Date",
		              "date": {
		                  "on_or_after": "%v"
		              }
				}
			}`, time.Now().Format(time.RFC3339))

	tasks, err := app.notionClient.Schedule.Query(query)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (app *application) queryMusicProjectDatabase() ([]parsedmodels.MusicProject, error) {

	tasks, err := app.notionClient.MusicProjects.Query("")
	if err != nil {
		return nil, err
	}
	return tasks, nil
}
