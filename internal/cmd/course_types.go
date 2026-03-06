package cmd

import (
	"fmt"
	"sort"
	"strings"
)

var courseActivityTypes = map[string]int{
	"running":         1,
	"cycling":         2,
	"hiking":          3,
	"other":           4,
	"mountain_biking": 5,
	"trail_running":   6,
	"street_running":  7,
	"walking":         9,
	"road_biking":     10,
	"gravel_cycling":  143,
	"e_bike_mountain": 175,
	"e_bike_fitness":  176,
	"ultra_run":       181,
}

func resolveCourseActivityType(typeKey string) (int, error) {
	if id, ok := courseActivityTypes[typeKey]; ok {
		return id, nil
	}

	keys := make([]string, 0, len(courseActivityTypes))
	for k := range courseActivityTypes {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return 0, fmt.Errorf("unknown activity type %q; valid types: %s", typeKey, strings.Join(keys, ", "))
}
