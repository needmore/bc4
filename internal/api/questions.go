package api

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// Questionnaire represents a Basecamp automated check-ins container
type Questionnaire struct {
	ID               int64     `json:"id"`
	Status           string    `json:"status"`
	VisibleToClients bool      `json:"visible_to_clients"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Title            string    `json:"title"`
	InheritsStatus   bool      `json:"inherits_status"`
	Type             string    `json:"type"`
	URL              string    `json:"url"`
	AppURL           string    `json:"app_url"`
	BookmarkURL      string    `json:"bookmark_url"`
	Position         int       `json:"position"`
	Bucket           *Bucket   `json:"bucket,omitempty"`
	Creator          *Person   `json:"creator,omitempty"`
	QuestionsCount   int       `json:"questions_count"`
	QuestionsURL     string    `json:"questions_url"`
}

// QuestionSchedule represents the schedule for a check-in question
type QuestionSchedule struct {
	Frequency string `json:"frequency"`      // "every_day", "every_week", "every_other_week", "every_four_weeks"
	Days      []int  `json:"days,omitempty"` // 0=Sunday through 6=Saturday
	Hour      int    `json:"hour"`           // 0-23
	Minute    int    `json:"minute"`         // 0-59
	StartDate string `json:"start_date,omitempty"`
}

// Question represents a Basecamp automated check-in question
type Question struct {
	ID                   int64                         `json:"id"`
	Status               string                        `json:"status"`
	VisibleToClients     bool                          `json:"visible_to_clients"`
	CreatedAt            time.Time                     `json:"created_at"`
	UpdatedAt            time.Time                     `json:"updated_at"`
	Title                string                        `json:"title"`
	InheritsStatus       bool                          `json:"inherits_status"`
	Type                 string                        `json:"type"`
	URL                  string                        `json:"url"`
	AppURL               string                        `json:"app_url"`
	BookmarkURL          string                        `json:"bookmark_url"`
	SubscriptionURL      string                        `json:"subscription_url"`
	Parent               *Parent                       `json:"parent,omitempty"`
	Bucket               *Bucket                       `json:"bucket,omitempty"`
	Creator              *Person                       `json:"creator,omitempty"`
	Paused               bool                          `json:"paused"`
	Schedule             *QuestionSchedule             `json:"schedule,omitempty"`
	AnswersCount         int                           `json:"answers_count"`
	AnswersURL           string                        `json:"answers_url"`
	NotificationSettings *QuestionNotificationSettings `json:"notification_settings,omitempty"`
}

// QuestionAnswer represents an answer to a check-in question
type QuestionAnswer struct {
	ID               int64     `json:"id"`
	Status           string    `json:"status"`
	VisibleToClients bool      `json:"visible_to_clients"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Title            string    `json:"title"`
	InheritsStatus   bool      `json:"inherits_status"`
	Type             string    `json:"type"`
	URL              string    `json:"url"`
	AppURL           string    `json:"app_url"`
	BookmarkURL      string    `json:"bookmark_url"`
	SubscriptionURL  string    `json:"subscription_url"`
	CommentsCount    int       `json:"comments_count"`
	CommentsURL      string    `json:"comments_url"`
	Parent           *Parent   `json:"parent,omitempty"`
	Bucket           *Bucket   `json:"bucket,omitempty"`
	Creator          *Person   `json:"creator,omitempty"`
	Content          string    `json:"content"`
	GroupOn          string    `json:"group_on"` // Date the answer is grouped under (YYYY-MM-DD)
}

// QuestionReminder represents a pending check-in reminder for the current user
type QuestionReminder struct {
	ID         int64     `json:"id"`
	RemindAt   time.Time `json:"remind_at"`
	GroupOn    string    `json:"group_on"`
	Question   *Question `json:"question,omitempty"`
	QuestionID int64     `json:"question_id"`
	Bucket     *Bucket   `json:"bucket,omitempty"`
}

// QuestionNotificationSettings represents notification settings for a question
type QuestionNotificationSettings struct {
	Responding bool `json:"responding"` // Whether to notify when someone responds
	Subscribed bool `json:"subscribed"` // Whether to receive question notifications
}

// QuestionCreateRequest represents the payload for creating a new check-in question
type QuestionCreateRequest struct {
	Title    string `json:"title"`
	Schedule string `json:"schedule"`         // "every_day", "every_week", "every_other_week", "every_four_weeks"
	Days     []int  `json:"days,omitempty"`   // 0=Sunday through 6=Saturday
	Hour     int    `json:"hour,omitempty"`   // 0-23
	Minute   int    `json:"minute,omitempty"` // 0-59
}

