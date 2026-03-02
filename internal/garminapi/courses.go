package garminapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// GetCourses returns all courses owned by the current user.
func (c *Client) GetCourses(ctx context.Context) (json.RawMessage, error) {
	return c.ConnectAPI(ctx, http.MethodGet, "/web-gateway/course/owner/", nil)
}

// GetCourse returns details for a specific course.
func (c *Client) GetCourse(ctx context.Context, courseID string) (json.RawMessage, error) {
	return c.ConnectAPI(ctx, http.MethodGet, "/course-service/course/"+courseID, nil)
}

// GetCourseFavorites returns the user's favorite courses.
func (c *Client) GetCourseFavorites(ctx context.Context) (json.RawMessage, error) {
	return c.ConnectAPI(ctx, http.MethodGet, "/course-service/course/favorites", nil)
}

// SendCourseToDevice sends a course to a device via the device message API.
func (c *Client) SendCourseToDevice(ctx context.Context, courseID, deviceID, courseName string) (json.RawMessage, error) {
	deviceIDInt, err := strconv.ParseInt(deviceID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse device ID: %w", err)
	}

	courseIDInt, err := strconv.ParseInt(courseID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse course ID: %w", err)
	}

	payload := []map[string]any{
		{
			"deviceId":    deviceIDInt,
			"messageUrl":  fmt.Sprintf("course-service/course/fit/%s/%s?elevation=true", courseID, deviceID),
			"messageType": "courses",
			"messageName": courseName,
			"groupName":   nil,
			"priority":    0,
			"fileType":    "FIT",
			"metaDataId":  courseIDInt,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal send payload: %w", err)
	}

	return c.ConnectAPI(ctx, http.MethodPost, "/device-service/devicemessage/messages", bytes.NewReader(body))
}
