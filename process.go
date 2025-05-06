package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

const layout = "15:04:05.000"

func createProcessor(config string, events string) *Processor {
	var cfg Config
	err := json.Unmarshal([]byte(config), &cfg)
	if err != nil {
		panic("invalid configuration: " + err.Error())
	}

	processor := &Processor{
		config:      cfg,
		competitors: make(map[int]*Competitor),
	}

	lines := strings.Split(events, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		timeEnd := strings.Index(line, "]")
		if timeEnd == -1 {
			continue
		}
		timeStr := line[1:timeEnd]
		eventTime, err := time.Parse(layout, timeStr)
		if err != nil {
			continue
		}

		remaining := strings.TrimSpace(line[timeEnd+1:])
		parts := strings.Fields(remaining)
		if len(parts) < 2 {
			continue
		}

		id, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		competitorID, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}

		event := Event{
			Time:         eventTime,
			ID:           id,
			CompetitorID: competitorID,
		}

		if len(parts) > 2 {
			event.ExtraParams = parts[2:]
		}

		processor.eventQueue = append(processor.eventQueue, event)

		if _, exists := processor.competitors[event.CompetitorID]; !exists {
			processor.competitors[event.CompetitorID] = &Competitor{
				ID:                event.CompetitorID,
				LapEndTimes:       []time.Time{},
				PenaltyStartTimes: []time.Time{},
				PenaltyEndTimes:   []time.Time{},
				Hits:              0,
			}
		}
	}

	return processor
}

func (p *Processor) ProcessEvents() string {
	var outputLines []string

	for _, event := range p.eventQueue {
		p.currentTime = event.Time
		competitor := p.competitors[event.CompetitorID]
		if competitor == nil {
			continue
		}

		var line string
		switch event.ID {
		case 1:
			line = p.handleRegistration(event, competitor)
		case 2:
			line = p.handleAssignedStart(event, competitor)
		case 3:
			line = p.handleOnStartLine(event, competitor)
		case 4:
			line = p.handleStarted(event, competitor)
		case 5:
			line = p.handleOnFiringRange(event, competitor)
		case 6:
			line = p.handleTargetHit(event, competitor)
		case 7:
			line = p.handleLeftFiringRange(event, competitor)
		case 8:
			line = p.handleEnteredPenaltyLaps(event, competitor)
		case 9:
			line = p.handleLeftPenaltyLaps(event, competitor)
		case 10:
			line = p.handleEndedMainLap(event, competitor)
		case 11:
			line = p.handleCannotContinue(event, competitor)
		}

		if line != "" {
			outputLines = append(outputLines, line)
		}
	}

	return strings.Join(outputLines, "\n")
}

func (p *Processor) GenerateSummary() string {
	var competitors []*Competitor

	for _, c := range p.competitors {
		if c.Registered {
			if c.Finished && c.ActualFinish != nil && c.AssignedStart != nil {
				c.TotalTime = c.ActualFinish.Sub(*c.AssignedStart)
			}
			competitors = append(competitors, c)
		}
	}

	sort.Slice(competitors, func(i, j int) bool {
		c1, c2 := competitors[i], competitors[j]

		if c1.Finished && c2.Finished {
			return c1.TotalTime < c2.TotalTime
		}
		if c1.Finished {
			return true
		}
		if c2.Finished {
			return false
		}
		if c1.CannotContinue && c2.Disqualified {
			return true
		}
		return c1.ID < c2.ID
	})

	var results []string
	for _, c := range competitors {
		results = append(results, formatCompetitorReport(c, p.config))
	}

	return strings.Join(results, "\n")
}

func (p *Processor) handleRegistration(event Event, competitor *Competitor) string {
	competitor.Registered = true
	return formatOutput(event.Time, fmt.Sprintf("The competitor(%d) registered", event.CompetitorID))
}

func (p *Processor) handleAssignedStart(event Event, competitor *Competitor) string {
	startTime, err := time.Parse(layout, event.ExtraParams[0])
	if err != nil {
		return ""
	}
	competitor.AssignedStart = &startTime
	return formatOutput(event.Time, fmt.Sprintf("The start time for the competitor(%d) was set by a draw to %s",
		event.CompetitorID, event.ExtraParams[0]))
}

func (p *Processor) handleOnStartLine(event Event, competitor *Competitor) string {
	return formatOutput(event.Time, fmt.Sprintf("The competitor(%d) is on the start line", event.CompetitorID))
}

func (p *Processor) handleStarted(event Event, competitor *Competitor) string {
	competitor.ActualStart = &event.Time

	if competitor.AssignedStart != nil {
		delta, err := parseDuration(p.config.StartDelta)
		if err == nil {
			maxStartTime := competitor.AssignedStart.Add(delta)
			if event.Time.After(maxStartTime) {
				competitor.Disqualified = true
			}
		}
	}

	return formatOutput(event.Time, fmt.Sprintf("The competitor(%d) has started", event.CompetitorID))
}

