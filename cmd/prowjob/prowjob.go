package prowjob

import (
	"github.com/spf13/cobra"
)

const (
	artifactDirEnv string = "ARTIFACT_DIR"
	githubTokenEnv string = "GITHUB_TOKEN" // #nosec G101
	prowJobIDEnv   string = "PROW_JOB_ID"

	artifactDirParamName     string = "artifact-dir"
	failIfUnhealthyParamName string = "fail-if-unhealthy"
	notifyOnPRParamName      string = "notify-on-pr"
	prowJobIDParamName       string = "prow-job-id"
)

var (
	artifactDir     string
	failIfUnhealthy bool
	notifyOnPR      bool
	prowJobID       string
)

// ProwjobCmd represents the prowjob command
var ProwjobCmd = &cobra.Command{
	Use:   "prowjob",
	Short: "Commands for processing Prow jobs",
}

func init() {
	ProwjobCmd.AddCommand(summaryReportCmd)
	ProwjobCmd.AddCommand(createReportCmd)
	ProwjobCmd.AddCommand(healthCheckCmd)

	createReportCmd.Flags().StringVar(&artifactDir, artifactDirParamName, "", "Path to the folder where to store produced files")
	healthCheckCmd.Flags().StringVar(&artifactDir, artifactDirParamName, "", "Path to the folder where to store produced files")
}
