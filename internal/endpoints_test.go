// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
)

func TestGenerateBackoff(t *testing.T) {

	expected := &backoff.ExponentialBackOff{
		InitialInterval:     DEFAULT_RETRY_INITIAL_INTERVAL,
		RandomizationFactor: DEFAULT_RETRY_RANDOMIZATION,
		Multiplier:          DEFAULT_RETRY_MULTIPLIER,
		MaxElapsedTime:      DEFAULT_RETRY_MAX_ELAPSED_TIME,
	}
	// Test the backoff generation
	backoff := generateBackoff()
	if backoff.InitialInterval != expected.InitialInterval {
		t.Errorf("Expected %v, got %v", expected.InitialInterval, backoff.InitialInterval)
	}
	if backoff.RandomizationFactor != expected.RandomizationFactor {
		t.Errorf("Expected %v, got %v", expected.RandomizationFactor, backoff.RandomizationFactor)
	}

	if backoff.Multiplier != expected.Multiplier {
		t.Errorf("Expected %v, got %v", expected.Multiplier, backoff.Multiplier)
	}

	if backoff.MaxElapsedTime != expected.MaxElapsedTime {
		t.Errorf("Expected %v, got %v", expected.MaxElapsedTime, backoff.MaxElapsedTime)
	}

}

func TestGetUserProfileData(t *testing.T) {

	tests := []struct {
		id            int
		userData      UserData
		ts            *httptest.Server
		errorExpected bool
	}{
		{
			0, UserData{
				FirstName: "testName",
				LastName:  "testLastName",
				UserID:    11111112,
				Email:     "test@email.com",
			},
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Method = "GET"
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{
					"user_id": 11111112,
					"email": "test@email.com",
					"first_name": "testName",
					"last_name": "testLastName"
				}`))
			})),
			false,
		},
		{
			0, UserData{},
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Method = "GET"
				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{
					"error": "Internal Server Error"`))
			})),
			true,
		},
		{
			0, UserData{},
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Method = "GET"
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{
					"unexpect: 42`))
			})),
			true,
		},
	}

	client := CreateHTTPClient()

	for index, test := range tests {
		test.id = index + 1
		defer test.ts.Close()
		ctx := context.Background()
		testUser := User{
			UserData: test.userData,
		}

		result, err := testUser.GetUserProfileData(ctx, client, test.ts.URL, "abc", "mock-user-agent")
		if err != nil && !test.errorExpected {
			t.Fatalf("Test Case %d - Unexpected error: %v", test.id, err)
		}

		if result != nil && result.UserID != test.userData.UserID {
			t.Errorf("Expected %v, got %v", test.userData.UserID, result.UserID)
		}

		if result != nil && result.Email != test.userData.Email {
			t.Errorf("Expected %v, got %v", test.userData.Email, result.Email)
		}

		if result != nil && result.FirstName != test.userData.FirstName {
			t.Errorf("Expected %v, got %v", test.userData.FirstName, result.FirstName)
		}

		if result != nil && result.LastName != test.userData.LastName {
			t.Errorf("Expected %v, got %v", test.userData.LastName, result.LastName)
		}

	}

}

func TestGetUserMeasurements(t *testing.T) {

	tests := []struct {
		id              int
		userMeasurement UserMesaurements
		ts              *httptest.Server
		errorExpected   bool
	}{
		{
			0, UserMesaurements{
				HeightMeter:    1.78,
				WeightKilogram: 66.678085,
				MaxHeartRate:   198,
			},
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Method = "GET"
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{
					"height_meter": 1.78,
					"weight_kilogram": 66.678085,
					"max_heart_rate": 198
				}`))
			})),
			false,
		},
		{
			0, UserMesaurements{},
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Method = "GET"
				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{
					"error": "Internal Server Error"`))
			})),
			true,
		},
		{
			0, UserMesaurements{},
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Method = "GET"
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{
					"unexpect: 42`))
			})),
			true,
		},
	}

	client := CreateHTTPClient()

	for index, test := range tests {
		test.id = index + 1
		defer test.ts.Close()
		ctx := context.Background()
		testUser := User{
			UserMesaurements: test.userMeasurement,
		}

		result, err := testUser.GetUserMeasurements(ctx, client, test.ts.URL, "abc", "mock-user-agent")
		if err != nil && !test.errorExpected {
			t.Fatalf("Test Case %d - Unexpected error: %v", test.id, err)
		}

		if result != nil && result.HeightMeter != test.userMeasurement.HeightMeter {
			t.Errorf("Expected %v, got %v", test.userMeasurement.HeightMeter, result.HeightMeter)
		}

		if result != nil && result.WeightKilogram != test.userMeasurement.WeightKilogram {
			t.Errorf("Expected %v, got %v", test.userMeasurement.WeightKilogram, result.WeightKilogram)
		}

		if result != nil && result.MaxHeartRate != test.userMeasurement.MaxHeartRate {
			t.Errorf("Expected %v, got %v", test.userMeasurement.MaxHeartRate, result.MaxHeartRate)
		}

	}

}

