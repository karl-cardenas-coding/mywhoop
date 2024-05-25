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
				_, err := w.Write([]byte(`{
					"user_id": 11111112,
					"email": "test@email.com",
					"first_name": "testName",
					"last_name": "testLastName"
				}`))
				if err != nil {
					t.Fatalf("Error writing response: %v", err)
				}
			})),
			false,
		},
		{
			0, UserData{},
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Method = "GET"
				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write([]byte(`{
					"error": "Internal Server Error"`))
				if err != nil {
					t.Fatalf("Error writing response: %v", err)
				}
			})),
			true,
		},
		{
			0, UserData{},
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Method = "GET"
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write([]byte(`{
					"unexpect: 42`))
				if err != nil {
					t.Fatalf("Error writing response: %v", err)
				}
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
				_, err := w.Write([]byte(`{
					"height_meter": 1.78,
					"weight_kilogram": 66.678085,
					"max_heart_rate": 198
				}`))
				if err != nil {
					t.Fatalf("Error writing response: %v", err)
				}
			})),
			false,
		},
		{
			0, UserMesaurements{},
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Method = "GET"
				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write([]byte(`{
					"error": "Internal Server Error"`))
				if err != nil {
					t.Fatalf("Error writing response: %v", err)
				}
			})),
			true,
		},
		{
			0, UserMesaurements{},
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Method = "GET"
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write([]byte(`{
					"unexpect: 42`))
				if err != nil {
					t.Fatalf("Error writing response: %v", err)
				}
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
				_, err := w.Write([]byte(`{"records":[{
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
				if err != nil {
					t.Fatalf("Error writing response: %v", err)
				}
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
				_, err := w.Write([]byte(`{"records":[{
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
				if err != nil {
					t.Fatalf("Error writing response: %v", err)
				}
			})),
			false,
		},
		{
			0, SleepCollection{},
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Method = "GET"
				w.WriteHeader(http.StatusForbidden)
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write([]byte(`{
					"unexpect: 42`))
				if err != nil {
					t.Fatalf("Error writing response: %v", err)
				}
			})),
			true,
		},
		{
			0, SleepCollection{},
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Method = "GET"
				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write([]byte(`{
					"unexpect: 42`))
				if err != nil {
					t.Fatalf("Error writing response: %v", err)
				}
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

func TestGetRecoveryCollection(t *testing.T) {

	tests := []struct {
		id                     int
		userRecoveryCollection RecoveryCollection
		ts                     *httptest.Server
		errorExpected          bool
	}{
		{
			0, RecoveryCollection{
				RecoveryRecords: []RecoveryRecords{
					{
						CycleID:    578161638,
						SleepID:    1059898994,
						UserID:     14465117,
						CreatedAt:  time.Date(2024, 05, 24, 9, 20, 18, 917000000, time.UTC),
						UpdatedAt:  time.Date(2024, 05, 24, 12, 45, 38, 5000000, time.UTC),
						ScoreState: "SCORED",
						Score: RecoveryScore{
							UserCalibrating:  false,
							RecoveryScore:    66,
							RestingHeartRate: 58,
							HrvRmssdMilli:    27.39852,
							Spo2Percentage:   94.72727,
							SkinTempCelsius:  35.2,
						},
					},
				},
				NextToken: "",
			},
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Method = "GET"
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write([]byte(`{
					"records": [
					  {
						"cycle_id": 578161638,
						"sleep_id": 1059898994,
						"user_id": 14465117,
						"created_at": "2024-05-24T09:20:18.917Z",
						"updated_at": "2024-05-24T12:45:38.005Z",
						"score_state": "SCORED",
						"score": {
						  "user_calibrating": false,
						  "recovery_score": 66,
						  "resting_heart_rate": 58,
						  "hrv_rmssd_milli": 27.39852,
						  "spo2_percentage": 94.72727,
						  "skin_temp_celsius": 35.2
						}
					  }
					]
				  }`))
				if err != nil {
					t.Fatalf("Error writing response: %v", err)
				}
			})),
			false,
		},
		{
			0, RecoveryCollection{
				RecoveryRecords: []RecoveryRecords{
					{
						CycleID:    578161638,
						SleepID:    1059898994,
						UserID:     14465117,
						CreatedAt:  time.Date(2024, 05, 24, 9, 20, 18, 917000000, time.UTC),
						UpdatedAt:  time.Date(2024, 05, 24, 12, 45, 38, 5000000, time.UTC),
						ScoreState: "SCORED",
						Score: RecoveryScore{
							UserCalibrating:  false,
							RecoveryScore:    66,
							RestingHeartRate: 58,
							HrvRmssdMilli:    27.39852,
							Spo2Percentage:   94.72727,
							SkinTempCelsius:  35.2,
						},
					},
				},
				NextToken: "",
			},
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Method = "GET"
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write([]byte(`{
					"records": [
					  {
						"cycle_id": 578161638,
						"sleep_id": 1059898994,
						"user_id": 14465117,
						"created_at": "2024-05-24T09:20:18.917Z",
						"updated_at": "2024-05-24T12:45:38.005Z",
						"score_state": "SCORED",
						"score": {
						  "user_calibrating": false,
						  "recovery_score": 66,
						  "resting_heart_rate": 58,
						  "hrv_rmssd_milli": 27.39852,
						  "spo2_percentage": 94.72727,
						  "skin_temp_celsius": 35.2
						}
					  }
					]
				  }`))
				if err != nil {
					t.Fatalf("Error writing response: %v", err)
				}
			})),
			false,
		},
		{
			0, RecoveryCollection{},
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Method = "GET"
				w.WriteHeader(http.StatusForbidden)
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write([]byte(`{
					"unexpect: 42`))
				if err != nil {
					t.Fatalf("Error writing response: %v", err)
				}

			})),
			true,
		},
		{
			0, RecoveryCollection{},
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Method = "GET"
				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write([]byte(`{
					"unexpect: 42`))
				if err != nil {
					t.Fatalf("Error writing response: %v", err)
				}
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
			RecoveryCollection: test.userRecoveryCollection,
		}

		t.Log("Test Case: ", test.id)

		result, err := testUser.GetRecoveryCollection(ctx, client, test.ts.URL, "abc", "", "mock-user-agent")
		if err != nil && !test.errorExpected {
			t.Fatalf("Test Case %d - Unexpected error: %v", test.id, err)
		}

		if result != nil {

			if len(result.RecoveryRecords) != len(test.userRecoveryCollection.RecoveryRecords) {
				t.Errorf("Test Case %d - Expected %v, got %v", test.id, len(test.userRecoveryCollection.RecoveryRecords), len(result.RecoveryRecords))
			}

			for i := 0; i < len(result.RecoveryRecords); i++ {

				if result.RecoveryRecords[i].CycleID != test.userRecoveryCollection.RecoveryRecords[i].CycleID {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userRecoveryCollection.RecoveryRecords[i].CycleID, result.RecoveryRecords[i].CycleID)
				}

				if result.RecoveryRecords[i].SleepID != test.userRecoveryCollection.RecoveryRecords[i].SleepID {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userRecoveryCollection.RecoveryRecords[i].SleepID, result.RecoveryRecords[i].SleepID)
				}

				if result.RecoveryRecords[i].UserID != test.userRecoveryCollection.RecoveryRecords[i].UserID {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userRecoveryCollection.RecoveryRecords[i].UserID, result.RecoveryRecords[i].UserID)
				}

				if result.RecoveryRecords[i].CreatedAt != test.userRecoveryCollection.RecoveryRecords[i].CreatedAt {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userRecoveryCollection.RecoveryRecords[i].CreatedAt, result.RecoveryRecords[i].CreatedAt)
				}

				if result.RecoveryRecords[i].UpdatedAt != test.userRecoveryCollection.RecoveryRecords[i].UpdatedAt {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userRecoveryCollection.RecoveryRecords[i].UpdatedAt, result.RecoveryRecords[i].UpdatedAt)
				}

				if result.RecoveryRecords[i].ScoreState != test.userRecoveryCollection.RecoveryRecords[i].ScoreState {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userRecoveryCollection.RecoveryRecords[i].ScoreState, result.RecoveryRecords[i].ScoreState)
				}

				if result.RecoveryRecords[i].Score.UserCalibrating != test.userRecoveryCollection.RecoveryRecords[i].Score.UserCalibrating {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userRecoveryCollection.RecoveryRecords[i].Score.UserCalibrating, result.RecoveryRecords[i].Score.UserCalibrating)
				}

				if result.RecoveryRecords[i].Score.RecoveryScore != test.userRecoveryCollection.RecoveryRecords[i].Score.RecoveryScore {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userRecoveryCollection.RecoveryRecords[i].Score.RecoveryScore, result.RecoveryRecords[i].Score.RecoveryScore)
				}

				if result.RecoveryRecords[i].Score.RestingHeartRate != test.userRecoveryCollection.RecoveryRecords[i].Score.RestingHeartRate {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userRecoveryCollection.RecoveryRecords[i].Score.RestingHeartRate, result.RecoveryRecords[i].Score.RestingHeartRate)
				}

				if result.RecoveryRecords[i].Score.HrvRmssdMilli != test.userRecoveryCollection.RecoveryRecords[i].Score.HrvRmssdMilli {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userRecoveryCollection.RecoveryRecords[i].Score.HrvRmssdMilli, result.RecoveryRecords[i].Score.HrvRmssdMilli)
				}

				if result.RecoveryRecords[i].Score.Spo2Percentage != test.userRecoveryCollection.RecoveryRecords[i].Score.Spo2Percentage {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userRecoveryCollection.RecoveryRecords[i].Score.Spo2Percentage, result.RecoveryRecords[i].Score.Spo2Percentage)
				}

				if result.RecoveryRecords[i].Score.SkinTempCelsius != test.userRecoveryCollection.RecoveryRecords[i].Score.SkinTempCelsius {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userRecoveryCollection.RecoveryRecords[i].Score.SkinTempCelsius, result.RecoveryRecords[i].Score.SkinTempCelsius)
				}

			}
		}

	}

}

