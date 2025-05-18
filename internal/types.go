// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"net/http"
	"strings"
	"time"

	"github.com/karl-cardenas-coding/mywhoop/export"
	"github.com/karl-cardenas-coding/mywhoop/notifications"
)

/*
* Whoop API Data Structures
 */

type AuthCredentials struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
}

type User struct {
	UserData           UserData           `json:"user_data,omitempty" csv:"user_data"`
	UserMesaurements   UserMesaurements   `json:"user_mesaurements,omitempty" csv:"user_mesaurements"`
	SleepCollection    SleepCollection    `json:"sleep_collection" csv:"sleep_collection"`
	RecoveryCollection RecoveryCollection `json:"recovery_collection" csv:"recovery_collection"`
	WorkoutCollection  WorkoutCollection  `json:"workout_collection" csv:"workout_collection"`
	CycleCollection    CycleCollection    `json:"cycle_collection" csv:"cycle_collection"`
}

type UserData struct {
	UserID    int    `json:"user_id" csv:"user_id"`
	Email     string `json:"email" csv:"email"`
	FirstName string `json:"first_name" csv:"first_name"`
	LastName  string `json:"last_name" csv:"last_name"`
}

type UserMesaurements struct {
	HeightMeter    float64 `json:"height_meter" csv:"height_meter"`
	WeightKilogram float64 `json:"weight_kilogram" csv:"weight_kilogram"`
	MaxHeartRate   int     `json:"max_heart_rate" csv:"max_heart_rate"`
}

type SleepCollection struct {
	SleepCollectionRecords []SleepCollectionRecords `json:"records" csv:"records"`
	NextToken              *string                  `json:"next_token,omitempty" csv:"next_token,omitempty"`
}
type StageSummary struct {
	TotalInBedTimeMilli         int `json:"total_in_bed_time_milli" csv:"total_in_bed_time_milli"`
	TotalAwakeTimeMilli         int `json:"total_awake_time_milli" csv:"total_awake_time_milli"`
	TotalNoDataTimeMilli        int `json:"total_no_data_time_milli" csv:"total_no_data_time_milli"`
	TotalLightSleepTimeMilli    int `json:"total_light_sleep_time_milli" csv:"total_light_sleep_time_milli"`
	TotalSlowWaveSleepTimeMilli int `json:"total_slow_wave_sleep_time_milli" csv:"total_slow_wave_sleep_time_milli"`
	TotalRemSleepTimeMilli      int `json:"total_rem_sleep_time_milli" csv:"total_rem_sleep_time_milli"`
	SleepCycleCount             int `json:"sleep_cycle_count" csv:"sleep_cycle_count"`
	DisturbanceCount            int `json:"disturbance_count" csv:"disturbance_count"`
}
type SleepNeeded struct {
	BaselineMilli             int `json:"baseline_milli" csv:"baseline_milli"`
	NeedFromSleepDebtMilli    int `json:"need_from_sleep_debt_milli" csv:"need_from_sleep_debt_milli"`
	NeedFromRecentStrainMilli int `json:"need_from_recent_strain_milli" csv:"need_from_recent_strain_milli"`
	NeedFromRecentNapMilli    int `json:"need_from_recent_nap_milli" csv:"need_from_recent_nap_milli"`
}
type Score struct {
	StageSummary               StageSummary `json:"stage_summary" csv:"stage_summary"`
	SleepNeeded                SleepNeeded  `json:"sleep_needed" csv:"sleep_needed"`
	RespiratoryRate            float64      `json:"respiratory_rate" csv:"respiratory_rate"`
	SleepPerformancePercentage float64      `json:"sleep_performance_percentage" csv:"sleep_performance_percentage"`
	SleepConsistencyPercentage float64      `json:"sleep_consistency_percentage" csv:"sleep_consistency_percentage"`
	SleepEfficiencyPercentage  float64      `json:"sleep_efficiency_percentage" csv:"sleep_efficiency_percentage"`
}
type SleepCollectionRecords struct {
	ID             int       `json:"id" csv:"id"`
	UserID         int       `json:"user_id" csv:"user_id"`
	CreatedAt      time.Time `json:"created_at" csv:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" csv:"updated_at"`
	Start          time.Time `json:"start" csv:"start"`
	End            time.Time `json:"end" csv:"end"`
	TimezoneOffset string    `json:"timezone_offset" csv:"timezone_offset"`
	Nap            bool      `json:"nap" csv:"nap"`
	ScoreState     string    `json:"score_state" csv:"score_state"`
	Score          Score     `json:"score" csv:"score"`
}