// QuestionUpdateRequest represents the payload for updating a check-in question
type QuestionUpdateRequest struct {
	Title    *string `json:"title,omitempty"`
	Schedule *string `json:"schedule,omitempty"`
	Days     []int   `json:"days,omitempty"`
	Hour     *int    `json:"hour,omitempty"`
	Minute   *int    `json:"minute,omitempty"`
}

// AnswerCreateRequest represents the payload for creating an answer
type AnswerCreateRequest struct {
	Content string `json:"content"`
}

// AnswerUpdateRequest represents the payload for updating an answer
type AnswerUpdateRequest struct {
	Content string `json:"content"`
}

// NotificationSettingsUpdateRequest represents the payload for updating notification settings
type NotificationSettingsUpdateRequest struct {
	Responding *bool `json:"responding,omitempty"`
	Subscribed *bool `json:"subscribed,omitempty"`
}

// AnswerListOptions represents options for listing answers
type AnswerListOptions struct {
	Date      string // Filter by date (YYYY-MM-DD)
	CreatorID int64  // Filter by creator person ID
}

// GetProjectQuestionnaire fetches the questionnaire (check-ins) for a project from its dock
func (c *Client) GetProjectQuestionnaire(ctx context.Context, projectID string) (*Questionnaire, error) {
	// First get the project to find its questionnaire
	project, err := c.GetProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Get project tools/features
	path := fmt.Sprintf("/projects/%d.json", project.ID)

	var projectData struct {
		Dock []struct {
			ID      int64  `json:"id"`
			Title   string `json:"title"`
			Name    string `json:"name"`
			URL     string `json:"url"`
			Enabled bool   `json:"enabled"`
		} `json:"dock"`
	}

	if err := c.Get(path, &projectData); err != nil {
		return nil, fmt.Errorf("failed to fetch project tools: %w", err)
	}

	// Find the questionnaire in the dock
	for _, tool := range projectData.Dock {
		if tool.Name == "questionnaire" {
			return &Questionnaire{
				ID:           tool.ID,
				Title:        tool.Title,
				QuestionsURL: tool.URL,
			}, nil
		}
	}

	return nil, fmt.Errorf("questionnaire (check-ins) not found for project")
}

// ListQuestions fetches all questions in a questionnaire
func (c *Client) ListQuestions(ctx context.Context, projectID string, questionnaireID int64) ([]Question, error) {
	var questions []Question
	path := fmt.Sprintf("/buckets/%s/questionnaires/%d/questions.json", projectID, questionnaireID)

	pr := NewPaginatedRequest(c)
	if err := pr.GetAll(path, &questions); err != nil {
		return nil, fmt.Errorf("failed to fetch questions: %w", err)
	}

	return questions, nil
}

// GetQuestion fetches a specific question by ID
func (c *Client) GetQuestion(ctx context.Context, projectID string, questionID int64) (*Question, error) {
	var question Question
	path := fmt.Sprintf("/buckets/%s/questions/%d.json", projectID, questionID)

	if err := c.Get(path, &question); err != nil {
		return nil, fmt.Errorf("failed to fetch question: %w", err)
	}

	return &question, nil
}

// CreateQuestion creates a new check-in question
func (c *Client) CreateQuestion(ctx context.Context, projectID string, questionnaireID int64, req QuestionCreateRequest) (*Question, error) {
	var question Question
	path := fmt.Sprintf("/buckets/%s/questionnaires/%d/questions.json", projectID, questionnaireID)

	if err := c.Post(path, req, &question); err != nil {
		return nil, fmt.Errorf("failed to create question: %w", err)
	}

	return &question, nil
}

// UpdateQuestion updates an existing check-in question
func (c *Client) UpdateQuestion(ctx context.Context, projectID string, questionID int64, req QuestionUpdateRequest) (*Question, error) {
	var question Question
	path := fmt.Sprintf("/buckets/%s/questions/%d.json", projectID, questionID)

	if err := c.Put(path, req, &question); err != nil {
		return nil, fmt.Errorf("failed to update question: %w", err)
	}

	return &question, nil
}

// PauseQuestion pauses a check-in question
func (c *Client) PauseQuestion(ctx context.Context, projectID string, questionID int64) error {
	path := fmt.Sprintf("/buckets/%s/questions/%d/pause.json", projectID, questionID)

	if err := c.Post(path, nil, nil); err != nil {
		return fmt.Errorf("failed to pause question: %w", err)
	}

	return nil
}

