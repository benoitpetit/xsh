// Package cmd provides job search commands for xsh.
package cmd

import (
	"fmt"
	"os"

	"github.com/benoitpetit/xsh/core"
	"github.com/benoitpetit/xsh/display"
	"github.com/spf13/cobra"
)

var (
	jobLocation       string
	jobLocationType   []string
	jobEmploymentType []string
	jobSeniority      []string
	jobCompany        string
	jobIndustry       string
	jobCount          int
	jobPages          int
)

// jobsCmd represents the jobs command group
var jobsCmd = &cobra.Command{
	Use:   "jobs",
	Short: "Job search commands",
	Long:  `Search for job listings on Twitter/X.`,
}

// jobsSearchCmd searches for jobs
var jobsSearchCmd = &cobra.Command{
	Use:   "search <keyword>",
	Short: "Search for jobs",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		keyword := args[0]

		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(core.ExitAuthError)
		}
		defer client.Close()

		var allJobs []interface{} // Use interface{} for display compatibility
		cursor := ""

		for i := 0; i < jobPages; i++ {
			response, err := core.SearchJobs(
				client,
				keyword,
				jobLocation,
				jobLocationType,
				jobEmploymentType,
				jobSeniority,
				jobCompany,
				jobIndustry,
				jobCount,
				cursor,
			)
			if err != nil {
				fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
				os.Exit(1)
			}

			for _, job := range response.Jobs {
				// Skip jobs without a title (empty results from API)
				if job.Title != "" {
					allJobs = append(allJobs, job)
				}
			}

			cursor = response.NextCursor
			if !response.HasMore {
				break
			}
		}

		output(allJobs, func() {
			fmt.Println(display.FormatJobs(allJobs))
		})
	},
}

// jobsViewCmd views a specific job
var jobsViewCmd = &cobra.Command{
	Use:   "view <job-id>",
	Short: "View job details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		jobID := args[0]

		client, err := getClient("")
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(core.ExitAuthError)
		}
		defer client.Close()

		job, err := core.GetJobDetail(client, jobID)
		if err != nil {
			fmt.Println(display.Error(fmt.Sprintf("Error: %v", err)))
			os.Exit(1)
		}

		if job == nil {
			fmt.Println(display.Error(fmt.Sprintf("Job %s not found", jobID)))
			os.Exit(1)
		}

		output(job, func() {
			fmt.Println(display.FormatJobDetail(job))
		})
	},
}

func init() {
	rootCmd.AddCommand(jobsCmd)
	jobsCmd.AddCommand(jobsSearchCmd)
	jobsCmd.AddCommand(jobsViewCmd)

	// Flags for search
	jobsSearchCmd.Flags().StringVarP(&jobLocation, "location", "l", "", "Location filter")
	jobsSearchCmd.Flags().StringArrayVar(&jobLocationType, "location-type", nil, "Location type: remote, onsite, hybrid")
	jobsSearchCmd.Flags().StringArrayVarP(&jobEmploymentType, "employment-type", "e", nil, "Employment type: full_time, part_time, contract, internship")
	jobsSearchCmd.Flags().StringArrayVarP(&jobSeniority, "seniority", "s", nil, "Seniority: entry_level, mid_level, senior")
	jobsSearchCmd.Flags().StringVar(&jobCompany, "company", "", "Company filter")
	jobsSearchCmd.Flags().StringVar(&jobIndustry, "industry", "", "Industry filter")
	jobsSearchCmd.Flags().IntVarP(&jobCount, "count", "n", 25, "Results per page")
	jobsSearchCmd.Flags().IntVarP(&jobPages, "pages", "p", 1, "Number of pages to fetch")
}