func TestGetSleepCollection(t *testing.T) {

	tests := []struct {
		id                  int
		userSleepCollection SleepCollection
		ts                  *httptest.Server
		errorExpected       bool
	}{
		{
			0, SleepCollection{
				SleepCollectionRecords: []SleepCollectionRecords{
					{
						ID:             1053725843,
						UserID:         11111110,
						CreatedAt:      time.Date(2024, 05, 19, 21, 16, 24, 607000000, time.UTC),
						UpdatedAt:      time.Date(2024, 05, 19, 21, 20, 9, 208000000, time.UTC),
						Start:          time.Date(2024, 05, 19, 20, 31, 01, 867000000, time.UTC),
						End:            time.Date(2024, 05, 19, 21, 16, 25, 295000000, time.UTC),
						TimezoneOffset: "-07:00",
						Nap:            true,
						ScoreState:     "SCORED",
						Score: Score{
							StageSummary: StageSummary{
								TotalInBedTimeMilli:         2723428,
								TotalAwakeTimeMilli:         156757,
								TotalNoDataTimeMilli:        0,
								TotalLightSleepTimeMilli:    1067020,
								TotalSlowWaveSleepTimeMilli: 1499651,
								TotalRemSleepTimeMilli:      0,
								SleepCycleCount:             0,
								DisturbanceCount:            0,
							},
							SleepNeeded: SleepNeeded{
								BaselineMilli:             27192867,
								NeedFromSleepDebtMilli:    1702867,
								NeedFromRecentStrainMilli: 309053,
								NeedFromRecentNapMilli:    0,
							},
							RespiratoryRate:            16.113281,
							SleepPerformancePercentage: 9.0,
							SleepConsistencyPercentage: 56.0,
							SleepEfficiencyPercentage:  99.670616,
						},
					},
				},
				NextToken: "",
			},
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Method = "GET"
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"records":[{
					"id": 1053725843,
					"user_id": 11111110,
					"created_at": "2024-05-19T21:16:24.607Z",
					"updated_at": "2024-05-19T21:20:09.208Z",
					"start": "2024-05-19T20:31:01.867Z",
					"end": "2024-05-19T21:16:25.295Z",
					"timezone_offset": "-07:00",
					"nap": true,
					"score_state": "SCORED",
					"score": {
						"stage_summary": {
							"total_in_bed_time_milli": 2723428,
							"total_awake_time_milli": 156757,
							"total_no_data_time_milli": 0,
							"total_light_sleep_time_milli": 1067020,
							"total_slow_wave_sleep_time_milli": 1499651,
							"total_rem_sleep_time_milli": 0,
							"sleep_cycle_count": 0,
							"disturbance_count": 0
						},
						"sleep_needed": {
							"baseline_milli": 27192867,
							"need_from_sleep_debt_milli": 1702867,
							"need_from_recent_strain_milli": 309053,
							"need_from_recent_nap_milli": 0
						},
						"respiratory_rate": 16.113281,
						"sleep_performance_percentage": 9.0,
						"sleep_consistency_percentage": 56.0,
						"sleep_efficiency_percentage": 99.670616
					}
				}]}`))
			})),
			false,
		},
		{
			0, SleepCollection{
				SleepCollectionRecords: []SleepCollectionRecords{
					{
						ID:             1052702379,
						UserID:         11111110,
						CreatedAt:      time.Date(2024, 05, 19, 05, 40, 30, 547000000, time.UTC),
						UpdatedAt:      time.Date(2024, 05, 19, 14, 50, 17, 496000000, time.UTC),
						Start:          time.Date(2024, 05, 19, 05, 35, 34, 349000000, time.UTC),
						End:            time.Date(2024, 05, 19, 14, 47, 04, 525000000, time.UTC),
						TimezoneOffset: "-07:00",
						Nap:            false,
						ScoreState:     "SCORED",
						Score: Score{
							StageSummary: StageSummary{
								TotalInBedTimeMilli:         33090176,
								TotalAwakeTimeMilli:         7591937,
								TotalNoDataTimeMilli:        0,
								TotalLightSleepTimeMilli:    10262516,
								TotalSlowWaveSleepTimeMilli: 5652487,
								TotalRemSleepTimeMilli:      9583236,
								SleepCycleCount:             7,
								DisturbanceCount:            24,
							},
							SleepNeeded: SleepNeeded{
								BaselineMilli:             27192867,
								NeedFromSleepDebtMilli:    7668000,
								NeedFromRecentStrainMilli: 150062,
								NeedFromRecentNapMilli:    -6863785,
							},
							RespiratoryRate:            15.703125,
							SleepPerformancePercentage: 91.0,
							SleepConsistencyPercentage: 52.0,
							SleepEfficiencyPercentage:  77.05682,
						},
					},
				},
				NextToken: "",
			},
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Method = "GET"
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"records":[{
				"id": 1052702379,
				"user_id": 11111110,
				"created_at": "2024-05-19T05:40:30.547Z",
				"updated_at": "2024-05-19T14:50:17.496Z",
				"start": "2024-05-19T05:35:34.349Z",
				"end": "2024-05-19T14:47:04.525Z",
				"timezone_offset": "-07:00",
				"nap": false,
				"score_state": "SCORED",
				"score": {
					"stage_summary": {
						"total_in_bed_time_milli": 33090176,
						"total_awake_time_milli": 7591937,
						"total_no_data_time_milli": 0,
						"total_light_sleep_time_milli": 10262516,
						"total_slow_wave_sleep_time_milli": 5652487,
						"total_rem_sleep_time_milli": 9583236,
						"sleep_cycle_count": 7,
						"disturbance_count": 24
					},
					"sleep_needed": {
						"baseline_milli": 27192867,
						"need_from_sleep_debt_milli": 7668000,
						"need_from_recent_strain_milli": 150062,
						"need_from_recent_nap_milli": -6863785
					},
					"respiratory_rate": 15.703125,
					"sleep_performance_percentage": 91.0,
					"sleep_consistency_percentage": 52.0,
					"sleep_efficiency_percentage": 77.05682
				}
				}]}`))
			})),
			false,
		},
		{
			0, SleepCollection{},
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Method = "GET"
				w.WriteHeader(http.StatusForbidden)
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{
					"unexpect: 42`))
			})),
			true,
		},
		{
			0, SleepCollection{},
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Method = "GET"
				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{
					"unexpect: 42`))
			})),
			true,
		},
	}

	client := CreateHTTPClient()

	for index, test := range tests {
		test.id = index + 1
		defer test.ts.Close()
		ctx := context.Background()
		testUser := User{
			SleepCollection: test.userSleepCollection,
		}

		t.Log("Test Case: ", test.id)

		result, err := testUser.GetSleepCollection(ctx, client, test.ts.URL, "abc", "", "mock-user-agent")
		if err != nil && !test.errorExpected {
			t.Fatalf("Test Case %d - Unexpected error: %v", test.id, err)
		}

		if result != nil {

			if len(result.SleepCollectionRecords) != len(test.userSleepCollection.SleepCollectionRecords) {
				t.Errorf("Test Case %d - Expected %v, got %v", test.id, len(test.userSleepCollection.SleepCollectionRecords), len(result.SleepCollectionRecords))
			}

			for i := 0; i < len(result.SleepCollectionRecords); i++ {
				if result.SleepCollectionRecords[i].ID != test.userSleepCollection.SleepCollectionRecords[i].ID {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userSleepCollection.SleepCollectionRecords[i].ID, result.SleepCollectionRecords[i].ID)
				}

				if result.SleepCollectionRecords[i].UserID != test.userSleepCollection.SleepCollectionRecords[i].UserID {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userSleepCollection.SleepCollectionRecords[i].UserID, result.SleepCollectionRecords[i].UserID)
				}

				if result.SleepCollectionRecords[i].CreatedAt != test.userSleepCollection.SleepCollectionRecords[i].CreatedAt {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userSleepCollection.SleepCollectionRecords[i].CreatedAt, result.SleepCollectionRecords[i].CreatedAt)
				}

				if result.SleepCollectionRecords[i].UpdatedAt != test.userSleepCollection.SleepCollectionRecords[i].UpdatedAt {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userSleepCollection.SleepCollectionRecords[i].UpdatedAt, result.SleepCollectionRecords[i].UpdatedAt)
				}

				if result.SleepCollectionRecords[i].Start != test.userSleepCollection.SleepCollectionRecords[i].Start {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userSleepCollection.SleepCollectionRecords[i].Start, result.SleepCollectionRecords[i].Start)
				}

				if result.SleepCollectionRecords[i].End != test.userSleepCollection.SleepCollectionRecords[i].End {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userSleepCollection.SleepCollectionRecords[i].End, result.SleepCollectionRecords[i].End)
				}

				if result.SleepCollectionRecords[i].TimezoneOffset != test.userSleepCollection.SleepCollectionRecords[i].TimezoneOffset {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userSleepCollection.SleepCollectionRecords[i].TimezoneOffset, result.SleepCollectionRecords[i].TimezoneOffset)
				}

				if result.SleepCollectionRecords[i].Nap != test.userSleepCollection.SleepCollectionRecords[i].Nap {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userSleepCollection.SleepCollectionRecords[i].Nap, result.SleepCollectionRecords[i].Nap)
				}

				if result.SleepCollectionRecords[i].ScoreState != test.userSleepCollection.SleepCollectionRecords[i].ScoreState {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userSleepCollection.SleepCollectionRecords[i].ScoreState,
						result.SleepCollectionRecords[i].ScoreState)
				}

				// Score Summary

				if result.SleepCollectionRecords[i].Score.StageSummary.TotalInBedTimeMilli != test.userSleepCollection.SleepCollectionRecords[i].Score.StageSummary.TotalInBedTimeMilli {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userSleepCollection.SleepCollectionRecords[i].Score.StageSummary.TotalInBedTimeMilli,
						result.SleepCollectionRecords[i].Score.StageSummary.TotalInBedTimeMilli)
				}

				if result.SleepCollectionRecords[i].Score.StageSummary.TotalAwakeTimeMilli != test.userSleepCollection.SleepCollectionRecords[i].Score.StageSummary.TotalAwakeTimeMilli {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userSleepCollection.SleepCollectionRecords[i].Score.StageSummary.TotalAwakeTimeMilli,
						result.SleepCollectionRecords[i].Score.StageSummary.TotalAwakeTimeMilli)
				}

				if result.SleepCollectionRecords[i].Score.StageSummary.TotalNoDataTimeMilli != test.userSleepCollection.SleepCollectionRecords[i].Score.StageSummary.TotalNoDataTimeMilli {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userSleepCollection.SleepCollectionRecords[i].Score.StageSummary.TotalNoDataTimeMilli,
						result.SleepCollectionRecords[i].Score.StageSummary.TotalNoDataTimeMilli)
				}

				if result.SleepCollectionRecords[i].Score.StageSummary.TotalLightSleepTimeMilli != test.userSleepCollection.SleepCollectionRecords[i].Score.StageSummary.TotalLightSleepTimeMilli {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userSleepCollection.SleepCollectionRecords[i].Score.StageSummary.TotalLightSleepTimeMilli,
						result.SleepCollectionRecords[i].Score.StageSummary.TotalLightSleepTimeMilli)
				}

				if result.SleepCollectionRecords[i].Score.StageSummary.TotalSlowWaveSleepTimeMilli != test.userSleepCollection.SleepCollectionRecords[i].Score.StageSummary.TotalSlowWaveSleepTimeMilli {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userSleepCollection.SleepCollectionRecords[i].Score.StageSummary.TotalSlowWaveSleepTimeMilli,
						result.SleepCollectionRecords[i].Score.StageSummary.TotalSlowWaveSleepTimeMilli)
				}

				if result.SleepCollectionRecords[i].Score.StageSummary.TotalRemSleepTimeMilli != test.userSleepCollection.SleepCollectionRecords[i].Score.StageSummary.TotalRemSleepTimeMilli {
					t.Errorf("Expected %v, got %v", test.userSleepCollection.SleepCollectionRecords[i].Score.StageSummary.TotalRemSleepTimeMilli,
						result.SleepCollectionRecords[i].Score.StageSummary.TotalRemSleepTimeMilli)
				}

				if result.SleepCollectionRecords[i].Score.StageSummary.SleepCycleCount != test.userSleepCollection.SleepCollectionRecords[i].Score.StageSummary.SleepCycleCount {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userSleepCollection.SleepCollectionRecords[i].Score.StageSummary.SleepCycleCount,
						result.SleepCollectionRecords[i].Score.StageSummary.SleepCycleCount)
				}

				if result.SleepCollectionRecords[i].Score.StageSummary.DisturbanceCount != test.userSleepCollection.SleepCollectionRecords[i].Score.StageSummary.DisturbanceCount {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userSleepCollection.SleepCollectionRecords[i].Score.StageSummary.DisturbanceCount,
						result.SleepCollectionRecords[i].Score.StageSummary.DisturbanceCount)
				}

				// Sleep Needed

				if result.SleepCollectionRecords[i].Score.SleepNeeded.BaselineMilli != test.userSleepCollection.SleepCollectionRecords[i].Score.SleepNeeded.BaselineMilli {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userSleepCollection.SleepCollectionRecords[i].Score.SleepNeeded.BaselineMilli,
						result.SleepCollectionRecords[i].Score.SleepNeeded.BaselineMilli)
				}

				if result.SleepCollectionRecords[i].Score.SleepNeeded.NeedFromSleepDebtMilli != test.userSleepCollection.SleepCollectionRecords[i].Score.SleepNeeded.NeedFromSleepDebtMilli {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userSleepCollection.SleepCollectionRecords[i].Score.SleepNeeded.NeedFromSleepDebtMilli,
						result.SleepCollectionRecords[i].Score.SleepNeeded.NeedFromSleepDebtMilli)
				}

				if result.SleepCollectionRecords[i].Score.SleepNeeded.NeedFromRecentStrainMilli != test.userSleepCollection.SleepCollectionRecords[i].Score.SleepNeeded.NeedFromRecentStrainMilli {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userSleepCollection.SleepCollectionRecords[i].Score.SleepNeeded.NeedFromRecentStrainMilli,
						result.SleepCollectionRecords[i].Score.SleepNeeded.NeedFromRecentStrainMilli)
				}

				if result.SleepCollectionRecords[i].Score.SleepNeeded.NeedFromRecentNapMilli != test.userSleepCollection.SleepCollectionRecords[i].Score.SleepNeeded.NeedFromRecentNapMilli {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userSleepCollection.SleepCollectionRecords[i].Score.SleepNeeded.NeedFromRecentNapMilli,
						result.SleepCollectionRecords[i].Score.SleepNeeded.NeedFromRecentNapMilli)
				}

				// Score

				if result.SleepCollectionRecords[i].Score.RespiratoryRate != test.userSleepCollection.SleepCollectionRecords[i].Score.RespiratoryRate {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userSleepCollection.SleepCollectionRecords[i].Score.RespiratoryRate,
						result.SleepCollectionRecords[i].Score.RespiratoryRate)
				}

				if result.SleepCollectionRecords[i].Score.SleepPerformancePercentage != test.userSleepCollection.SleepCollectionRecords[i].Score.SleepPerformancePercentage {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userSleepCollection.SleepCollectionRecords[i].Score.SleepPerformancePercentage,
						result.SleepCollectionRecords[i].Score.SleepPerformancePercentage)
				}

				if result.SleepCollectionRecords[i].Score.SleepConsistencyPercentage != test.userSleepCollection.SleepCollectionRecords[i].Score.SleepConsistencyPercentage {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userSleepCollection.SleepCollectionRecords[i].Score.SleepConsistencyPercentage,
						result.SleepCollectionRecords[i].Score.SleepConsistencyPercentage)
				}

				if result.SleepCollectionRecords[i].Score.SleepEfficiencyPercentage != test.userSleepCollection.SleepCollectionRecords[i].Score.SleepEfficiencyPercentage {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userSleepCollection.SleepCollectionRecords[i].Score.SleepEfficiencyPercentage,
						result.SleepCollectionRecords[i].Score.SleepEfficiencyPercentage)
				}

			}
		}

	}

}