func TestGetWorkoutCollection(t *testing.T) {

	tests := []struct {
		id                    int
		userWorkoutCollection WorkoutCollection
		ts                    *httptest.Server
		errorExpected         bool
	}{
		{
			0, WorkoutCollection{
				Records: []WorkoutRecords{
					{
						ID:             1059528008,
						UserID:         19999999,
						CreatedAt:      time.Date(2024, 05, 24, 01, 48, 52, 178000000, time.UTC),
						UpdatedAt:      time.Date(2024, 05, 24, 01, 56, 11, 338000000, time.UTC),
						Start:          time.Date(2024, 05, 24, 01, 9, 00, 475000000, time.UTC),
						End:            time.Date(2024, 05, 24, 01, 32, 59, 418000000, time.UTC),
						TimezoneOffset: "-07:00",
						SportID:        71,
						ScoreState:     "SCORED",
						Score: WorkoutScore{
							Strain:              7.9998,
							AverageHeartRate:    132,
							MaxHeartRate:        162,
							Kilojoule:           866.2699,
							PercentRecorded:     100,
							DistanceMeter:       0,
							AltitudeGainMeter:   0,
							AltitudeChangeMeter: 0,
							ZoneDuration: ZoneDuration{
								ZoneZeroMilli:  4806,
								ZoneOneMilli:   159578,
								ZoneTwoMilli:   151924,
								ZoneThreeMilli: 724754,
								ZoneFourMilli:  397920,
								ZoneFiveMilli:  0,
							},
						},
					},
				},
				NextToken: "",
			},
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Method = "GET"
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write([]byte(`{
					"records": [
					  {
						"id": 1059528008,
						"user_id": 19999999,
						"created_at": "2024-05-24T01:48:52.178Z",
						"updated_at": "2024-05-24T01:56:11.338Z",
						"start": "2024-05-24T01:09:00.475Z",
						"end": "2024-05-24T01:32:59.418Z",
						"timezone_offset": "-07:00",
						"sport_id": 71,
						"score_state": "SCORED",
						"score": {
						  "strain": 7.9998,
						  "average_heart_rate": 132,
						  "max_heart_rate": 162,
						  "kilojoule": 866.2699,
						  "percent_recorded": 100,
						  "distance_meter": 0,
						  "altitude_gain_meter": 0,
						  "altitude_change_meter": 0,
						  "zone_duration": {
							"zone_zero_milli": 4806,
							"zone_one_milli": 159578,
							"zone_two_milli": 151924,
							"zone_three_milli": 724754,
							"zone_four_milli": 397920,
							"zone_five_milli": 0
						  }
						}
					  }
					]
				  }`))
				if err != nil {
					t.Fatalf("Error writing response: %v", err)
				}
			})),
			false,
		},
		{
			0, WorkoutCollection{
				Records: []WorkoutRecords{
					{
						ID:             1059528008,
						UserID:         19999999,
						CreatedAt:      time.Date(2024, 05, 24, 01, 48, 52, 178000000, time.UTC),
						UpdatedAt:      time.Date(2024, 05, 24, 01, 56, 11, 338000000, time.UTC),
						Start:          time.Date(2024, 05, 24, 01, 9, 00, 475000000, time.UTC),
						End:            time.Date(2024, 05, 24, 01, 32, 59, 418000000, time.UTC),
						TimezoneOffset: "-07:00",
						SportID:        71,
						ScoreState:     "SCORED",
						Score: WorkoutScore{
							Strain:              7.9998,
							AverageHeartRate:    132,
							MaxHeartRate:        162,
							Kilojoule:           866.2699,
							PercentRecorded:     100,
							DistanceMeter:       0,
							AltitudeGainMeter:   0,
							AltitudeChangeMeter: 0,
							ZoneDuration: ZoneDuration{
								ZoneZeroMilli:  4806,
								ZoneOneMilli:   159578,
								ZoneTwoMilli:   151924,
								ZoneThreeMilli: 724754,
								ZoneFourMilli:  397920,
								ZoneFiveMilli:  0,
							},
						},
					},
				},
				NextToken: "",
			},
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Method = "GET"
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write([]byte(`{
					"records": [
					  {
						"id": 1059528008,
						"user_id": 19999999,
						"created_at": "2024-05-24T01:48:52.178Z",
						"updated_at": "2024-05-24T01:56:11.338Z",
						"start": "2024-05-24T01:09:00.475Z",
						"end": "2024-05-24T01:32:59.418Z",
						"timezone_offset": "-07:00",
						"sport_id": 71,
						"score_state": "SCORED",
						"score": {
						  "strain": 7.9998,
						  "average_heart_rate": 132,
						  "max_heart_rate": 162,
						  "kilojoule": 866.2699,
						  "percent_recorded": 100,
						  "distance_meter": 0,
						  "altitude_gain_meter": 0,
						  "altitude_change_meter": 0,
						  "zone_duration": {
							"zone_zero_milli": 4806,
							"zone_one_milli": 159578,
							"zone_two_milli": 151924,
							"zone_three_milli": 724754,
							"zone_four_milli": 397920,
							"zone_five_milli": 0
						  }
						}
					  }
					]
				  }`))
				if err != nil {
					t.Fatalf("Error writing response: %v", err)
				}
			})),
			false,
		},
		{
			0, WorkoutCollection{},
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Method = "GET"
				w.WriteHeader(http.StatusForbidden)
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write([]byte(`{
					"unexpect: 42`))
				if err != nil {
					t.Fatalf("Error writing response: %v", err)
				}
			})),
			true,
		},
		{
			0, WorkoutCollection{},
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Method = "GET"
				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write([]byte(`{
					"unexpect: 42`))
				if err != nil {
					t.Fatalf("Error writing response: %v", err)
				}
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
			WorkoutCollection: test.userWorkoutCollection,
		}

		t.Log("Test Case: ", test.id)

		result, err := testUser.GetWorkoutCollection(ctx, client, test.ts.URL, "abc", "", "mock-user-agent")
		if err != nil && !test.errorExpected {
			t.Fatalf("Test Case %d - Unexpected error: %v", test.id, err)
		}

		if result != nil {

			if len(result.Records) != len(test.userWorkoutCollection.Records) {
				t.Errorf("Test Case %d - Expected %v, got %v", test.id, len(test.userWorkoutCollection.Records), len(result.Records))
			}

			for i := 0; i < len(result.Records); i++ {

				if result.Records[i].ID != test.userWorkoutCollection.Records[i].ID {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userWorkoutCollection.Records[i].ID, result.Records[i].ID)
				}

				if result.Records[i].UserID != test.userWorkoutCollection.Records[i].UserID {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userWorkoutCollection.Records[i].UserID, result.Records[i].UserID)
				}

				if result.Records[i].CreatedAt != test.userWorkoutCollection.Records[i].CreatedAt {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userWorkoutCollection.Records[i].CreatedAt, result.Records[i].CreatedAt)
				}

				if result.Records[i].UpdatedAt != test.userWorkoutCollection.Records[i].UpdatedAt {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userWorkoutCollection.Records[i].UpdatedAt, result.Records[i].UpdatedAt)
				}

				if result.Records[i].Start != test.userWorkoutCollection.Records[i].Start {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userWorkoutCollection.Records[i].Start, result.Records[i].Start)
				}

				if result.Records[i].End != test.userWorkoutCollection.Records[i].End {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userWorkoutCollection.Records[i].End, result.Records[i].End)
				}

				if result.Records[i].TimezoneOffset != test.userWorkoutCollection.Records[i].TimezoneOffset {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userWorkoutCollection.Records[i].TimezoneOffset, result.Records[i].TimezoneOffset)
				}

				if result.Records[i].SportID != test.userWorkoutCollection.Records[i].SportID {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userWorkoutCollection.Records[i].SportID, result.Records[i].SportID)
				}

				if result.Records[i].ScoreState != test.userWorkoutCollection.Records[i].ScoreState {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userWorkoutCollection.Records[i].ScoreState, result.Records[i].ScoreState)
				}

				if result.Records[i].Score.Strain != test.userWorkoutCollection.Records[i].Score.Strain {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userWorkoutCollection.Records[i].Score.Strain, result.Records[i].Score.Strain)
				}

				if result.Records[i].Score.AverageHeartRate != test.userWorkoutCollection.Records[i].Score.AverageHeartRate {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userWorkoutCollection.Records[i].Score.AverageHeartRate, result.Records[i].Score.AverageHeartRate)
				}

				if result.Records[i].Score.MaxHeartRate != test.userWorkoutCollection.Records[i].Score.MaxHeartRate {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userWorkoutCollection.Records[i].Score.MaxHeartRate, result.Records[i].Score.MaxHeartRate)
				}

				if result.Records[i].Score.Kilojoule != test.userWorkoutCollection.Records[i].Score.Kilojoule {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userWorkoutCollection.Records[i].Score.Kilojoule, result.Records[i].Score.Kilojoule)
				}

				if result.Records[i].Score.PercentRecorded != test.userWorkoutCollection.Records[i].Score.PercentRecorded {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userWorkoutCollection.Records[i].Score.PercentRecorded, result.Records[i].Score.PercentRecorded)
				}

				if result.Records[i].Score.DistanceMeter != test.userWorkoutCollection.Records[i].Score.DistanceMeter {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userWorkoutCollection.Records[i].Score.DistanceMeter, result.Records[i].Score.DistanceMeter)
				}

				if result.Records[i].Score.AltitudeGainMeter != test.userWorkoutCollection.Records[i].Score.AltitudeGainMeter {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userWorkoutCollection.Records[i].Score.AltitudeGainMeter, result.Records[i].Score.AltitudeGainMeter)
				}

				if result.Records[i].Score.AltitudeChangeMeter != test.userWorkoutCollection.Records[i].Score.AltitudeChangeMeter {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userWorkoutCollection.Records[i].Score.AltitudeChangeMeter, result.Records[i].Score.AltitudeChangeMeter)
				}

				if result.Records[i].Score.ZoneDuration.ZoneZeroMilli != test.userWorkoutCollection.Records[i].Score.ZoneDuration.ZoneZeroMilli {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userWorkoutCollection.Records[i].Score.ZoneDuration.ZoneZeroMilli, result.Records[i].Score.ZoneDuration.ZoneZeroMilli)
				}

				if result.Records[i].Score.ZoneDuration.ZoneOneMilli != test.userWorkoutCollection.Records[i].Score.ZoneDuration.ZoneOneMilli {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userWorkoutCollection.Records[i].Score.ZoneDuration.ZoneOneMilli, result.Records[i].Score.ZoneDuration.ZoneOneMilli)
				}

				if result.Records[i].Score.ZoneDuration.ZoneTwoMilli != test.userWorkoutCollection.Records[i].Score.ZoneDuration.ZoneTwoMilli {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userWorkoutCollection.Records[i].Score.ZoneDuration.ZoneTwoMilli, result.Records[i].Score.ZoneDuration.ZoneTwoMilli)
				}

				if result.Records[i].Score.ZoneDuration.ZoneThreeMilli != test.userWorkoutCollection.Records[i].Score.ZoneDuration.ZoneThreeMilli {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userWorkoutCollection.Records[i].Score.ZoneDuration.ZoneThreeMilli, result.Records[i].Score.ZoneDuration.ZoneThreeMilli)
				}

				if result.Records[i].Score.ZoneDuration.ZoneFourMilli != test.userWorkoutCollection.Records[i].Score.ZoneDuration.ZoneFourMilli {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userWorkoutCollection.Records[i].Score.ZoneDuration.ZoneFourMilli, result.Records[i].Score.ZoneDuration.ZoneFourMilli)
				}

				if result.Records[i].Score.ZoneDuration.ZoneFiveMilli != test.userWorkoutCollection.Records[i].Score.ZoneDuration.ZoneFiveMilli {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userWorkoutCollection.Records[i].Score.ZoneDuration.ZoneFiveMilli, result.Records[i].Score.ZoneDuration.ZoneFiveMilli)
				}

			}
		}

	}

}