// ResumeQuestion resumes a paused check-in question
func (c *Client) ResumeQuestion(ctx context.Context, projectID string, questionID int64) error {
	path := fmt.Sprintf("/buckets/%s/questions/%d/pause.json", projectID, questionID)

	if err := c.Delete(path); err != nil {
		return fmt.Errorf("failed to resume question: %w", err)
	}

	return nil
}

// UpdateNotificationSettings updates notification settings for a question
func (c *Client) UpdateNotificationSettings(ctx context.Context, projectID string, questionID int64, req NotificationSettingsUpdateRequest) (*QuestionNotificationSettings, error) {
	var settings QuestionNotificationSettings
	path := fmt.Sprintf("/buckets/%s/questions/%d/notification_settings.json", projectID, questionID)

	if err := c.Put(path, req, &settings); err != nil {
		return nil, fmt.Errorf("failed to update notification settings: %w", err)
	}

	return &settings, nil
}

// ListAnswers fetches answers for a question with optional filtering
func (c *Client) ListAnswers(ctx context.Context, projectID string, questionID int64, opts *AnswerListOptions) ([]QuestionAnswer, error) {
	var answers []QuestionAnswer

	// Build query parameters
	params := url.Values{}
	if opts != nil {
		if opts.Date != "" {
			params.Set("date", opts.Date)
		}
		if opts.CreatorID > 0 {
			params.Set("creator_id", fmt.Sprintf("%d", opts.CreatorID))
		}
	}

	path := fmt.Sprintf("/buckets/%s/questions/%d/answers.json", projectID, questionID)
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	pr := NewPaginatedRequest(c)
	if err := pr.GetAll(path, &answers); err != nil {
		return nil, fmt.Errorf("failed to fetch answers: %w", err)
	}

	return answers, nil
}

// ListAnswerers fetches all people who have answered a question
func (c *Client) ListAnswerers(ctx context.Context, projectID string, questionID int64) ([]Person, error) {
	var answerers []Person
	path := fmt.Sprintf("/buckets/%s/questions/%d/answers/by.json", projectID, questionID)

	pr := NewPaginatedRequest(c)
	if err := pr.GetAll(path, &answerers); err != nil {
		return nil, fmt.Errorf("failed to fetch answerers: %w", err)
	}

	return answerers, nil
}

// GetAnswersByPerson fetches all answers from a specific person for a question
func (c *Client) GetAnswersByPerson(ctx context.Context, projectID string, questionID int64, personID int64) ([]QuestionAnswer, error) {
	var answers []QuestionAnswer
	path := fmt.Sprintf("/buckets/%s/questions/%d/answers/by/%d.json", projectID, questionID, personID)

	pr := NewPaginatedRequest(c)
	if err := pr.GetAll(path, &answers); err != nil {
		return nil, fmt.Errorf("failed to fetch answers by person: %w", err)
	}

	return answers, nil
}

// GetAnswer fetches a specific answer by ID
func (c *Client) GetAnswer(ctx context.Context, projectID string, answerID int64) (*QuestionAnswer, error) {
	var answer QuestionAnswer
	path := fmt.Sprintf("/buckets/%s/question_answers/%d.json", projectID, answerID)

	if err := c.Get(path, &answer); err != nil {
		return nil, fmt.Errorf("failed to fetch answer: %w", err)
	}

	return &answer, nil
}

// CreateAnswer creates a new answer for a question
func (c *Client) CreateAnswer(ctx context.Context, projectID string, questionID int64, req AnswerCreateRequest) (*QuestionAnswer, error) {
	var answer QuestionAnswer
	path := fmt.Sprintf("/buckets/%s/questions/%d/answers.json", projectID, questionID)

	if err := c.Post(path, req, &answer); err != nil {
		return nil, fmt.Errorf("failed to create answer: %w", err)
	}

	return &answer, nil
}

// UpdateAnswer updates an existing answer
func (c *Client) UpdateAnswer(ctx context.Context, projectID string, answerID int64, req AnswerUpdateRequest) (*QuestionAnswer, error) {
	var answer QuestionAnswer
	path := fmt.Sprintf("/buckets/%s/question_answers/%d.json", projectID, answerID)

	if err := c.Put(path, req, &answer); err != nil {
		return nil, fmt.Errorf("failed to update answer: %w", err)
	}

	return &answer, nil
}

// ListMyReminders fetches all pending check-in reminders for the current user
func (c *Client) ListMyReminders(ctx context.Context) ([]QuestionReminder, error) {
	var reminders []QuestionReminder
	path := "/my/question_reminders.json"

	pr := NewPaginatedRequest(c)
	if err := pr.GetAll(path, &reminders); err != nil {
		return nil, fmt.Errorf("failed to fetch reminders: %w", err)
	}

	return reminders, nil
}