type CycleCollection struct {
	Records   []CycleRecords `json:"records" csv:"records"`
	NextToken *string        `json:"next_token,omitempty" csv:"next_token,omitempty"`
}
type CycleScore struct {
	Strain           float64 `json:"strain" csv:"strain"`
	Kilojoule        float64 `json:"kilojoule" csv:"kilojoule"`
	AverageHeartRate int     `json:"average_heart_rate" csv:"average_heart_rate"`
	MaxHeartRate     int     `json:"max_heart_rate" csv:"max_heart_rate"`
}
type CycleRecords struct {
	ID             int        `json:"id" csv:"id"`
	UserID         int        `json:"user_id" csv:"user_id"`
	CreatedAt      time.Time  `json:"created_at" csv:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" csv:"updated_at"`
	Start          time.Time  `json:"start" csv:"start"`
	End            time.Time  `json:"end" csv:"end"`
	TimezoneOffset string     `json:"timezone_offset" csv:"timezone_offset"`
	ScoreState     string     `json:"score_state" csv:"score_state"`
	Score          CycleScore `json:"score" csv:"score"`
}

type RecoveryCollection struct {
	RecoveryRecords []RecoveryRecords `json:"records" csv:"records"`
	NextToken       *string           `json:"next_token,omitempty"`
}
type RecoveryScore struct {
	UserCalibrating  bool    `json:"user_calibrating" csv:"user_calibrating"`
	RecoveryScore    float64 `json:"recovery_score" csv:"recovery_score"`
	RestingHeartRate float64 `json:"resting_heart_rate" csv:"resting_heart_rate"`
	HrvRmssdMilli    float64 `json:"hrv_rmssd_milli" csv:"hrv_rmssd_milli"`
	Spo2Percentage   float64 `json:"spo2_percentage" csv:"spo2_percentage"`
	SkinTempCelsius  float64 `json:"skin_temp_celsius" csv:"skin_temp_celsius"`
}
type RecoveryRecords struct {
	CycleID    int           `json:"cycle_id" csv:"cycle_id"`
	SleepID    int           `json:"sleep_id" csv:"sleep_id"`
	UserID     int           `json:"user_id" csv:"user_id"`
	CreatedAt  time.Time     `json:"created_at" csv:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at" csv:"updated_at"`
	ScoreState string        `json:"score_state" csv:"score_state"`
	Score      RecoveryScore `json:"score" csv:"score"`
}

type WorkoutCollection struct {
	Records   []WorkoutRecords `json:"records" csv:"records"`
	NextToken *string          `json:"next_token,omitempty"`
}
type ZoneDuration struct {
	ZoneZeroMilli  int `json:"zone_zero_milli" csv:"zone_zero_milli"`
	ZoneOneMilli   int `json:"zone_one_milli" csv:"zone_one_milli"`
	ZoneTwoMilli   int `json:"zone_two_milli" csv:"zone_two_milli"`
	ZoneThreeMilli int `json:"zone_three_milli" csv:"zone_three_milli"`
	ZoneFourMilli  int `json:"zone_four_milli" csv:"zone_four_milli"`
	ZoneFiveMilli  int `json:"zone_five_milli" csv:"zone_five_milli"`
}
type WorkoutScore struct {
	Strain              float64      `json:"strain" csv:"strain"`
	AverageHeartRate    int          `json:"average_heart_rate" csv:"average_heart_rate"`
	MaxHeartRate        int          `json:"max_heart_rate" csv:"max_heart_rate"`
	Kilojoule           float64      `json:"kilojoule" csv:"kilojoule"`
	PercentRecorded     float64      `json:"percent_recorded" csv:"percent_recorded"`
	DistanceMeter       float64      `json:"distance_meter" csv:"distance_meter"`
	AltitudeGainMeter   float64      `json:"altitude_gain_meter" csv:"altitude_gain_meter"`
	AltitudeChangeMeter float64      `json:"altitude_change_meter" csv:"altitude_change_meter"`
	ZoneDuration        ZoneDuration `json:"zone_duration" csv:"zone_duration"`
}
type WorkoutRecords struct {
	ID             int          `json:"id" csv:"id"`
	UserID         int          `json:"user_id" csv:"user_id"`
	CreatedAt      time.Time    `json:"created_at" csv:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at" csv:"updated_at"`
	Start          time.Time    `json:"start" csv:"start"`
	End            time.Time    `json:"end" csv:"end"`
	TimezoneOffset string       `json:"timezone_offset" csv:"timezone_offset"`
	SportID        int          `json:"sport_id" csv:"sport_id"`
	SportName      string       `json:"sport_name" csv:"sport_name"`
	ScoreState     string       `json:"score_state" csv:"score_state"`
	Score          WorkoutScore `json:"score" csv:"score"`
}

