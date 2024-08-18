// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"bytes"
	"slices"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		description string
		input       []byte
		expected    string
	}{
		{
			description: "Empty slice should return '0 B'",
			input:       make([]byte, 0),
			expected:    "0 B",
		},
		{
			description: "Size less than 1 KB should return the exact number of bytes",
			input:       make([]byte, 999),
			expected:    "999 B",
		},
		{
			description: "Exactly 1 KB should return '1.0 KB'",
			input:       make([]byte, 1024),
			expected:    "1.0 KB",
		},
		{
			description: "Exactly 1 MB should return '1.0 MB'",
			input:       make([]byte, 1048576),
			expected:    "1.0 MB",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result := formatBytes(test.input)
			if result != test.expected {
				t.Errorf("For %s, expected %s, but got %s", test.description, test.expected, result)
			}
		})
	}
}

func TestConvertToExcel(t *testing.T) {
	tests := []struct {
		name          string
		userData      User
		expectedError bool
	}{
		{
			name: "Valid User Data",
			userData: User{
				UserData: UserData{
					UserID:    111111111,
					FirstName: "John",
					LastName:  "Doe",
					Email:     "john.doe@example.com",
				},
				UserMesaurements: UserMesaurements{
					HeightMeter:    180,
					WeightKilogram: 80,
					MaxHeartRate:   175,
				},
				SleepCollection: SleepCollection{
					SleepCollectionRecords: []SleepCollectionRecords{
						{
							ID:             123456789,
							UserID:         111111111,
							CreatedAt:      time.Date(2024, time.August, 15, 0, 0, 0, 0, time.UTC),
							UpdatedAt:      time.Date(2024, time.August, 15, 1, 0, 0, 0, time.UTC),
							Start:          time.Date(2024, time.August, 15, 0, 0, 0, 0, time.UTC),
							End:            time.Date(2024, time.August, 15, 8, 0, 0, 0, time.UTC),
							TimezoneOffset: "+02:00",
							Nap:            false,
							ScoreState:     "good",
							Score: Score{
								StageSummary: StageSummary{
									TotalInBedTimeMilli:         28800000,
									TotalAwakeTimeMilli:         3600000,
									TotalNoDataTimeMilli:        0,
									TotalLightSleepTimeMilli:    14400000,
									TotalSlowWaveSleepTimeMilli: 7200000,
									TotalRemSleepTimeMilli:      4320000,
									SleepCycleCount:             4,
									DisturbanceCount:            2,
								},
								SleepNeeded: SleepNeeded{
									BaselineMilli:             28800000,
									NeedFromSleepDebtMilli:    0,
									NeedFromRecentStrainMilli: 0,
									NeedFromRecentNapMilli:    0,
								},
								RespiratoryRate:            15,
								SleepPerformancePercentage: 85.5,
								SleepConsistencyPercentage: 80.2,
								SleepEfficiencyPercentage:  90.0,
							},
						},
					},
				},
				RecoveryCollection: RecoveryCollection{
					RecoveryRecords: []RecoveryRecords{
						{
							CycleID:    78942566,
							SleepID:    655468825,
							UserID:     111111111,
							CreatedAt:  time.Date(2024, time.August, 15, 0, 0, 0, 0, time.UTC),
							UpdatedAt:  time.Date(2024, time.August, 15, 1, 0, 0, 0, time.UTC),
							ScoreState: "SCORED",
							Score: RecoveryScore{
								UserCalibrating:  false,
								RecoveryScore:    85,
								RestingHeartRate: 60,
								HrvRmssdMilli:    50,
								Spo2Percentage:   98.0,
								SkinTempCelsius:  36.5,
							},
						},
					},
				},
				WorkoutCollection: WorkoutCollection{
					Records: []WorkoutRecords{
						{
							ID:             48588744545,
							UserID:         111111111,
							CreatedAt:      time.Date(2024, time.August, 15, 0, 6, 25, 0, time.UTC),
							UpdatedAt:      time.Date(2024, time.August, 15, 0, 7, 0, 0, time.UTC),
							Start:          time.Date(2024, time.August, 15, 0, 6, 25, 0, time.UTC),
							End:            time.Date(2024, time.August, 15, 0, 6, 42, 0, time.UTC),
							TimezoneOffset: "-07:00",
							SportID:        44,
							ScoreState:     "SCORED",
							Score: WorkoutScore{
								Strain:              7,
								AverageHeartRate:    150,
								MaxHeartRate:        180,
								Kilojoule:           600,
								PercentRecorded:     100,
								DistanceMeter:       10000,
								AltitudeGainMeter:   500,
								AltitudeChangeMeter: 500,
								ZoneDuration: ZoneDuration{
									ZoneZeroMilli:  0,
									ZoneOneMilli:   600000,
									ZoneTwoMilli:   1200000,
									ZoneThreeMilli: 1800000,
									ZoneFourMilli:  3600000,
									ZoneFiveMilli:  600000,
								},
							},
						},
					},
				},
				CycleCollection: CycleCollection{
					Records: []CycleRecords{
						{
							ID:             123456789,
							UserID:         111111111,
							CreatedAt:      time.Date(2024, time.August, 15, 0, 9, 0, 0, time.UTC),
							UpdatedAt:      time.Date(2024, time.August, 15, 0, 9, 0, 0, time.UTC),
							Start:          time.Date(2024, time.August, 15, 0, 9, 0, 0, time.UTC),
							End:            time.Date(2024, time.August, 15, 0, 10, 0, 0, time.UTC),
							TimezoneOffset: "-07:00",
							ScoreState:     "SCORED",
							Score: CycleScore{
								Strain:           8,
								Kilojoule:        650,
								AverageHeartRate: 155,
								MaxHeartRate:     185,
							},
						},
					},
				},
			},
			expectedError: false,
		},
		{
			name:          "Invalid Excel Sheet Creation",
			userData:      User{},
			expectedError: false,
		},
		{
			name: "Missing Sleep Collection Data",
			userData: User{
				UserData: UserData{
					UserID:    111111111,
					FirstName: "John",
					LastName:  "Doe",
					Email:     "john.doe@example.com",
				},
				UserMesaurements: UserMesaurements{
					HeightMeter:    180,
					WeightKilogram: 80,
					MaxHeartRate:   175,
				},
				RecoveryCollection: RecoveryCollection{
					RecoveryRecords: []RecoveryRecords{
						{
							CycleID:    78942566,
							SleepID:    655468825,
							UserID:     111111111,
							CreatedAt:  time.Date(2024, time.August, 15, 0, 0, 0, 0, time.UTC),
							UpdatedAt:  time.Date(2024, time.August, 15, 1, 0, 0, 0, time.UTC),
							ScoreState: "SCORED",
							Score: RecoveryScore{
								UserCalibrating:  false,
								RecoveryScore:    85,
								RestingHeartRate: 60,
								HrvRmssdMilli:    50,
								Spo2Percentage:   98.0,
								SkinTempCelsius:  36.5,
							},
						},
					},
				},
				WorkoutCollection: WorkoutCollection{
					Records: []WorkoutRecords{
						{
							ID:             48588744545,
							UserID:         111111111,
							CreatedAt:      time.Date(2024, time.August, 15, 0, 6, 25, 0, time.UTC),
							UpdatedAt:      time.Date(2024, time.August, 15, 0, 7, 0, 0, time.UTC),
							Start:          time.Date(2024, time.August, 15, 0, 6, 25, 0, time.UTC),
							End:            time.Date(2024, time.August, 15, 0, 6, 42, 0, time.UTC),
							TimezoneOffset: "-07:00",
							SportID:        44,
							ScoreState:     "SCORED",
							Score: WorkoutScore{
								Strain:              7,
								AverageHeartRate:    150,
								MaxHeartRate:        180,
								Kilojoule:           600,
								PercentRecorded:     100,
								DistanceMeter:       10000,
								AltitudeGainMeter:   500,
								AltitudeChangeMeter: 500,
								ZoneDuration: ZoneDuration{
									ZoneZeroMilli:  0,
									ZoneOneMilli:   600000,
									ZoneTwoMilli:   1200000,
									ZoneThreeMilli: 1800000,
									ZoneFourMilli:  3600000,
									ZoneFiveMilli:  600000,
								},
							},
						},
					},
				},
				CycleCollection: CycleCollection{
					Records: []CycleRecords{
						{
							ID:             123456789,
							UserID:         111111111,
							CreatedAt:      time.Date(2024, time.August, 15, 0, 9, 0, 0, time.UTC),
							UpdatedAt:      time.Date(2024, time.August, 15, 0, 9, 0, 0, time.UTC),
							Start:          time.Date(2024, time.August, 15, 0, 9, 0, 0, time.UTC),
							End:            time.Date(2024, time.August, 15, 0, 10, 0, 0, time.UTC),
							TimezoneOffset: "-07:00",
							ScoreState:     "SCORED",
							Score: CycleScore{
								Strain:           8,
								Kilojoule:        650,
								AverageHeartRate: 155,
								MaxHeartRate:     185,
							},
						},
					},
				},
			},
			expectedError: false,
		},
		// Additional test cases for other edge cases can be added here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertToExcel(tt.userData)

			if tt.expectedError {
				if err == nil {
					t.Errorf("%s - expected an error but got nil", tt.name)
				}
			}

			if !tt.expectedError {

				if err != nil {
					t.Errorf("%s - unexpected error: %v", tt.name, err)
				}
				if got == nil {
					t.Errorf("%s - expected non-nil result but got nil", tt.name)
				}

				f, err := excelize.OpenReader(bytes.NewReader(got))
				if err != nil {
					t.Errorf("%s - failed to open Excel file: %v", tt.name, err)
					return
				}
				defer f.Close()

				// Check if the expected sheets exist
				sheets := []string{"User Data", "Sleep Data", "Recovery Data", "Workout Data", "Cycle Data"}
				actualSheets := f.GetSheetList()

				for _, sheet := range sheets {
					found := slices.Contains(actualSheets, sheet)
					if !found {
						t.Errorf("%s - expected sheet %q not found", tt.name, sheet)
					}
				}
			}

		})
	}
}
