package utils

import (
	"time"
)

type Config struct {
	Laps        int     `json:"laps"`
	LapLen      float64 `json:"lapLen"`
	PenaltyLen  float64 `json:"penaltyLen"`
	FiringLines int     `json:"firingLines"`
	Start       string  `json:"start"`
	StartDelta  string  `json:"startDelta"`
}

type Event struct {
	Time         time.Time
	ID           int
	CompetitorID int
	ExtraParams  []string
}

type Competitor struct {
	ID                int
	Registered        bool
	AssignedStart     *time.Time
	ActualStart       *time.Time
	ActualFinish      *time.Time
	Finished          bool
	Disqualified      bool
	CannotContinue    bool
	CannotContinueMsg string

	LapEndTimes       []time.Time
	PenaltyStartTimes []time.Time
	PenaltyEndTimes   []time.Time
	TotalTime   time.Duration
	Hits              int
}

type Processor struct {
	config      Config
	competitors map[int]*Competitor
	eventQueue  []Event
	currentTime time.Time
}
