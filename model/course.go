package model

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Course int

const (
	CourseUnknown   = Course(0)
	CourseSapporo   = Course(1)
	CourseHakodate  = Course(2)
	CourseFukushima = Course(3)
	CourseNiigata   = Course(4)
	CourseTokyo     = Course(5)
	CourseNakayama  = Course(6)
	CourseChukyo    = Course(7)
	CourseKyoto     = Course(8)
	CourseHanshin   = Course(9)
	CourseKokura    = Course(10)
)

var courseNames = map[Course]struct {
	JP string
	EN string
}{
	CourseSapporo:   {"札幌", "Sapporo"},
	CourseHakodate:  {"函館", "Hakodate"},
	CourseFukushima: {"福島", "Fukushima"},
	CourseNiigata:   {"新潟", "Niigata"},
	CourseTokyo:     {"東京", "Tokyo"},
	CourseNakayama:  {"中山", "Nakayama"},
	CourseChukyo:    {"中京", "Chukyo"},
	CourseKyoto:     {"京都", "Kyoto"},
	CourseHanshin:   {"阪神", "Hanshin"},
	CourseKokura:    {"小倉", "Kokura"},
}

var courseJPToCourse = (func() map[string]Course {
	data := make(map[string]Course, len(courseNames))
	for c, v := range courseNames {
		data[v.JP] = c
	}
	return data
}())

// courseENToCourse is a reverse lookup map from lowercased English course name to Course ID.
// It is built at package initialization using an immediately-invoked function literal (IIFE)
// to avoid polluting the package namespace with a temporary variable.
var courseENToCourse = (func() map[string]Course {
	data := make(map[string]Course, len(courseNames))
	for c, v := range courseNames {
		data[strings.ToLower(v.EN)] = c
	}
	return data
}())

func getCourseByName(name string) (Course, error) {
	c, ok := courseJPToCourse[name]
	if !ok {
		return CourseUnknown, fmt.Errorf("invalid course name: %s", name)
	}
	return c, nil
}

func CourseByENName(name string) (Course, error) {
	c, ok := courseENToCourse[strings.ToLower(name)]
	if !ok {
		return CourseUnknown, fmt.Errorf("invalid course name: %s", name)
	}
	return c, nil
}

func (c Course) String() string {
	if c == CourseUnknown {
		return "unknown"
	}

	name, ok := courseNames[c]
	if !ok {
		return "invalid"
	}

	return strings.ToLower(name.EN)
}

func (c Course) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}
