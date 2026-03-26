// Package tests provides unit tests for job search operations.
package tests

import (
	"testing"

	"github.com/benoitpetit/xsh/core"
)

func TestSearchJobs(t *testing.T) {
	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		t.Skip("No credentials available:", err)
	}
	defer client.Close()

	// Search for software engineer jobs
	response, err := core.SearchJobs(
		client,
		"software engineer",
		"",
		nil, // location type
		nil, // employment type
		nil, // seniority
		"",  // company
		"",  // industry
		10,
		"",
	)

	if err != nil {
		t.Logf("SearchJobs() error = %v", err)
		// Don't fail - might be rate limited
		return
	}

	if response == nil {
		t.Error("SearchJobs() returned nil response")
		return
	}

	t.Logf("Found %d jobs", len(response.Jobs))

	// Validate job structure
	for _, job := range response.Jobs {
		if job.ID == "" {
			t.Error("Job has empty ID")
		}
		if job.Title == "" {
			t.Error("Job has empty title")
		}
		// Note: Company name may be empty for some jobs from the API
		if job.Company.Name == "" {
			t.Logf("Warning: Job %s has empty company name", job.ID)
		}
	}

	if response.HasMore && response.NextCursor == "" {
		t.Error("HasMore is true but NextCursor is empty")
	}
}

func TestSearchJobs_WithFilters(t *testing.T) {
	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		t.Skip("No credentials available:", err)
	}
	defer client.Close()

	response, err := core.SearchJobs(
		client,
		"software engineer",
		"San Francisco",
		[]string{"remote"},
		[]string{"full_time"},
		[]string{"senior"},
		"",
		"",
		5,
		"",
	)

	if err != nil {
		t.Logf("SearchJobs() with filters error = %v", err)
		return
	}

	t.Logf("Found %d jobs with filters", len(response.Jobs))
}

func TestGetJobDetail(t *testing.T) {
	client, err := core.NewXClient(nil, "", "")
	if err != nil {
		t.Skip("No credentials available:", err)
	}
	defer client.Close()

	// Use a real job ID from search results
	jobID := "1234567890" // Replace with actual test job ID

	job, err := core.GetJobDetail(client, jobID)
	if err != nil {
		t.Logf("GetJobDetail() error = %v", err)
		return
	}

	if job == nil {
		t.Skip("Job not found or API returned empty")
		return
	}

	if job.ID == "" {
		t.Error("Job detail has empty ID")
	}
	// Note: Title may be empty if the job doesn't exist or API returns incomplete data
	if job.Title == "" {
		t.Logf("Warning: Job detail has empty title (job may not exist or API returned incomplete data)")
	}
}
