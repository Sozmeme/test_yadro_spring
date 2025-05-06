package utils

import (
	"strings"
	"testing"
)

func TestCompetitionScenario(t *testing.T) {
	config := `{
		"laps": 2,
		"lapLen": 3651,
		"penaltyLen": 50,
		"firingLines": 1,
		"start": "09:30:00.000",
		"startDelta": "00:00:30"
	}`

	events := `[09:05:59.867] 1 1
[09:15:00.841] 2 1 09:30:00.000
[09:29:45.734] 3 1
[09:30:01.005] 4 1
[09:49:31.659] 5 1 1
[09:49:33.123] 6 1 1
[09:49:34.650] 6 1 2
[09:49:35.937] 6 1 4
[09:49:37.364] 6 1 5
[09:49:38.339] 7 1
[09:49:55.915] 8 1
[09:51:48.391] 9 1
[09:59:03.872] 10 1
[09:59:03.872] 11 1 Lost in the forest`

	expectedLog := `[09:05:59.867] The competitor(1) registered
[09:15:00.841] The start time for the competitor(1) was set by a draw to 09:30:00.000
[09:29:45.734] The competitor(1) is on the start line
[09:30:01.005] The competitor(1) has started
[09:49:31.659] The competitor(1) is on the firing range(1)
[09:49:33.123] The target(1) has been hit by competitor(1)
[09:49:34.650] The target(2) has been hit by competitor(1)
[09:49:35.937] The target(4) has been hit by competitor(1)
[09:49:37.364] The target(5) has been hit by competitor(1)
[09:49:38.339] The competitor(1) left the firing range
[09:49:55.915] The competitor(1) entered the penalty laps
[09:51:48.391] The competitor(1) left the penalty laps
[09:59:03.872] The competitor(1) ended the main lap
[09:59:03.872] The competitor(1) can't continue: Lost in the forest`

	expectedReport := "[NotFinished] 1 [{00:29:03.872, 2.094}, {,}] {00:01:52.476, 0.445} 4/5" //  wrong data in readme - fixed here

	p, err := NewProcessor(config, events)
	if err != nil {
		t.Fatalf("NewProcessor failed: %v", err)
	}
	logOutput := p.ProcessEvents()
	reportOutput := p.GenerateSummary()

	if logOutput != expectedLog {
		t.Errorf("Log output doesn't match expected:\nGot:\n%s\n\nExpected:\n%s", logOutput, expectedLog)
	}

	if reportOutput != expectedReport {
		t.Errorf("Report output doesn't match expected:\nGot: %s\nExpected: %s", reportOutput, expectedReport)
	}

	competitor := p.competitors[1]
	if competitor == nil {
		t.Fatal("Competitor not found")
	}

	if !competitor.CannotContinue {
		t.Errorf("Expected status NotFinished")
	}

	if competitor.Hits != 4 {
		t.Errorf("Expected 4 hits, got %d", competitor.Hits)
	}
}

func TestRegistrationAndStart(t *testing.T) {
	config := `{
		"laps": 2,
		"lapLen": 3651,
		"penaltyLen": 50,
		"firingLines": 1,
		"start": "09:30:00.000",
		"startDelta": "00:00:30"
	}`

	events := `[09:05:59.867] 1 1
[09:15:00.841] 2 1 09:30:00.000
[09:29:45.734] 3 1
[09:30:01.005] 4 1`

	p, err := NewProcessor(config, events)
	if err != nil {
		t.Fatalf("NewProcessor failed: %v", err)
	}
	log := p.ProcessEvents()

	if len(p.competitors) != 1 {
		t.Errorf("Expected 1 competitor, got %d", len(p.competitors))
	}

	expectedLog := `[09:05:59.867] The competitor(1) registered
[09:15:00.841] The start time for the competitor(1) was set by a draw to 09:30:00.000
[09:29:45.734] The competitor(1) is on the start line
[09:30:01.005] The competitor(1) has started`
	if log != expectedLog {
		t.Errorf("Log output doesn't match expected:\nGot:\n%s\n\nExpected:\n%s", log, expectedLog)
	}

	c := p.competitors[1]
	if c.AssignedStart == nil || c.ActualStart == nil {
		t.Error("Start times not set correctly")
	}
}

func TestDisqualification(t *testing.T) {
	config := `{
		"laps": 2,
		"lapLen": 3651,
		"penaltyLen": 50,
		"firingLines": 1,
		"start": "09:30:00.000",
		"startDelta": "00:00:30"
	}`

	events := `[09:05:59.867] 1 1
[09:15:00.841] 2 1 09:30:00.000
[09:29:45.734] 3 1
[09:30:31.005] 4 1`

	p, err := NewProcessor(config, events)
	if err != nil {
		t.Fatalf("NewProcessor failed: %v", err)
	}
	p.ProcessEvents()

	c := p.competitors[1]
	if !c.Disqualified {
		t.Error("Competitor should be disqualified")
	}
}

func TestCannotContinue(t *testing.T) {
	config := `{
		"laps": 2,
		"lapLen": 3651,
		"penaltyLen": 50,
		"firingLines": 1,
		"start": "09:30:00.000",
		"startDelta": "00:00:30"
	}`

	events := `[09:05:59.867] 1 1
[09:15:00.841] 2 1 09:30:00.000
[09:29:45.734] 3 1
[09:30:01.005] 4 1
[09:59:03.872] 11 1 Lost in the forest`

	p, err := NewProcessor(config, events)
	if err != nil {
		t.Fatalf("NewProcessor failed: %v", err)
	}
	p.ProcessEvents()
	report := p.GenerateSummary()

	c := p.competitors[1]
	if !c.CannotContinue {
		t.Error("Competitor should have 'CannotContinue' status")
	}

	expectedReport := "[NotFinished] 1 [{,}, {,}] {00:00:00.000, 0.000} 0/5"
	if report != expectedReport {
		t.Errorf("Report doesn't match expected:\nGot: %s\nExpected: %s", report, expectedReport)
	}
}

func TestEventCount(t *testing.T) {
	config := `{
		"laps": 2,
		"lapLen": 3651,
		"penaltyLen": 50,
		"firingLines": 1,
		"start": "09:30:00",
		"startDelta": "00:00:30"
	}`

	events := `[09:05:59.867] 1 1
[09:15:00.841] 2 1 09:30:00.000
[09:29:45.734] 3 1
[09:30:01.005] 4 1
[09:49:31.659] 5 1 1
[09:49:33.123] 6 1 1
[09:49:34.650] 6 1 2
[09:49:35.937] 6 1 4
[09:49:37.364] 6 1 5
[09:49:38.339] 7 1
[09:49:55.915] 8 1
[09:51:48.391] 9 1
[09:59:03.872] 10 1
[09:59:03.872] 11 1 Lost in the forest`

	p, err := NewProcessor(config, events)
	if err != nil {
		t.Fatalf("NewProcessor failed: %v", err)
	}

	inputLines := strings.Split(events, "\n")
	if len(p.eventQueue) != len(inputLines) {
		t.Errorf("Event count mismatch. Expected %d events (one per line), got %d",
			len(inputLines), len(p.eventQueue))
	}
}
