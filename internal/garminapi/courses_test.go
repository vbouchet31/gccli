package garminapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
)

// --- GetCourses tests ---

func TestGetCourses_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/web-gateway/course/owner/" {
			t.Errorf("path = %s, want /web-gateway/course/owner/", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		_, _ = w.Write([]byte(`{"coursesForUser":[{"courseId":123,"courseName":"Morning Run"}]}`))
	})

	_, client := testServer(t, handler)
	data, err := client.GetCourses(context.Background())
	if err != nil {
		t.Fatalf("GetCourses: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if result["coursesForUser"] == nil {
		t.Error("expected coursesForUser field")
	}
}

func TestGetCourses_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("error"))
	})

	_, client := testServer(t, handler)
	_, err := client.GetCourses(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}

	var apiErr *GarminAPIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected GarminAPIError, got %T: %v", err, err)
	}
}

// --- GetCourse tests ---

func TestGetCourse_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/course-service/course/12345" {
			t.Errorf("path = %s, want /course-service/course/12345", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		_, _ = w.Write([]byte(`{"courseId":12345,"courseName":"Morning Run","distanceInMeters":5000}`))
	})

	_, client := testServer(t, handler)
	data, err := client.GetCourse(context.Background(), "12345")
	if err != nil {
		t.Fatalf("GetCourse: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if result["courseName"] != "Morning Run" {
		t.Errorf("courseName = %v, want Morning Run", result["courseName"])
	}
}

func TestGetCourse_NotFound(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	})

	_, client := testServer(t, handler)
	_, err := client.GetCourse(context.Background(), "99999")
	if err == nil {
		t.Fatal("expected error")
	}

	var apiErr *GarminAPIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected GarminAPIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want %d", apiErr.StatusCode, http.StatusNotFound)
	}
}

// --- GetCourseFavorites tests ---

func TestGetCourseFavorites_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/course-service/course/favorites" {
			t.Errorf("path = %s, want /course-service/course/favorites", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		_, _ = w.Write([]byte(`[{"courseId":123,"courseName":"Fav Run"}]`))
	})

	_, client := testServer(t, handler)
	data, err := client.GetCourseFavorites(context.Background())
	if err != nil {
		t.Fatalf("GetCourseFavorites: %v", err)
	}

	var result []map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 course, got %d", len(result))
	}
	if result[0]["courseName"] != "Fav Run" {
		t.Errorf("courseName = %v, want Fav Run", result[0]["courseName"])
	}
}

func TestGetCourseFavorites_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("error"))
	})

	_, client := testServer(t, handler)
	_, err := client.GetCourseFavorites(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}

	var apiErr *GarminAPIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected GarminAPIError, got %T: %v", err, err)
	}
}

// --- SendCourseToDevice tests ---

func TestSendCourseToDevice_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/device-service/devicemessage/messages" {
			t.Errorf("path = %s, want /device-service/devicemessage/messages", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		var payload []map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("unmarshal request body: %v", err)
		}
		if len(payload) != 1 {
			t.Fatalf("expected 1 message, got %d", len(payload))
		}
		msg := payload[0]
		if msg["messageType"] != "courses" {
			t.Errorf("messageType = %v, want courses", msg["messageType"])
		}
		if msg["messageName"] != "Morning Run" {
			t.Errorf("messageName = %v, want Morning Run", msg["messageName"])
		}
		if msg["fileType"] != "FIT" {
			t.Errorf("fileType = %v, want FIT", msg["fileType"])
		}
		if msg["messageUrl"] != "course-service/course/fit/12345/67890?elevation=true" {
			t.Errorf("messageUrl = %v, want course-service/course/fit/12345/67890?elevation=true", msg["messageUrl"])
		}
		// deviceId should be numeric
		if msg["deviceId"] != float64(67890) {
			t.Errorf("deviceId = %v, want 67890", msg["deviceId"])
		}
		// metaDataId should be numeric
		if msg["metaDataId"] != float64(12345) {
			t.Errorf("metaDataId = %v, want 12345", msg["metaDataId"])
		}

		_, _ = w.Write([]byte(`[{"messageId":999,"messageStatus":"new","deviceName":"Forerunner 265"}]`))
	})

	_, client := testServer(t, handler)
	data, err := client.SendCourseToDevice(context.Background(), "12345", "67890", "Morning Run")
	if err != nil {
		t.Fatalf("SendCourseToDevice: %v", err)
	}

	var result []map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0]["messageStatus"] != "new" {
		t.Errorf("messageStatus = %v, want new", result[0]["messageStatus"])
	}
}

func TestSendCourseToDevice_InvalidDeviceID(t *testing.T) {
	_, client := testServer(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	_, err := client.SendCourseToDevice(context.Background(), "12345", "not-a-number", "Test")
	if err == nil {
		t.Fatal("expected error for invalid device ID")
	}
}

func TestSendCourseToDevice_InvalidCourseID(t *testing.T) {
	_, client := testServer(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	_, err := client.SendCourseToDevice(context.Background(), "not-a-number", "67890", "Test")
	if err == nil {
		t.Fatal("expected error for invalid course ID")
	}
}

func TestSendCourseToDevice_ServerError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("error"))
	})

	_, client := testServer(t, handler)
	_, err := client.SendCourseToDevice(context.Background(), "12345", "67890", "Test")
	if err == nil {
		t.Fatal("expected error")
	}

	var apiErr *GarminAPIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected GarminAPIError, got %T: %v", err, err)
	}
}
