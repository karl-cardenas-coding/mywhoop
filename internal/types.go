package internal

import "time"

type User struct {
	UserData         UserData         `json:"user_data"`
	UserMesaurements UserMesaurements `json:"user_mesaurements"`
	SleepCollection  SleepCollection  `json:"sleep_collection"`
}

type UserData struct {
	UserID    int    `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type UserMesaurements struct {
	HeightMeter    float64 `json:"height_meter"`
	WeightKilogram float64 `json:"weight_kilogram"`
	MaxHeartRate   int     `json:"max_heart_rate"`
}

type SleepCollection struct {
	SleepCollectionRecords []SleepCollectionRecords `json:"records"`
	NextToken              string                   `json:"next_token,omitempty"`
}
type StageSummary struct {
	TotalInBedTimeMilli         int `json:"total_in_bed_time_milli"`
	TotalAwakeTimeMilli         int `json:"total_awake_time_milli"`
	TotalNoDataTimeMilli        int `json:"total_no_data_time_milli"`
	TotalLightSleepTimeMilli    int `json:"total_light_sleep_time_milli"`
	TotalSlowWaveSleepTimeMilli int `json:"total_slow_wave_sleep_time_milli"`
	TotalRemSleepTimeMilli      int `json:"total_rem_sleep_time_milli"`
	SleepCycleCount             int `json:"sleep_cycle_count"`
	DisturbanceCount            int `json:"disturbance_count"`
}
type SleepNeeded struct {
	BaselineMilli             int `json:"baseline_milli"`
	NeedFromSleepDebtMilli    int `json:"need_from_sleep_debt_milli"`
	NeedFromRecentStrainMilli int `json:"need_from_recent_strain_milli"`
	NeedFromRecentNapMilli    int `json:"need_from_recent_nap_milli"`
}
type Score struct {
	StageSummary               StageSummary `json:"stage_summary"`
	SleepNeeded                SleepNeeded  `json:"sleep_needed"`
	RespiratoryRate            float64      `json:"respiratory_rate"`
	SleepPerformancePercentage float64      `json:"sleep_performance_percentage"`
	SleepConsistencyPercentage float64      `json:"sleep_consistency_percentage"`
	SleepEfficiencyPercentage  float64      `json:"sleep_efficiency_percentage"`
}
type SleepCollectionRecords struct {
	ID             int       `json:"id"`
	UserID         int       `json:"user_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Start          time.Time `json:"start"`
	End            time.Time `json:"end"`
	TimezoneOffset string    `json:"timezone_offset"`
	Nap            bool      `json:"nap"`
	ScoreState     string    `json:"score_state"`
	Score          Score     `json:"score"`
}

type Cycle struct {
	CycleRecords []CycleRecords `json:"records"`
	NextToken    string         `json:"next_token"`
}
type CycleScore struct {
	Strain           float64 `json:"strain"`
	Kilojoule        float64 `json:"kilojoule"`
	AverageHeartRate int     `json:"average_heart_rate"`
	MaxHeartRate     int     `json:"max_heart_rate"`
}
type CycleRecords struct {
	ID             int         `json:"id"`
	UserID         int         `json:"user_id"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
	Start          time.Time   `json:"start"`
	End            interface{} `json:"end"`
	TimezoneOffset string      `json:"timezone_offset"`
	ScoreState     string      `json:"score_state"`
	Score          CycleScore  `json:"score"`
}
