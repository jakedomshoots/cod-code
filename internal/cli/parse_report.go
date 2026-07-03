package cli

import (
	"fmt"
	"strconv"
)

func parseReportFlag(args []string, index int, opts *options) (bool, int, error) {
	switch args[index] {
	case "--review-queue":
		opts.showReviewQueue = true
		return true, index, nil
	case "--review-details":
		opts.reviewDetails = true
		return true, index, nil
	case "--history":
		opts.showHistory = true
		return true, index, nil
	case "--provider-health":
		opts.showProviderHealth = true
		return true, index, nil
	case "--doctor-provider":
		value, err := parseNextValue(args, index, "--doctor-provider requires a provider name")
		if err != nil {
			return true, index, err
		}
		opts.doctorProviderName = value
		return true, index + 1, nil
	case "--provider":
		value, err := parseNextValue(args, index, "--provider requires a name")
		if err != nil {
			return true, index, err
		}
		opts.providerFilter = value
		return true, index + 1, nil
	case "--recommendation":
		value, err := parseNextValue(args, index, "--recommendation requires a label")
		if err != nil {
			return true, index, err
		}
		recommendation, err := parseProviderRecommendation(value)
		if err != nil {
			return true, index, err
		}
		opts.recommendationFilter = recommendation
		return true, index + 1, nil
	case "--top-providers":
		value, err := parsePositiveIntFlag(args, index, "--top-providers")
		if err != nil {
			return true, index, err
		}
		opts.topProviders = value
		return true, index + 1, nil
	case "--format":
		value, err := parseNextValue(args, index, "--format requires json, text, or events")
		if err != nil {
			return true, index, err
		}
		format, err := parseReportFormat(value)
		if err != nil {
			return true, index, err
		}
		opts.reportFormat = format
		opts.reportFormatSet = true
		return true, index + 1, nil
	case "--verdict":
		value, err := parseNextValue(args, index, "--verdict requires a value")
		if err != nil {
			return true, index, err
		}
		opts.historyVerdict = value
		return true, index + 1, nil
	case "--task":
		value, err := parseNextValue(args, index, "--task requires text")
		if err != nil {
			return true, index, err
		}
		opts.historyTask = value
		return true, index + 1, nil
	case "--summary-only":
		opts.historySummaryOnly = true
		return true, index, nil
	case "--limit":
		value, err := parsePositiveIntFlag(args, index, "--limit")
		if err != nil {
			return true, index, err
		}
		opts.historyLimit = value
		return true, index + 1, nil
	case "--since":
		value, err := parseNextValue(args, index, "--since requires a timestamp")
		if err != nil {
			return true, index, err
		}
		opts.historySince = value
		return true, index + 1, nil
	case "--until":
		value, err := parseNextValue(args, index, "--until requires a timestamp")
		if err != nil {
			return true, index, err
		}
		opts.historyUntil = value
		return true, index + 1, nil
	case "--job":
		value, err := parseNextValue(args, index, "--job requires an id")
		if err != nil {
			return true, index, err
		}
		opts.jobID = value
		return true, index + 1, nil
	case "--job-context":
		value, err := parseNextValue(args, index, "--job-context requires an id")
		if err != nil {
			return true, index, err
		}
		opts.jobContextID = value
		return true, index + 1, nil
	case "--context-trace":
		value, err := parseNextValue(args, index, "--context-trace requires an id")
		if err != nil {
			return true, index, err
		}
		opts.contextTraceID = value
		return true, index + 1, nil
	case "--with-job-context":
		value, err := parseNextValue(args, index, "--with-job-context requires an id")
		if err != nil {
			return true, index, err
		}
		opts.priorJobContextID = value
		return true, index + 1, nil
	case "--job-report":
		value, err := parseNextValue(args, index, "--job-report requires an id")
		if err != nil {
			return true, index, err
		}
		opts.jobReportID = value
		return true, index + 1, nil
	case "--explain-failure":
		value, err := parseNextValue(args, index, "--explain-failure requires an id")
		if err != nil {
			return true, index, err
		}
		opts.explainFailureJobID = value
		return true, index + 1, nil
	case "--job-events":
		value, err := parseNextValue(args, index, "--job-events requires an id")
		if err != nil {
			return true, index, err
		}
		opts.jobEventsID = value
		return true, index + 1, nil
	case "--judge-job":
		value, err := parseNextValue(args, index, "--judge-job requires an id")
		if err != nil {
			return true, index, err
		}
		opts.judgeJobID = value
		return true, index + 1, nil
	case "--human-verdict":
		value, err := parseNextValue(args, index, "--human-verdict requires accept or reject")
		if err != nil {
			return true, index, err
		}
		opts.humanVerdict = value
		return true, index + 1, nil
	case "--judgment-note":
		value, err := parseNextValue(args, index, "--judgment-note requires text")
		if err != nil {
			return true, index, err
		}
		opts.judgmentNote = value
		return true, index + 1, nil
	case "--rerun":
		value, err := parseNextValue(args, index, "--rerun requires an id")
		if err != nil {
			return true, index, err
		}
		opts.rerunJobID = value
		return true, index + 1, nil
	case "--continue-job":
		value, err := parseNextValue(args, index, "--continue-job requires an id")
		if err != nil {
			return true, index, err
		}
		opts.continueJobID = value
		return true, index + 1, nil
	case "--resume":
		value, err := parseNextValue(args, index, "--resume requires an id")
		if err != nil {
			return true, index, err
		}
		opts.resumeJobID = value
		return true, index + 1, nil
	case "--answer":
		value, err := parseNextValue(args, index, "--answer requires text")
		if err != nil {
			return true, index, err
		}
		opts.resumeAnswers = append(opts.resumeAnswers, value)
		return true, index + 1, nil
	default:
		return false, index, nil
	}
}

func parsePositiveIntFlag(args []string, index int, flag string) (int, error) {
	raw, err := parseNextValue(args, index, flag+" requires a number")
	if err != nil {
		return 0, err
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return 0, fmt.Errorf("%s must be a positive integer", flag)
	}
	return value, nil
}
