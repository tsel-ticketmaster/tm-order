package gctasks

import (
	"context"
	"fmt"
	"math"
	"time"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"cloud.google.com/go/cloudtasks/apiv2/cloudtaskspb"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Client interface {
	CreateQueue(id string) (err error)
	CreateTask(queueID string, request Request) (err error)
	DeferCreateTaskInDuration(queueID string, request Request, duration time.Duration) (err error)
	DeferCreateTaskInTime(queueID string, request Request, schedule time.Time) (err error)
	Close() error
}

const (
	locationID = "asia-southeast2"
)

type tasksClientImpl struct {
	projectID string
	logger    *logrus.Logger
	client    *cloudtasks.Client
}
type Request struct {
	URL    string
	Method cloudtaskspb.HttpMethod
	Header map[string]string
	Body   []byte
}

func NewGCTasks(logger *logrus.Logger, projectID string, credsJson []byte) Client {
	c, err := cloudtasks.NewClient(context.Background(), option.WithCredentialsJSON(credsJson))
	if err != nil {
		logger.WithField("object", "gctasks").Error(err)
		return nil
	}
	if err != nil {
		logger.WithField("object", "gctasks").Error(err)
		return nil
	}

	return &tasksClientImpl{
		logger:    logger,
		client:    c,
		projectID: projectID,
	}
}

func (tc *tasksClientImpl) Close() error {
	return tc.client.Close()
}

func (tc *tasksClientImpl) CreateQueue(id string) (err error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", tc.projectID, locationID)
	queuePath := fmt.Sprintf("%s/queues/%s", parent, id)

	// Define the queue configuration.
	queue := &cloudtaskspb.Queue{
		Name: queuePath,
	}

	// Create the queue.
	_, err = tc.client.CreateQueue(context.Background(), &cloudtaskspb.CreateQueueRequest{
		Parent: parent,
		Queue:  queue,
	})

	if err != nil {
		tc.logger.WithField("object", "gctasks").Error(err)
		return err
	}

	return nil
}

func (tc *tasksClientImpl) CreateTask(queueID string, request Request) (err error) {
	queuePath := fmt.Sprintf("projects/%s/locations/%s/queues/%s", tc.projectID, locationID, queueID)

	// Define the task to add to the queue.
	task := &cloudtaskspb.Task{
		MessageType: &cloudtaskspb.Task_HttpRequest{
			HttpRequest: &cloudtaskspb.HttpRequest{
				Url:        request.URL,
				HttpMethod: request.Method,
				Headers:    request.Header,
				Body:       request.Body,
			},
		},
	}

	// Create a task request.
	createTaskRequest := &cloudtaskspb.CreateTaskRequest{
		Parent: queuePath,
		Task:   task,
	}

	// Enqueue the task.
	_, err = tc.client.CreateTask(context.Background(), createTaskRequest)
	if err != nil {
		tc.logger.WithField("object", "gctasks").Error(err)
		return err
	}

	return nil
}

func (tc *tasksClientImpl) DeferCreateTaskInDuration(queueID string, request Request, duration time.Duration) (err error) {
	queuePath := fmt.Sprintf("projects/%s/locations/%s/queues/%s", tc.projectID, locationID, queueID)

	// Define the task to add to the queue.
	task := &cloudtaskspb.Task{
		MessageType: &cloudtaskspb.Task_HttpRequest{
			HttpRequest: &cloudtaskspb.HttpRequest{
				Url:        request.URL,
				HttpMethod: request.Method,
				Headers:    request.Header,
				Body:       request.Body,
			},
		},

		ScheduleTime: &timestamppb.Timestamp{
			Seconds: time.Now().Add(duration).Unix(),
		},
	}

	// Create a task request.
	createTaskRequest := &cloudtaskspb.CreateTaskRequest{
		Parent: queuePath,
		Task:   task,
	}

	// Enqueue the task.
	_, err = tc.client.CreateTask(context.Background(), createTaskRequest)
	if err != nil {
		tc.logger.WithField("object", "gctasks").Error(err)
		return err
	}

	return nil
}

func (tc *tasksClientImpl) DeferCreateTaskInTime(queueID string, request Request, schedule time.Time) (err error) {
	queuePath := fmt.Sprintf("projects/%s/locations/%s/queues/%s", tc.projectID, locationID, queueID)

	now := time.Now()
	difference := schedule.Sub(now).Seconds()
	duration := math.Floor(difference)

	// Define the task to add to the queue.
	task := &cloudtaskspb.Task{
		MessageType: &cloudtaskspb.Task_HttpRequest{
			HttpRequest: &cloudtaskspb.HttpRequest{
				Url:        request.URL,
				HttpMethod: request.Method,
				Headers:    request.Header,
				Body:       request.Body,
			},
		},

		ScheduleTime: &timestamppb.Timestamp{
			Seconds: time.Now().Add(time.Duration(duration) * time.Second).Unix(),
		},
	}

	// Create a task request.
	createTaskRequest := &cloudtaskspb.CreateTaskRequest{
		Parent: queuePath,
		Task:   task,
	}

	// Enqueue the task.
	_, err = tc.client.CreateTask(context.Background(), createTaskRequest)
	if err != nil {
		tc.logger.WithFields(logrus.Fields{
			"object":    "gctasks",
			"queueId":   queueID,
			"queuePath": queuePath,
		}).Error(err)
		return err
	}

	return nil
}