/*

Authorization Data Structures

*/

type AuthRequest struct {
	// Authorization URL for the Whoop API
	AuthorizationURL string
	// AuthToken is the OAuth2 token for the Whoop API
	AuthToken string
	// RefreshToken is the OAuth2 refresh token for the Whoop API
	RefreshToken string
	// client is the HTTP client to use for making requests
	Client *http.Client
	// The client ID for the Whoop API
	ClientID string
	// The client secret for the Whoop API
	ClientSecret string
	//Token URL for the Whoop API
	TokenURL string
}

/*
* Configuration File Data Structures
 */

type ConfigurationData struct {
	// Credentials is the configuration settings for Whoop API authentication credentials
	Credentials Credentials `yaml:"credentials"`
	// Debug flag. Allowed values are DEBUG, WARN, INFO, TRACE
	Debug string `yaml:"debug"`
	// Export is the configuration block for setting up data exporters
	Export ConfigExport `yaml:"export" validate:"required"`
	// Notification is the configuration block for setting up notifications
	Notification NotificationConfig `yaml:"notification"`
	// Server is the configuration settings for server mode
	Server Server `yaml:"server"`
}

type ConfigExport struct {
	Method     string            `yaml:"method" validate:"oneof=file s3"`
	FileExport export.FileExport `yaml:"fileExport" validate:"required_if=Method file"`
	AWSS3      export.AWS_S3     `yaml:"awsS3" validate:"required_if=Method s3"`
	// Add more supported export methods here
}

type NotificationConfig struct {
	// Method is the notification method to use. If no method is specified, then no external notification is sent.
	Method string `yaml:"method" validate:"oneof=ntfy ''"`
	// Ntfy is the configuration settings for the Ntfy notification service.
	Ntfy notifications.Ntfy `yaml:"ntfy" validate:"required_if=Method ntfy"`
}

type Server struct {
	// Set to true to enable server mode. Default is false.
	Enabled bool `yaml:"enabled"`
	// A cron tab string to schedule the server to run at specific times. Default is every 24 hours at 1300 hours -  0 13 * * *.
	Crontab string `yaml:"crontab"`
	//JWTRefreshDuration is the duration to refresh the JWT token in minutes. Default is 45 minutes. This value must be greater than 0 and less than 59 minutes.
	JWTRefreshDuration int `yaml:"jwtRefreshDuration"`
}

type Credentials struct {
	// The file path to the credentials file. By default, a local file by the name of "token.json" is looked for.
	CredentialsFile string `yaml:"credentialsFile"`
}

/* Event

Event is a struct that contains the event data for the event. Used to determine the type of event to sennd in the notification.

*/

type Event string

const (
	EventErrors  Event = "errors"
	EventSuccess Event = "success"
	EventAll     Event = "all"
)

// eventFromString converts a string to an Event type.
func EventFromString(s string) Event {

	switch strings.ToLower(s) {
	case "errors":
		return EventErrors
	case "success":
		return EventSuccess
	case "all":
		return EventAll
	default:
		return EventErrors
	}
}

// String returns the string representation of an Event type.
func (e Event) String() string {
	return string(e)
}

// Export is the interface for exporting data
type Export interface {
	Setup() error
	Export(data []byte) error
	CleanUp() error
}