func TestGetCycleCollection(t *testing.T) {

	tests := []struct {
		id                  int
		userCycleCollection CycleCollection
		ts                  *httptest.Server
		errorExpected       bool
	}{
		{
			0, CycleCollection{
				Records: []CycleRecords{
					{
						ID:             571373432,
						UserID:         19999999,
						CreatedAt:      time.Date(2024, 05, 14, 10, 12, 58, 83000000, time.UTC),
						UpdatedAt:      time.Date(2024, 05, 15, 13, 27, 45, 950000000, time.UTC),
						Start:          time.Date(2024, 05, 14, 05, 15, 14, 199000000, time.UTC),
						End:            time.Date(2024, 05, 15, 05, 06, 01, 83000000, time.UTC),
						TimezoneOffset: "-07:00",
						ScoreState:     "SCORED",
						Score: CycleScore{
							Strain:           14.583735,
							Kilojoule:        9503.399,
							AverageHeartRate: 76,
							MaxHeartRate:     157,
						},
					},
					{
						ID:             570921306,
						UserID:         19999999,
						CreatedAt:      time.Date(2024, 05, 13, 14, 10, 10, 248000000, time.UTC),
						UpdatedAt:      time.Date(2024, 05, 14, 10, 13, 03, 243000000, time.UTC),
						Start:          time.Date(2024, 05, 13, 04, 20, 13, 925000000, time.UTC),
						End:            time.Date(2024, 05, 14, 05, 15, 14, 99000000, time.UTC),
						TimezoneOffset: "-07:00",
						ScoreState:     "SCORED",
						Score: CycleScore{
							Strain:           9.334525,
							Kilojoule:        8259.693,
							AverageHeartRate: 74,
							MaxHeartRate:     154,
						},
					},
				},
				NextToken: "",
			},
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Method = "GET"
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write([]byte(`{
					"records": [
					  {
						"id": 571373432,
						"user_id": 19999999,
						"created_at": "2024-05-14T10:12:58.083Z",
						"updated_at": "2024-05-15T13:27:45.950Z",
						"start": "2024-05-14T05:15:14.199Z",
						"end": "2024-05-15T05:06:01.083Z",
						"timezone_offset": "-07:00",
						"score_state": "SCORED",
						"score": {
						  "strain": 14.583735,
						  "kilojoule": 9503.399,
						  "average_heart_rate": 76,
						  "max_heart_rate": 157
						}
					  },
					  {
						"id": 570921306,
						"user_id": 19999999,
						"created_at": "2024-05-13T14:10:10.248Z",
						"updated_at": "2024-05-14T10:13:03.243Z",
						"start": "2024-05-13T04:20:13.925Z",
						"end": "2024-05-14T05:15:14.099Z",
						"timezone_offset": "-07:00",
						"score_state": "SCORED",
						"score": {
						  "strain": 9.334525,
						  "kilojoule": 8259.693,
						  "average_heart_rate": 74,
						  "max_heart_rate": 154
						}
					  }
					]
				  }`))
				if err != nil {
					t.Fatalf("Error writing response: %v", err)
				}
			})),
			false,
		},
		{
			0, CycleCollection{},
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Method = "GET"
				w.WriteHeader(http.StatusForbidden)
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write([]byte(`{
					"unexpect: 42`))
				if err != nil {
					t.Fatalf("Error writing response: %v", err)
				}
			})),
			true,
		},
		{
			0, CycleCollection{},
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Method = "GET"
				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write([]byte(`{
					"unexpect: 42`))
				if err != nil {
					t.Fatalf("Error writing response: %v", err)
				}
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
			CycleCollection: test.userCycleCollection,
		}

		t.Log("Test Case: ", test.id)

		result, err := testUser.GetCycleCollection(ctx, client, test.ts.URL, "abc", "", "mock-user-agent")
		if err != nil && !test.errorExpected {
			t.Fatalf("Test Case %d - Unexpected error: %v", test.id, err)
		}

		if result != nil {

			if len(result.Records) != len(test.userCycleCollection.Records) {
				t.Errorf("Test Case %d - Expected %v, got %v", test.id, len(test.userCycleCollection.Records), len(result.Records))
			}

			for i := 0; i < len(result.Records); i++ {

				if result.Records[i].ID != test.userCycleCollection.Records[i].ID {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userCycleCollection.Records[i].ID, result.Records[i].ID)
				}

				if result.Records[i].UserID != test.userCycleCollection.Records[i].UserID {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userCycleCollection.Records[i].UserID, result.Records[i].UserID)
				}

				if result.Records[i].CreatedAt != test.userCycleCollection.Records[i].CreatedAt {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userCycleCollection.Records[i].CreatedAt, result.Records[i].CreatedAt)
				}

				if result.Records[i].UpdatedAt != test.userCycleCollection.Records[i].UpdatedAt {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userCycleCollection.Records[i].UpdatedAt, result.Records[i].UpdatedAt)
				}

				if result.Records[i].Start != test.userCycleCollection.Records[i].Start {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userCycleCollection.Records[i].Start, result.Records[i].Start)
				}

				if result.Records[i].End != test.userCycleCollection.Records[i].End {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userCycleCollection.Records[i].End, result.Records[i].End)
				}

				if result.Records[i].TimezoneOffset != test.userCycleCollection.Records[i].TimezoneOffset {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userCycleCollection.Records[i].TimezoneOffset, result.Records[i].TimezoneOffset)
				}

				if result.Records[i].ScoreState != test.userCycleCollection.Records[i].ScoreState {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userCycleCollection.Records[i].ScoreState, result.Records[i].ScoreState)
				}

				if result.Records[i].Score.Strain != test.userCycleCollection.Records[i].Score.Strain {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userCycleCollection.Records[i].Score.Strain, result.Records[i].Score.Strain)
				}

				if result.Records[i].Score.Kilojoule != test.userCycleCollection.Records[i].Score.Kilojoule {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userCycleCollection.Records[i].Score.Kilojoule, result.Records[i].Score.Kilojoule)
				}

				if result.Records[i].Score.AverageHeartRate != test.userCycleCollection.Records[i].Score.AverageHeartRate {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userCycleCollection.Records[i].Score.AverageHeartRate, result.Records[i].Score.AverageHeartRate)
				}

				if result.Records[i].Score.MaxHeartRate != test.userCycleCollection.Records[i].Score.MaxHeartRate {
					t.Errorf("Test Case %d - Expected %v, got %v", test.id, test.userCycleCollection.Records[i].Score.MaxHeartRate, result.Records[i].Score.MaxHeartRate)
				}
			}
		}

	}

}
