// Package core provides job search operations for Twitter/X.
package core

import (
	"github.com/benoitpetit/xsh/models"
)

// Hardcoded Query IDs for jobs API
const (
	QueryJobSearch = "jVMK9qcOUB5xQQdSLr5ECg"
	QueryJobDetail = "8uZH_OBKTFNIMzTJaV5lbQ"
)

// SearchJobs searches for job listings on Twitter/X
func SearchJobs(
	client *XClient,
	keyword,
	location string,
	locationType,
	employmentType,
	seniorityLevel []string,
	company,
	industry string,
	count int,
	cursor string,
) (*models.JobSearchResponse, error) {

	if count > 25 {
		count = 25 // Twitter's limit
	}

	// Build searchParams dynamically to avoid sending null values
	searchParams := map[string]interface{}{
		"keyword": keyword,
	}
	if location != "" {
		searchParams["job_location"] = location
	}
	if len(locationType) > 0 {
		searchParams["job_location_type"] = locationType
	}
	if len(seniorityLevel) > 0 {
		searchParams["seniority_level"] = seniorityLevel
	}
	if company != "" {
		searchParams["company_name"] = company
	}
	if len(employmentType) > 0 {
		searchParams["employment_type"] = employmentType
	}
	if industry != "" {
		searchParams["industry"] = industry
	}

	variables := map[string]interface{}{
		"count":        count,
		"searchParams": searchParams,
	}

	if cursor != "" {
		variables["cursor"] = cursor
	}

	data, err := client.GraphQLGetRaw(QueryJobSearch, "JobSearchQueryScreenJobsQuery", variables)
	if err != nil {
		return nil, err
	}

	return parseJobSearch(data), nil
}

// parseJobSearch parses the job search response
func parseJobSearch(data map[string]interface{}) *models.JobSearchResponse {
	response := &models.JobSearchResponse{
		Jobs: []models.Job{},
	}

	// Navigate to job_search
	dataMap, ok := data["data"].(map[string]interface{})
	if !ok {
		return response
	}

	jobSearch, ok := dataMap["job_search"].(map[string]interface{})
	if !ok {
		return response
	}

	// Extract items
	items, ok := jobSearch["items_results"].([]interface{})
	if ok {
		for _, item := range items {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			job := models.JobFromSearchResult(itemMap)
			if job != nil {
				response.Jobs = append(response.Jobs, *job)
			}
		}
	}

	// Extract next cursor
	if sliceInfo, ok := jobSearch["slice_info"].(map[string]interface{}); ok {
		response.NextCursor, _ = sliceInfo["next_cursor"].(string)
		response.HasMore = response.NextCursor != ""
	}

	return response
}

// GetJobDetail fetches detailed information about a specific job
func GetJobDetail(client *XClient, jobID string) (*models.Job, error) {
	variables := map[string]interface{}{
		"jobId":    jobID,
		"loggedIn": true,
	}

	data, err := client.GraphQLGetRaw(QueryJobDetail, "JobScreenQuery", variables)
	if err != nil {
		return nil, err
	}

	return models.JobFromDetailResult(data), nil
}
