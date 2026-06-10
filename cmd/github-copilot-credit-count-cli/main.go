// Copyright (c) 2026 Heino Stömmer.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github-copilot-credit-count/internal/adapter/repository"
	"github-copilot-credit-count/internal/domain"
	"github-copilot-credit-count/internal/usecase"
)

var Version = "dev"

type cliOutput struct {
	Overall         domain.TokenUsage    `json:"overall"`
	AvailableMonths []string             `json:"availableMonths"`
	MonthSummary    *domain.MonthSummary `json:"monthSummary,omitempty"`
}

func main() {
	// Define flags
	pathFlag := flag.String("path", "", "Custom VS Code workspace storage path")
	formatFlag := flag.String("format", "text", "Output format: 'text' or 'json'")
	monthFlag := flag.String("month", "", "Specific month to show details for (format: YYYY-MM)")
	versionFlag := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("GitHub Copilot Credit Count CLI version %s\n", Version)
		os.Exit(0)
	}

	// Initialize Clean Architecture layers
	repo := repository.NewCopilotLogRepository(*pathFlag)
	uc := usecase.NewTokenUsecase(repo)

	// Scan files
	if err := uc.ScanAndCache(); err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning logs: %v\n", err)
		os.Exit(1)
	}

	overall := uc.GetOverallTotal()
	months := uc.GetAvailableMonths()

	var targetMonthSummary *domain.MonthSummary
	if *monthFlag != "" {
		summary, err := uc.GetSummaryForMonth(*monthFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting summary for month %s: %v\n", *monthFlag, err)
			os.Exit(1)
		}
		targetMonthSummary = &summary
	}

	if *formatFlag == "json" {
		out := cliOutput{
			Overall:         overall,
			AvailableMonths: months,
			MonthSummary:    targetMonthSummary,
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(out); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding JSON output: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Render text output
	fmt.Println("GitHub Copilot Credit Count - CLI")
	fmt.Println("=================================")
	fmt.Println()
	fmt.Println("Overall Usage Across All Logs:")
	printTokenUsage(overall)
	fmt.Println()

	fmt.Println("Available Months:")
	if len(months) == 0 {
		fmt.Println("  No logs found.")
	} else {
		for _, m := range months {
			fmt.Printf("  - %s\n", m)
		}
	}
	fmt.Println()

	if targetMonthSummary != nil {
		fmt.Printf("Details for Month %s:\n", targetMonthSummary.Month)
		printTokenUsage(targetMonthSummary.TotalTokens)
		fmt.Println()

		fmt.Println("Workspaces:")
		if len(targetMonthSummary.Workspaces) == 0 {
			fmt.Println("  No workspace data for this month.")
		} else {
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "  Workspace\tPath\tTotal Tokens\tPrompt\tCompletion\tAIC\tRequests")
			fmt.Fprintln(w, "  ---------\t----\t------------\t------\t----------\t---\t--------")
			for _, wsSummary := range targetMonthSummary.Workspaces {
				fmt.Fprintf(w, "  %s\t%s\t%d\t%d\t%d\t%.2f\t%d\n",
					wsSummary.Workspace.Name,
					wsSummary.Workspace.Path,
					wsSummary.Tokens.Total,
					wsSummary.Tokens.Prompt,
					wsSummary.Tokens.Completion,
					wsSummary.Tokens.AIC,
					wsSummary.Tokens.Requests,
				)
			}
			w.Flush()
		}
	} else if len(months) > 0 {
		// If no month is selected, auto-select the latest month to be helpful
		latestMonth := months[0]
		summary, err := uc.GetSummaryForMonth(latestMonth)
		if err == nil {
			fmt.Printf("Details for Latest Month (%s):\n", latestMonth)
			printTokenUsage(summary.TotalTokens)
			fmt.Println()

			fmt.Println("Workspaces:")
			if len(summary.Workspaces) == 0 {
				fmt.Println("  No workspace data.")
			} else {
				w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
				fmt.Fprintln(w, "  Workspace\tPath\tTotal Tokens\tPrompt\tCompletion\tAIC\tRequests")
				fmt.Fprintln(w, "  ---------\t----\t------------\t------\t----------\t---\t--------")
				for _, wsSummary := range summary.Workspaces {
					fmt.Fprintf(w, "  %s\t%s\t%d\t%d\t%d\t%.2f\t%d\n",
						wsSummary.Workspace.Name,
						wsSummary.Workspace.Path,
						wsSummary.Tokens.Total,
						wsSummary.Tokens.Prompt,
						wsSummary.Tokens.Completion,
						wsSummary.Tokens.AIC,
						wsSummary.Tokens.Requests,
					)
				}
				w.Flush()
			}
			fmt.Println()
			fmt.Println("Hint: Use --month <YYYY-MM> to get details for other months.")
		}
	}
}

func printTokenUsage(t domain.TokenUsage) {
	fmt.Printf("  Total Tokens:      %d\n", t.Total)
	fmt.Printf("  Prompt Tokens:     %d\n", t.Prompt)
	fmt.Printf("  Completion Tokens: %d\n", t.Completion)
	fmt.Printf("  AI Credits (AIC):  %.2f\n", t.AIC)
	fmt.Printf("  AI Usage (AIU):    %.0f\n", t.AIU)
	fmt.Printf("  Total Requests:    %d\n", t.Requests)
}
