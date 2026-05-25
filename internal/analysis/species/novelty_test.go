package species

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tphakala/birdnet-go/internal/conf"
)

func testNoveltyTracker(windowDays int) *SpeciesTracker {
	return NewTrackerFromSettings(nil, &conf.SpeciesTrackingSettings{
		Enabled:              true,
		NewSpeciesWindowDays: windowDays,
		YearlyTracking: conf.YearlyTrackingSettings{
			Enabled: false,
		},
		SeasonalTracking: conf.SeasonalTrackingSettings{
			Enabled: false,
		},
	})
}

func TestCheckAndUpdateSpeciesWithNovelty_FirstEverEpisode(t *testing.T) {
	tracker := testNoveltyTracker(7)
	detectionTime := time.Date(2026, 5, 23, 8, 0, 0, 0, time.UTC)

	isNew, daysSinceFirst, novelty := tracker.CheckAndUpdateSpeciesWithNovelty("Setophaga castanea", detectionTime)

	assert.True(t, isNew)
	assert.Equal(t, 0, daysSinceFirst)
	assert.True(t, novelty.NoveltyEpisodeActive)
	assert.Equal(t, inactiveNoveltyValue, novelty.DaysSinceLastSeen)
	assert.Equal(t, firstEverNoveltyEpisodeDays, novelty.NoveltyEpisodeDays)
	assert.Equal(t, "first_ever", novelty.NoveltyReason)
	assert.Equal(t, 0, novelty.NoveltyDaysActive)
	assert.Equal(t, detectionTime, novelty.NoveltyEpisodeStart)
}

func TestCheckAndUpdateSpeciesWithNovelty_ReturnAfterAbsenceEpisode(t *testing.T) {
	tracker := testNoveltyTracker(7)
	firstTime := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)
	returnTime := firstTime.AddDate(0, 0, 12)

	_, _, _ = tracker.CheckAndUpdateSpeciesWithNovelty("Setophaga castanea", firstTime)
	_, daysSinceFirst, novelty := tracker.CheckAndUpdateSpeciesWithNovelty("Setophaga castanea", returnTime)

	assert.Equal(t, 12, daysSinceFirst)
	assert.True(t, novelty.NoveltyEpisodeActive)
	assert.Equal(t, 12, novelty.DaysSinceLastSeen)
	assert.Equal(t, 12, novelty.NoveltyEpisodeDays)
	assert.Equal(t, "return_after_absence", novelty.NoveltyReason)
	assert.Equal(t, 0, novelty.NoveltyDaysActive)
	assert.Equal(t, returnTime, novelty.NoveltyEpisodeStart)
}

func TestCheckAndUpdateSpeciesWithNovelty_EpisodePersistsForWindow(t *testing.T) {
	tracker := testNoveltyTracker(7)
	firstTime := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)
	returnTime := firstTime.AddDate(0, 0, 12)
	nextDay := returnTime.AddDate(0, 0, 1)

	_, _, _ = tracker.CheckAndUpdateSpeciesWithNovelty("Setophaga castanea", firstTime)
	_, _, _ = tracker.CheckAndUpdateSpeciesWithNovelty("Setophaga castanea", returnTime)
	_, _, novelty := tracker.CheckAndUpdateSpeciesWithNovelty("Setophaga castanea", nextDay)

	assert.True(t, novelty.NoveltyEpisodeActive)
	assert.Equal(t, 1, novelty.DaysSinceLastSeen)
	assert.Equal(t, 12, novelty.NoveltyEpisodeDays)
	assert.Equal(t, 1, novelty.NoveltyDaysActive)
	assert.Equal(t, returnTime, novelty.NoveltyEpisodeStart)
}

func TestCheckAndUpdateSpeciesWithNovelty_NoEpisodeForSameDayDetection(t *testing.T) {
	tracker := testNoveltyTracker(7)
	detectionTime := time.Date(2026, 5, 23, 8, 0, 0, 0, time.UTC)
	const scientificName = "Setophaga castanea"

	tracker.speciesFirstSeen[scientificName] = detectionTime.AddDate(0, 0, -30)
	tracker.speciesLastSeen[scientificName] = detectionTime

	_, _, novelty := tracker.CheckAndUpdateSpeciesWithNovelty(scientificName, detectionTime.Add(2*time.Hour))

	assert.False(t, novelty.NoveltyEpisodeActive)
	assert.Equal(t, 0, novelty.DaysSinceLastSeen)
	assert.Equal(t, inactiveNoveltyValue, novelty.NoveltyEpisodeDays)
	assert.Equal(t, inactiveNoveltyValue, novelty.NoveltyDaysActive)
}
