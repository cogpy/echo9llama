package orchestration

import (
	"context"
	"time"

	"github.com/ollama/ollama/api"
)

// Agent represents an orchestration agent that can coordinate multiple models and tasks
type Agent struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Models      []string               `json:"models"`
	Config      map[string]interface{} `json:"config"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// Task represents a task that can be executed by an orchestration agent
type Task struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Input       string                 `json:"input"`
	Output      string                 `json:"output,omitempty"`
	Status      string                 `json:"status"`
	AgentID     string                 `json:"agent_id"`
	ModelName   string                 `json:"model_name,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// TaskStatus constants
const (
	TaskStatusPending   = "pending"
	TaskStatusRunning   = "running"
	TaskStatusCompleted = "completed"
	TaskStatusFailed    = "failed"
)

// TaskType constants
const (
	TaskTypeGenerate = "generate"
	TaskTypeChat     = "chat"
	TaskTypeEmbed    = "embed"
	TaskTypeCustom   = "custom"
)

// OrchestrationRequest represents a request to orchestrate multiple tasks
type OrchestrationRequest struct {
	AgentID     string                 `json:"agent_id"`
	Tasks       []TaskRequest          `json:"tasks"`
	Sequential  bool                   `json:"sequential"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Stream      *bool                  `json:"stream,omitempty"`
	KeepAlive   *api.Duration          `json:"keep_alive,omitempty"`
}

// TaskRequest represents a single task within an orchestration request
type TaskRequest struct {
	Type       string                 `json:"type"`
	Input      string                 `json:"input"`
	ModelName  string                 `json:"model_name,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// OrchestrationResponse represents the response from an orchestration request
type OrchestrationResponse struct {
	ID        string `json:"id"`
	AgentID   string `json:"agent_id"`
	Status    string `json:"status"`
	Tasks     []Task `json:"tasks"`
	Results   []TaskResult `json:"results,omitempty"`
	Error     string `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// TaskResult represents the result of a completed task
type TaskResult struct {
	TaskID    string `json:"task_id"`
	Output    string `json:"output"`
	ModelUsed string `json:"model_used,omitempty"`
	Metrics   TaskMetrics `json:"metrics,omitempty"`
}

// TaskMetrics contains performance metrics for a completed task
type TaskMetrics struct {
	Duration     time.Duration `json:"duration"`
	TokensUsed   int           `json:"tokens_used,omitempty"`
	PromptTokens int           `json:"prompt_tokens,omitempty"`
	OutputTokens int           `json:"output_tokens,omitempty"`
}

// AgentManager interface defines methods for managing orchestration agents
type AgentManager interface {
	CreateAgent(ctx context.Context, agent *Agent) error
	GetAgent(ctx context.Context, id string) (*Agent, error)
	ListAgents(ctx context.Context) ([]*Agent, error)
	UpdateAgent(ctx context.Context, agent *Agent) error
	DeleteAgent(ctx context.Context, id string) error
}

// TaskExecutor interface defines methods for executing tasks
type TaskExecutor interface {
	ExecuteTask(ctx context.Context, task *Task, agent *Agent) (*TaskResult, error)
	ExecuteTasks(ctx context.Context, tasks []*Task, agent *Agent, sequential bool) ([]*TaskResult, error)
}

// Orchestrator interface combines agent management and task execution
type Orchestrator interface {
	AgentManager
	TaskExecutor
	OrchestrateTasks(ctx context.Context, req *OrchestrationRequest) (*OrchestrationResponse, error)
}