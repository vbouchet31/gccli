package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bpauli/gccli/internal/outfmt"
)

// CoursesCmd groups course subcommands.
type CoursesCmd struct {
	List      CoursesListCmd      `cmd:"" default:"withargs" help:"List courses."`
	Favorites CoursesFavoritesCmd `cmd:"" help:"List favorite courses."`
	Detail    CourseDetailCmd     `cmd:"" help:"View course details."`
	Send      CourseSendCmd       `cmd:"" help:"Send a course to a device."`
}

// CoursesListCmd lists the user's courses.
type CoursesListCmd struct{}

func (c *CoursesListCmd) Run(g *Globals) error {
	client, err := resolveClient(g)
	if err != nil {
		return err
	}

	data, err := client.GetCourses(g.Context)
	if err != nil {
		return fmt.Errorf("list courses: %w", err)
	}

	if outfmt.IsJSON(g.Context) {
		return outfmt.WriteJSON(os.Stdout, json.RawMessage(data))
	}

	courses, err := parseCourses(data)
	if err != nil {
		return err
	}

	rows := formatCourseRows(courses)
	header := []string{"ID", "NAME", "TYPE", "DISTANCE", "CREATED"}

	if outfmt.IsPlain(g.Context) {
		return outfmt.WritePlain(os.Stdout, rows)
	}
	return outfmt.WriteTable(os.Stdout, header, rows)
}

// CoursesFavoritesCmd lists favorite courses.
type CoursesFavoritesCmd struct{}

func (c *CoursesFavoritesCmd) Run(g *Globals) error {
	client, err := resolveClient(g)
	if err != nil {
		return err
	}

	data, err := client.GetCourseFavorites(g.Context)
	if err != nil {
		return fmt.Errorf("list favorite courses: %w", err)
	}

	if outfmt.IsJSON(g.Context) {
		return outfmt.WriteJSON(os.Stdout, json.RawMessage(data))
	}

	var courses []map[string]any
	if err := json.Unmarshal(data, &courses); err != nil {
		return fmt.Errorf("parse favorite courses: %w", err)
	}

	rows := formatCourseRows(courses)
	header := []string{"ID", "NAME", "TYPE", "DISTANCE", "CREATED"}

	if outfmt.IsPlain(g.Context) {
		return outfmt.WritePlain(os.Stdout, rows)
	}
	return outfmt.WriteTable(os.Stdout, header, rows)
}

// CourseDetailCmd shows details for a course.
type CourseDetailCmd struct {
	ID string `arg:"" help:"Course ID."`
}

func (c *CourseDetailCmd) Run(g *Globals) error {
	client, err := resolveClient(g)
	if err != nil {
		return err
	}

	data, err := client.GetCourse(g.Context, c.ID)
	if err != nil {
		return fmt.Errorf("get course: %w", err)
	}

	return outfmt.WriteJSON(os.Stdout, json.RawMessage(data))
}

// CourseSendCmd sends a course to a device.
type CourseSendCmd struct {
	CourseID string `arg:"" help:"Course ID."`
	DeviceID string `arg:"" help:"Device ID."`
}

func (c *CourseSendCmd) Run(g *Globals) error {
	client, err := resolveClient(g)
	if err != nil {
		return err
	}

	// Fetch course detail to get the course name.
	courseData, err := client.GetCourse(g.Context, c.CourseID)
	if err != nil {
		return fmt.Errorf("get course: %w", err)
	}

	var course map[string]any
	if err := json.Unmarshal(courseData, &course); err != nil {
		return fmt.Errorf("parse course: %w", err)
	}
	courseName := jsonString(course, "courseName")

	data, err := client.SendCourseToDevice(g.Context, c.CourseID, c.DeviceID, courseName)
	if err != nil {
		return fmt.Errorf("send course to device: %w", err)
	}

	if outfmt.IsJSON(g.Context) {
		return outfmt.WriteJSON(os.Stdout, json.RawMessage(data))
	}

	// Extract response details for success message.
	var msgs []map[string]any
	if err := json.Unmarshal(data, &msgs); err == nil && len(msgs) > 0 {
		deviceName := jsonString(msgs[0], "deviceName")
		messageID := jsonString(msgs[0], "messageId")
		g.UI.Successf("Sent course %q to device %s (message %s)", courseName, deviceName, messageID)
	} else {
		g.UI.Successf("Sent course %q to device %s", courseName, c.DeviceID)
	}

	return nil
}

// parseCourses extracts the course list from the API response.
// The list endpoint returns {coursesForUser: [...]}.
func parseCourses(data json.RawMessage) ([]map[string]any, error) {
	var wrapper map[string]any
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("parse courses: %w", err)
	}

	arr, ok := wrapper["coursesForUser"]
	if !ok {
		return nil, nil
	}

	items, ok := arr.([]any)
	if !ok {
		return nil, nil
	}

	courses := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if m, ok := item.(map[string]any); ok {
			courses = append(courses, m)
		}
	}
	return courses, nil
}

// formatCourseRows extracts table rows from course data.
func formatCourseRows(courses []map[string]any) [][]string {
	rows := make([][]string, 0, len(courses))
	for _, c := range courses {
		rows = append(rows, []string{
			jsonString(c, "courseId"),
			jsonString(c, "courseName"),
			courseTypeKey(c),
			formatDistance(jsonFloat(c, "distanceInMeters")),
			formatDate(jsonString(c, "createdDate")),
		})
	}
	return rows
}

// courseTypeKey extracts the course type from nested courseType.typeKey.
func courseTypeKey(c map[string]any) string {
	ct, ok := c["courseType"]
	if !ok || ct == nil {
		return ""
	}
	if m, ok := ct.(map[string]any); ok {
		return jsonString(m, "typeKey")
	}
	return ""
}
