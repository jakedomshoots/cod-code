package cli

func reviewQueueRequestFromOptions(opts options) reviewQueueRequest {
	return reviewQueueRequest{
		Query: historyQuery{
			workspaceDir: opts.workspaceDir,
			task:         opts.historyTask,
			limit:        opts.historyLimit,
			since:        opts.historySince,
			until:        opts.historyUntil,
		},
		Format:         opts.reportFormat,
		IncludeDetails: opts.reviewDetails,
	}
}
