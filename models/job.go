// Package models provides data models for job listings.
package models

import (
	"encoding/json"
	"fmt"
)

// JobCompany represents a company that posted a job
type JobCompany struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	LogoURL string `json:"logo_url,omitempty"`
}

// Job represents a job listing on Twitter/X
type Job struct {
	ID                 string     `json:"id"`
	Title              string     `json:"title"`
	Company            JobCompany `json:"company"`
	Location           string     `json:"location,omitempty"`
	LocationType       string     `json:"location_type,omitempty"` // remote, onsite, hybrid
	WorkplaceType      string     `json:"workplace_type,omitempty"` // alias for compatibility
	RedirectURL        string     `json:"redirect_url,omitempty"`
	ApplyURL           string     `json:"apply_url,omitempty"` // alias
	SalaryMin          *int       `json:"salary_min,omitempty"`
	SalaryMax          *int       `json:"salary_max,omitempty"`
	SalaryCurrency     string     `json:"salary_currency,omitempty"`
	FormattedSalary    string     `json:"formatted_salary,omitempty"`
	Salary             string     `json:"salary,omitempty"` // formatted alias
	Team               string     `json:"team,omitempty"`
	Description        string     `json:"description,omitempty"`
	EmploymentType     string     `json:"employment_type,omitempty"`
	Industry           string     `json:"industry,omitempty"`
	SeniorityLevel     string     `json:"seniority_level,omitempty"`
	JobURL             string     `json:"job_url,omitempty"`
	PosterHandle       string     `json:"poster_handle,omitempty"`
	PosterName         string     `json:"poster_name,omitempty"`
	PosterVerified     bool       `json:"poster_verified,omitempty"`
	PosterVerifiedType string     `json:"poster_verified_type,omitempty"`
}