func (p *Processor) handleOnFiringRange(event Event, competitor *Competitor) string {
	return formatOutput(event.Time, fmt.Sprintf("The competitor(%d) is on the firing range(%s)",
		event.CompetitorID, event.ExtraParams[0]))
}

func (p *Processor) handleTargetHit(event Event, competitor *Competitor) string {
	competitor.Hits++
	return formatOutput(event.Time, fmt.Sprintf("The target(%s) has been hit by competitor(%d)",
		event.ExtraParams[0], event.CompetitorID))
}

func (p *Processor) handleLeftFiringRange(event Event, competitor *Competitor) string {
	return formatOutput(event.Time, fmt.Sprintf("The competitor(%d) left the firing range", event.CompetitorID))
}

func (p *Processor) handleEnteredPenaltyLaps(event Event, competitor *Competitor) string {
	competitor.PenaltyStartTimes = append(competitor.PenaltyStartTimes, event.Time)
	return formatOutput(event.Time, fmt.Sprintf("The competitor(%d) entered the penalty laps", event.CompetitorID))
}

func (p *Processor) handleLeftPenaltyLaps(event Event, competitor *Competitor) string {
	competitor.PenaltyEndTimes = append(competitor.PenaltyEndTimes, event.Time)
	return formatOutput(event.Time, fmt.Sprintf("The competitor(%d) left the penalty laps", event.CompetitorID))
}

func (p *Processor) handleEndedMainLap(event Event, competitor *Competitor) string {
	competitor.LapEndTimes = append(competitor.LapEndTimes, event.Time)

	if len(competitor.LapEndTimes) == p.config.Laps && !competitor.CannotContinue {
		competitor.ActualFinish = &event.Time
		competitor.Finished = true
	}

	return formatOutput(event.Time, fmt.Sprintf("The competitor(%d) ended the main lap", event.CompetitorID))
}

func (p *Processor) handleCannotContinue(event Event, competitor *Competitor) string {
	competitor.CannotContinue = true
	competitor.CannotContinueMsg = strings.Join(event.ExtraParams, " ")
	return formatOutput(event.Time, fmt.Sprintf("The competitor(%d) can't continue: %s",
		event.CompetitorID, competitor.CannotContinueMsg))
}

func formatOutput(t time.Time, msg string) string {
	return fmt.Sprintf("[%s] %s", t.Format(layout), msg)
}

func parseDuration(input string) (time.Duration, error) {
	parts := strings.Split(input, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid time format")
	}

	h, _ := strconv.Atoi(parts[0])
	m, _ := strconv.Atoi(parts[1])
	s, _ := strconv.Atoi(parts[2])

	return time.Duration(h)*time.Hour +
		time.Duration(m)*time.Minute +
		time.Duration(s)*time.Second, nil
}

func formatCompetitorReport(c *Competitor, cfg Config) string {
	statusPrefix := ""
	if c.Finished {
		statusPrefix = fmt.Sprintf("[%s]", formatDuration(c.TotalTime))
	} else if c.Disqualified {
		statusPrefix = "[NotStarted]"
	} else if c.CannotContinue {
		statusPrefix = "[NotFinished]"
	} else {
		statusPrefix = "[Unknown]"
	}

	var lapInfo []string
	for i := 0; i < cfg.Laps; i++ {
		if i < len(c.LapEndTimes) {
			var start time.Time
			if i == 0 {
				if c.ActualStart != nil {
					start = *c.AssignedStart
				} else {
					lapInfo = append(lapInfo, "{,}")
					continue
				}
			} else {
				start = c.LapEndTimes[i-1]
			}
			dur := c.LapEndTimes[i].Sub(start)
			speed := 0.0
			if dur > 0 {
				speed = cfg.LapLen / dur.Seconds()
			}
			lapInfo = append(lapInfo, fmt.Sprintf("{%s, %.3f}", formatDuration(dur), speed))
		} else {
			lapInfo = append(lapInfo, "{,}")
		}
	}
	laps := "[" + strings.Join(lapInfo, ", ") + "]"

	penaltyDur := time.Duration(0)
	for i := 0; i < len(c.PenaltyStartTimes) && i < len(c.PenaltyEndTimes); i++ {
		penaltyDur += c.PenaltyEndTimes[i].Sub(c.PenaltyStartTimes[i])
	}
	penaltySpeed := 0.0
	if penaltyDur > 0 {
		penaltySpeed = cfg.PenaltyLen / penaltyDur.Seconds()
	}
	penalty := fmt.Sprintf("{%s, %.3f}", formatDuration(penaltyDur), penaltySpeed)

	totalTargets := cfg.FiringLines * 5
	shooting := fmt.Sprintf("%d/%d", c.Hits, totalTargets)

	return fmt.Sprintf("%s %d %s %s %s",
		statusPrefix,
		c.ID,
		laps,
		penalty,
		shooting,
	)
}
func formatDuration(d time.Duration) string {
	d = d.Round(time.Millisecond)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	ms := (d - s*time.Second) / time.Millisecond
	return fmt.Sprintf("%02d:%02d:%02d.%03d", h, m, s, ms)
}
