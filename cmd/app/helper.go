package main

import (
	"errors"
	"strings"
)

func isANotionScheduleEvent(eventDescription string) bool {
	if strings.Contains(eventDescription, "TaskID") {
		return true
	}
	return false
}

func splitMultiLineString(multilineString string) []string {
	stringList := strings.Split(multilineString, "\n")
	var stringListTrimmed []string
	for _, str := range stringList {
		stringTrimmed := strings.TrimSpace(str)
		stringListTrimmed = append(stringListTrimmed, stringTrimmed)
	}
	return stringListTrimmed
}

func getNotionTaskIdFromDescription(multilineString string) (string, error) {
	stringList := splitMultiLineString(multilineString)

	if len(stringList) < 4 {
		return "", errors.New("description has the wrong format, not enough lines, probably do not contain id or lastEdited")
	}
	return getLines(len(stringList)-3, stringList), nil
}

func getNotionLastEditedFromDescription(multilineString string) (string, error) {
	stringList := splitMultiLineString(multilineString)
	if len(stringList) < 4 {
		return "", errors.New("description has the wrong format, not enough lines, probably do not contain id or lastEdited")
	}
	return getLines(len(stringList)-1, stringList), nil
}

func getLines(index int, stringList []string) string {
	return stringList[index]
}
