package main

import (
	"fmt"
	"github.com/enrico-laboratory/notion-api-personal-client/cmd/notionclient/models/parsedmodels"
)

func (app *application) queryNotionScheduleForFutureTasks(now string) ([]parsedmodels.Task, error) {

	query := fmt.Sprintf(`{
			"filter": {
	 "and": [
	   {
	     "property": "Do Date",
	     "date": {
	       "on_or_after": "%v"
	     }
	   },
       {
		"or": [
				{
					"property": "Type",
					"select": {
					  "equals": "Rehearsal"
    				}
				},
				{
					"property": "Type",
					"select": {
					  "equals": "Concert"
    				}
				},
				{
					"property": "Type",
					"select": {
					  "equals": "Meeting"
    				}
				}
			  ]
	}
	 ]
	}
			}`, now)

	return app.queryNotionSchedule(query)

}

func (app *application) queryNotionScheduleForTasksInRange(onOrAfter, onOrBefore string) ([]parsedmodels.Task, error) {

	query := fmt.Sprintf(`{
			"filter": {
	"and": [
	  {
	    "property": "Do Date",
	    "date": {
	      "on_or_after": "%v"
	    }
	  },
	  {
	    "property": "Do Date",
	    "date": {
	      "on_or_before": "%v"
	    }
	  }, 
	  {
		"or": [
				{
					"property": "Type",
					"select": {
					  "equals": "Rehearsal"
					}
				},
				{
					"property": "Type",
					"select": {
					  "equals": "Concert"
					}
				},
				{
					"property": "Type",
					"select": {
					  "equals": "Meeting"
					}
				}
			  ]
	}
	]
	}
			}`, onOrAfter, onOrBefore)

	return app.queryNotionSchedule(query)
}
func (app *application) queryNotionSchedule(query string) ([]parsedmodels.Task, error) {

	tasks, err := app.notionClient.Schedule.Query(query)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (app *application) queryMusicProjects() ([]parsedmodels.MusicProject, error) {

	tasks, err := app.notionClient.MusicProjects.Query("")
	if err != nil {
		return nil, err
	}
	return tasks, nil
}
