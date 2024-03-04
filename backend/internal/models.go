package internal

type (
	User struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	Response struct {
		Message string `json:"message"`
	}

	CaptionTask struct {
		Id                int    `json:"id"`
		Prompt            string `json:"prompt"`
		Status            string `json:"status"`
		TotalSubtasks     int    `json:"total_subtasks"`
		CompletedSubtasks int    `json:"completed_subtasks"`
		PercentComplete   int    `json:"percent_complete"`
	}

	CaptionTaskEvent struct {
		Id     int    `json:"id"`
		Prompt string `json:"prompt"`
		Url    string `json:"url"`
	}
)