// Notification is an interface that defines the methods for a notification service.
// It requires two method functions SetUp and Send.
// Consumers can use the Publish method to send notifications using the notification service.
type Notification interface {
	// SetUp sets up the notification service and returns an error if the setup fails.
	SetUp() error
	// Send sends a notification using the notification service with the provided data and event.
	Publish(client *http.Client, data []byte, event string) error
}

// InitActivityIDTable initializes the activity ID table and returns a map of activity IDs and their corresponding names.
func InitActivityIDTable() map[int]string {

	return map[int]string{
		-1:  "Activity",
		0:   "Running",
		1:   "Cycling",
		16:  "Baseball",
		17:  "Basketball",
		18:  "Rowing",
		19:  "Fencing",
		20:  "Field Hockey",
		21:  "Football",
		22:  "Golf",
		24:  "Ice Hockey",
		25:  "Lacrosse",
		27:  "Rugby",
		28:  "Sailing",
		29:  "Skiing",
		30:  "Soccer",
		31:  "Softball",
		32:  "Squash",
		33:  "Swimming",
		34:  "Tennis",
		35:  "Track & Field",
		36:  "Volleyball",
		37:  "Water Polo",
		38:  "Wrestling",
		39:  "Boxing",
		42:  "Dance",
		43:  "Pilates",
		44:  "Yoga",
		45:  "Weightlifting",
		47:  "Cross Country Skiing",
		48:  "Functional Fitness",
		49:  "Duathlon",
		51:  "Gymnastics",
		52:  "Hiking/Rucking",
		53:  "Horseback Riding",
		55:  "Kayaking",
		56:  "Martial Arts",
		57:  "Mountain Biking",
		59:  "Powerlifting",
		60:  "Rock Climbing",
		61:  "Paddleboarding",
		62:  "Triathlon",
		63:  "Walking",
		64:  "Surfing",
		65:  "Elliptical",
		66:  "Stairmaster",
		70:  "Meditation",
		71:  "Other",
		73:  "Diving",
		74:  "Operations - Tactical",
		75:  "Operations - Medical",
		76:  "Operations - Flying",
		77:  "Operations - Water",
		82:  "Ultimate",
		83:  "Climber",
		84:  "Jumping Rope",
		85:  "Australian Football",
		86:  "Skateboarding",
		87:  "Coaching",
		88:  "Ice Bath",
		89:  "Commuting",
		90:  "Gaming",
		91:  "Snowboarding",
		92:  "Motocross",
		93:  "Caddying",
		94:  "Obstacle Course Racing",
		95:  "Motor Racing",
		96:  "HIIT",
		97:  "Spin",
		98:  "Jiu Jitsu",
		99:  "Manual Labor",
		100: "Cricket",
		101: "Pickleball",
		102: "Inline Skating",
		103: "Box Fitness",
		104: "Spikeball",
		105: "Wheelchair Pushing",
		106: "Paddle Tennis",
		107: "Barre",
		108: "Stage Performance",
		109: "High Stress Work",
		110: "Parkour",
		111: "Gaelic Football",
		112: "Hurling/Camogie",
		113: "Circus Arts",
		121: "Massage Therapy",
		123: "Strength Trainer",
		125: "Watching Sports",
		126: "Assault Bike",
		127: "Kickboxing",
		128: "Stretching",
		230: "Table Tennis",
		231: "Badminton",
		232: "Netball",
		233: "Sauna",
		234: "Disc Golf",
		235: "Yard Work",
		236: "Air Compression",
		237: "Percussive Massage",
		238: "Paintball",
		239: "Ice Skating",
		240: "Handball",
		248: "F45 Training",
		249: "Padel",
		250: "Barry's",
		251: "Dedicated Parenting",
		252: "Stroller Walking",
		253: "Stroller Jogging",
		254: "Toddlerwearing",
		255: "Babywearing",
		258: "Barre3",
		259: "Hot Yoga",
		261: "Stadium Steps",
		262: "Polo",
		263: "Musical Performance",
		264: "Kite Boarding",
		266: "Dog Walking",
		267: "Water Skiing",
		268: "Wakeboarding",
		269: "Cooking",
		270: "Cleaning",
		272: "Public Speaking",
	}
}
