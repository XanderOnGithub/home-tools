package fitness

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"path/filepath"
	"slices"
	"strings"

	"github.com/XanderOnGithub/home-tools/internal/jsonfile"
)

// This file: per-user fitness profile (fitness.json) and weight log
// (weights.json). Both are optional files in users/<id>/.

// loadBody loads one user's fitness.json and weights.json into s (called
// by Open, after routines are loaded, so schedule references can be checked).
func (s *Store) loadBody(userDir, userID string) error {
	profilePath := filepath.Join(userDir, "fitness.json")
	p, err := jsonfile.Read[Profile](profilePath)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		// Not onboarded yet: fine.
	case err != nil:
		return err
	default:
		if p.UserID != userID {
			return fmt.Errorf("load %s: user_id %q does not match folder %q", profilePath, p.UserID, userID)
		}
		if err := p.Validate(); err != nil {
			return fmt.Errorf("load %s: %w", profilePath, err)
		}
		if err := s.checkProfileRefs(p); err != nil {
			return fmt.Errorf("load %s: %w", profilePath, err)
		}
		s.profiles[userID] = p
	}

	weightsPath := filepath.Join(userDir, "weights.json")
	list, err := jsonfile.Read[[]WeightEntry](weightsPath)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, w := range list {
		if err := w.Validate(); err != nil {
			return fmt.Errorf("load %s: %w", weightsPath, err)
		}
	}
	// Hand edits may reorder entries; sort, then reject duplicate dates
	// (one entry per day is the rule saves keep).
	slices.SortFunc(list, compareDate)
	for i := 1; i < len(list); i++ {
		if list[i].Date == list[i-1].Date {
			return fmt.Errorf("load %s: two entries for %s", weightsPath, list[i].Date)
		}
	}
	s.weights[userID] = list
	return nil
}

// Profile returns userID's fitness profile, and whether they have one
// (false = show onboarding). The schedule map is a copy: callers can't
// modify the store's memory through it.
func (s *Store) Profile(userID string) (Profile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.profiles[userID]
	p.Schedule = maps.Clone(p.Schedule)
	return p, ok
}

// SaveProfile validates p, checks its user and scheduled routines exist,
// and writes it to users/<user_id>/fitness.json, then to memory.
func (s *Store) SaveProfile(p Profile) error {
	if err := p.Validate(); err != nil {
		return err
	}
	p.Schedule = maps.Clone(p.Schedule) // keep our own copy, not the caller's map

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.checkProfileRefs(p); err != nil {
		return err
	}
	if err := jsonfile.Write(filepath.Join(s.dir, "users", p.UserID, "fitness.json"), p); err != nil {
		return err
	}
	s.profiles[p.UserID] = p
	return nil
}

// Weights returns userID's weight log, oldest first, as a new slice.
func (s *Store) Weights(userID string) []WeightEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return slices.Clone(s.weights[userID])
}

// SaveWeight records w for userID, replacing any entry on the same date,
// and rewrites users/<id>/weights.json (the whole log: it's small, about
// one entry a week). Memory changes only after the file is written.
func (s *Store) SaveWeight(userID string, w WeightEntry) error {
	if err := w.Validate(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.users.User(userID); !ok {
		return fmt.Errorf("%w weight: unknown user %q", ErrInvalid, userID)
	}

	// Binary search by date: O(log n) to find the spot, O(n) to insert.
	next := slices.Clone(s.weights[userID])
	i, found := slices.BinarySearchFunc(next, w, compareDate)
	if found {
		next[i] = w
	} else {
		next = slices.Insert(next, i, w)
	}

	if err := jsonfile.Write(filepath.Join(s.dir, "users", userID, "weights.json"), next); err != nil {
		return err
	}
	s.weights[userID] = next
	return nil
}

// checkProfileRefs checks p's user exists and every scheduled routine is
// in the store. Same locking rule as checkSessionRefs.
func (s *Store) checkProfileRefs(p Profile) error {
	if _, ok := s.users.User(p.UserID); !ok {
		return fmt.Errorf("%w fitness profile: unknown user %q", ErrInvalid, p.UserID)
	}
	for day, routineID := range p.Schedule {
		if _, ok := s.routines[routineID]; !ok {
			return fmt.Errorf("%w fitness profile %s: %s has unknown routine %q", ErrInvalid, p.UserID, day, routineID)
		}
	}
	return nil
}

// compareDate orders weight entries by date. "YYYY-MM-DD" strings sort
// the same as the dates they name, so no parsing is needed.
func compareDate(a, b WeightEntry) int { return strings.Compare(a.Date, b.Date) }
