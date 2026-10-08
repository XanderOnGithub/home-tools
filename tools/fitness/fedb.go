package fitness

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

// SourceFreeExerciseDB marks exercises imported from
// https://github.com/yuhonas/free-exercise-db (public domain, Unlicense).
const SourceFreeExerciseDB = "free-exercise-db"

// fedbExercise is one record of free-exercise-db's dist/exercises.json.
// JSON nulls decode as "". Fields we don't use (force, mechanic) are ignored.
type fedbExercise struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Level            string   `json:"level"`
	Equipment        string   `json:"equipment"`
	Category         string   `json:"category"`
	PrimaryMuscles   []string `json:"primaryMuscles"`
	SecondaryMuscles []string `json:"secondaryMuscles"`
	Instructions     []string `json:"instructions"`
	Images           []string `json:"images"`
}

// fedbEquipment maps the dataset's equipment to ours. "" (none) means no
// equipment; ok=false means an unknown value.
var fedbEquipment = map[string]Equipment{
	"barbell":       EquipmentBarbell,
	"dumbbell":      EquipmentDumbbell,
	"kettlebells":   EquipmentKettlebell,
	"cable":         EquipmentCable,
	"machine":       EquipmentMachine,
	"bands":         EquipmentBand,
	"e-z curl bar":  EquipmentEZBar,
	"medicine ball": EquipmentMedicineBall,
	"exercise ball": EquipmentExerciseBall,
	"foam roll":     EquipmentFoamRoller,
	"body only":     "",
	"other":         "",
	"":              "",
}

// fedbBodyweight is equipment that adds no measurable load, so logging
// weight is optional (0 = bodyweight only).
var fedbBodyweight = []string{"body only", "other", "", "bands", "exercise ball", "foam roll"}

// ParseFreeExerciseDB converts free-exercise-db's dist/exercises.json into
// validated exercises. Mapping:
//   - primary muscles → activation 1.0, secondary → 0.5
//   - stretching and cardio track duration; everything else reps + weight
//   - bodyweight-style equipment makes weight optional
//
// Any value outside the dataset's known vocabulary is an error, so a
// dataset change fails loudly instead of importing bad data.
func ParseFreeExerciseDB(data []byte) ([]Exercise, error) {
	var raw []fedbExercise
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("decode free-exercise-db: %w", err)
	}

	out := make([]Exercise, 0, len(raw))
	for _, r := range raw {
		ex, err := r.toExercise()
		if err != nil {
			return nil, err
		}
		out = append(out, ex)
	}
	return out, nil
}

func (r fedbExercise) toExercise() (Exercise, error) {
	ex := Exercise{
		ID:           r.ID,
		Name:         r.Name,
		Activation:   make(map[Muscle]float64, len(r.PrimaryMuscles)+len(r.SecondaryMuscles)),
		Category:     Category(snake(r.Category)),
		Level:        Level(r.Level),
		Instructions: r.Instructions,
		Images:       r.Images,
		Bodyweight:   slices.Contains(fedbBodyweight, r.Equipment),
		Source:       SourceFreeExerciseDB,
	}

	eq, ok := fedbEquipment[r.Equipment]
	if !ok {
		return Exercise{}, fmt.Errorf("%w free-exercise-db %s: unknown equipment %q", ErrInvalid, r.ID, r.Equipment)
	}
	if eq != "" {
		ex.Equipment = []Equipment{eq}
	}

	switch ex.Category {
	case CategoryStretching, CategoryCardio:
		ex.Metrics = []Metric{MetricDuration}
	default:
		ex.Metrics = []Metric{MetricReps, MetricWeight}
	}

	// Secondary first, then primary, so a muscle listed in both ends at 1.0.
	for _, m := range r.SecondaryMuscles {
		ex.Activation[Muscle(snake(m))] = 0.5
	}
	for _, m := range r.PrimaryMuscles {
		ex.Activation[Muscle(snake(m))] = 1
	}

	if err := ex.Validate(); err != nil {
		return Exercise{}, fmt.Errorf("free-exercise-db: %w", err)
	}
	return ex, nil
}

// snake converts the dataset's "lower back" style to our "lower_back".
func snake(s string) string {
	return strings.ReplaceAll(s, " ", "_")
}
