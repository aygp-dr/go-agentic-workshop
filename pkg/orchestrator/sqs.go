package orchestrator

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// TaskMessage represents a task in SQS
type TaskMessage struct {
	WorkflowID string                 `json:"workflow_id"`
	TaskType   string                 `json:"task_type"`
	Payload    map[string]interface{} `json:"payload"`
	Priority   int                    `json:"priority"`
}

// SQSOrchestrator handles async task execution
type SQSOrchestrator struct {
	client   *sqs.Client
	queueURL string
}

// EnqueueTask adds a task to the queue
func (o *SQSOrchestrator) EnqueueTask(ctx context.Context, task *TaskMessage) error {
	body, err := json.Marshal(task)
	if err != nil {
		return err
	}

	_, err = o.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    &o.queueURL,
		MessageBody: aws.String(string(body)),
	})

	return err
}