// JobSearchResponse represents the response from a job search
type JobSearchResponse struct {
	Jobs       []Job  `json:"jobs"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// JobCompanyFromAPIData parses company from API response data
func JobCompanyFromAPIData(data map[string]interface{}) JobCompany {
	company := JobCompany{}
	
	if id, ok := data["rest_id"].(string); ok {
		company.ID = id
	}
	
	if core, ok := data["core"].(map[string]interface{}); ok {
		company.Name = GetString(core, "name")
	}
	
	if logo, ok := data["logo"].(map[string]interface{}); ok {
		company.LogoURL = GetString(logo, "normal_url")
	}
	
	return company
}

// JobFromSearchResult creates a Job from search result data
func JobFromSearchResult(data map[string]interface{}) *Job {
	job := &Job{}
	
	// Extract job ID
	if id, ok := data["rest_id"].(string); ok {
		job.ID = id
	} else if id, ok := data["id"].(string); ok {
		job.ID = id
	}
	
	// Try new format first (job_card)
	if content, ok := data["job_card"].(map[string]interface{}); ok {
		return jobFromJobCard(job, content)
	}
	
	// Try standard format (result wrapper)
	result, ok := data["result"].(map[string]interface{})
	if !ok {
		result = data
	}
	
	core, ok := result["core"].(map[string]interface{})
	if !ok {
		// Fallback to old format
		return jobFromLegacyFormat(job, data)
	}
	
	job.Title = GetString(core, "title")
	job.Location = GetString(core, "location")
	job.RedirectURL = GetString(core, "redirect_url")
	job.FormattedSalary = GetString(core, "formatted_salary")
	job.Salary = job.FormattedSalary // alias
	job.Team = GetString(core, "team")
	job.LocationType = GetString(core, "location_type")
	
	// Salary parsing
	if min, ok := core["salary_min"].(float64); ok {
		minInt := int(min)
		job.SalaryMin = &minInt
	}
	if max, ok := core["salary_max"].(float64); ok {
		maxInt := int(max)
		job.SalaryMax = &maxInt
	}
	job.SalaryCurrency = GetString(core, "salary_currency_code")
	
	// Job URL
	job.JobURL = fmt.Sprintf("https://x.com/i/jobs/%s", job.ID)
	if pageURL := GetString(core, "job_page_url"); pageURL != "" {
		job.JobURL = pageURL
	}
	
	// Company
	if companyData, ok := result["company_profile_results"].(map[string]interface{}); ok {
		if companyResult, ok := companyData["result"].(map[string]interface{}); ok {
			job.Company = JobCompanyFromAPIData(companyResult)
		}
	}
	
	// Poster info
	if userData, ok := result["user_results"].(map[string]interface{}); ok {
		if userResult, ok := userData["result"].(map[string]interface{}); ok {
			if userCore, ok := userResult["core"].(map[string]interface{}); ok {
				job.PosterHandle = GetString(userCore, "screen_name")
				job.PosterName = GetString(userCore, "name")
			}
			if verification, ok := userResult["verification"].(map[string]interface{}); ok {
				job.PosterVerified, _ = verification["verified"].(bool)
				job.PosterVerifiedType = GetString(verification, "verified_type")
			}
		}
	}
	
	return job
}

// jobFromJobCard parses job from job_card format
func jobFromJobCard(job *Job, content map[string]interface{}) *Job {
	// Title
	if title, ok := content["title"].(map[string]interface{}); ok {
		job.Title = GetString(title, "text")
	}
	
	// Company
	if company, ok := content["company"].(map[string]interface{}); ok {
		job.Company.Name = GetString(company, "text")
	}
	
	// Location
	if location, ok := content["location"].(map[string]interface{}); ok {
		job.Location = GetString(location, "text")
	}
	
	// Workplace type
	if wt, ok := content["workplace_type"].(map[string]interface{}); ok {
		job.WorkplaceType = GetString(wt, "text")
		job.LocationType = job.WorkplaceType
	}
	
	// Description
	if desc, ok := content["description"].(map[string]interface{}); ok {
		job.Description = GetString(desc, "text")
	}
	
	// Apply URL
	if apply, ok := content["apply_url"].(map[string]interface{}); ok {
		job.ApplyURL = GetString(apply, "url")
		job.RedirectURL = job.ApplyURL
	}
	
	// Employment type
	if et, ok := content["employment_type"].(map[string]interface{}); ok {
		job.EmploymentType = GetString(et, "text")
	}
	
	// Salary
	if salary, ok := content["salary"].(map[string]interface{}); ok {
		job.Salary = GetString(salary, "text")
		job.FormattedSalary = job.Salary
	}
	
	// Job URL
	job.JobURL = fmt.Sprintf("https://x.com/i/jobs/%s", job.ID)
	
	return job
}

// jobFromLegacyFormat handles old API format
func jobFromLegacyFormat(job *Job, data map[string]interface{}) *Job {
	if core, ok := data["core"].(map[string]interface{}); ok {
		job.Title = GetString(core, "title")
		job.Location = GetString(core, "location")
	}
	
	if company, ok := data["company"].(map[string]interface{}); ok {
		job.Company.Name = GetString(company, "text")
	}
	
	job.JobURL = fmt.Sprintf("https://x.com/i/jobs/%s", job.ID)
	
	return job
}

// JobFromDetailResult creates a Job from detail result data
func JobFromDetailResult(data map[string]interface{}) *Job {
	jobData, ok := data["data"].(map[string]interface{})
	if !ok {
		return nil
	}
	
	// Try jobData format
	jobResult, ok := jobData["jobData"].(map[string]interface{})
	if !ok {
		// Try direct job format
		jobResult, ok = jobData["job"].(map[string]interface{})
		if !ok {
			return nil
		}
	}
	
	job := JobFromSearchResult(jobResult)
	if job == nil {
		return nil
	}
	
	// Parse detailed description if available
	result, ok := jobResult["result"].(map[string]interface{})
	if !ok {
		result = jobResult
	}
	
	core, ok := result["core"].(map[string]interface{})
	if !ok {
		return job
	}
	
	// Extended fields for detail view
	job.LocationType = GetString(core, "location_type")
	job.WorkplaceType = job.LocationType
	job.RedirectURL = GetString(core, "external_url")
	job.ApplyURL = job.RedirectURL
	
	// Parse job description (Draft.js format)
	rawDesc := GetString(core, "job_description")
	if rawDesc != "" {
		var descData map[string]interface{}
		if err := json.Unmarshal([]byte(rawDesc), &descData); err == nil {
			// Will be converted to markdown by article.go
			job.Description = rawDesc
		} else {
			job.Description = rawDesc
		}
	}
	
	if pageURL := GetString(core, "job_page_url"); pageURL != "" {
		job.JobURL = pageURL
	}
	
	return job
}
