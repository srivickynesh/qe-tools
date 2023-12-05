package prowjob

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// summaryReportCmd returns the periodic-report command
var summaryReportCmd = &cobra.Command{
	Use:   "summary-report",
	Short: "Analyzes the build log from latest ci jobs and returns a short job summary",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		requiredEnvVars := []string{"prow_url"}

		for _, e := range requiredEnvVars {
			if viper.GetString(e) == "" {
				return fmt.Errorf("%+v env var not set", strings.ToUpper(e))
			}
		}
		return nil
	},
	RunE: run,
}

func removeANSIEscapeSequences(text string) string {
	regex := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	return regex.ReplaceAllString(text, "")
}

func fetchTextContent(url string) (string, error) {
	// #nosec G107
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error fetching the webpage: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading the webpage content: %w", err)
	}

	return string(bodyBytes), nil
}

func constructMessage(bodyString string) (string, bool) {
	cleanBody := removeANSIEscapeSequences(bodyString)

	if isJobFailed(cleanBody) {
		message := "Test Suite Summary:\n"
		message += extractTestResultsAndSummary(cleanBody)
		message += extractDuration(cleanBody)
		message += extractFailureSummary(cleanBody)
		return message, false
	}
	return "Job Succeeded", true
}

func isJobFailed(body string) bool {
	stateRegexp := regexp.MustCompile(`Reporting job state '(\w+)'`)
	stateMatches := stateRegexp.FindStringSubmatch(body)

	return len(stateMatches) == 2 && stateMatches[1] == "failed"
}

func extractFailureSummary(body string) string {
	failureMatches := regexp.MustCompile(`(?s)(Summarizing.*?Test Suite Failed)`).FindStringSubmatch(body)

	if failureMatches == nil && isJobFailed(removeANSIEscapeSequences(body)) {
		return ""
	}

	return formatFailures(failureMatches[1]) + "\n"
}

func extractTestResultsAndSummary(body string) string {
	pattern := `Ran (\d+) of (\d+) Specs in ([\d.]+) seconds\nFAIL! -- (\d+) Passed \| (\d+) Failed \| (\d+) Pending \| (\d+) Skipped`
	matches := regexp.MustCompile(pattern).FindStringSubmatch(body)

	if matches == nil {
		return "Infrastructure setup issues or failures unrelated to tests were found\n"
	}

	return fmt.Sprintf("Test Results: %s Passed | %s Failed | %s Pending | %s Skipped\nRan %s of %s Specs in %s seconds\n",
		matches[4], matches[5], matches[6], matches[7], matches[1], matches[2], matches[3])
}

func extractDuration(body string) string {
	matches := regexp.MustCompile(`Ran for ([\dhms]+)`).FindStringSubmatch(body)
	if matches == nil {
		return ""
	}
	return fmt.Sprintf("Total Duration: %s\n", matches[1])
}

func formatFailures(failures string) string {
	var formattedFailures strings.Builder
	formattedFailures.WriteString("Failures:\n")

	for _, line := range strings.Split(failures, "\n") {
		if strings.Contains(line, "[FAIL]") {
			formattedFailures.WriteString("- " + strings.TrimSpace(line) + "\n")
		}
	}

	if formattedFailures.String() == "Failures:\n" {
		return "No specific failures captured in the report.\n"
	}

	return formattedFailures.String()
}

func run(cmd *cobra.Command, args []string) error {
	// Required build.log PATH for latest build
	prowURL := os.Getenv("PROW_URL")
	bodyString, err := fetchTextContent(prowURL)
	if err != nil {
		return err
	}

	message, _ := constructMessage(bodyString)
	fmt.Println(message)
	return nil
}
