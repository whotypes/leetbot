package data

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
)

type FirestoreStorage struct {
    client *firestore.Client
}

func (fs *FirestoreStorage) Close() error {
    return fs.client.Close()
}

func NewFirestoreClient(ctx context.Context, projectID string, databaseID string) (*FirestoreStorage, error) {
    if projectID == "" {
        return nil, fmt.Errorf("projectID is required")
    }
    if databaseID == "" {
        return nil, fmt.Errorf("databaseID is required - we do not allow connections to the default database")
    }

    // Using Application Default Credentials (ADC) - no explicit credentials needed
    client, err := firestore.NewClientWithDatabase(ctx, projectID, databaseID)
    if err != nil {
        return nil, err
    }
    return &FirestoreStorage{client: client}, nil
}

// AddProcess implements the Storage interface by creating a new process
// document and returning its generated ID.
func (fs *FirestoreStorage) AddProcess(ctx context.Context, process Process) (string, error) {
    if err := GetValidator().Struct(process); err != nil {
        return "", fmt.Errorf("invalid process: %v", err)
    }
    ref, _, err := fs.client.Collection("processes").Add(ctx, process)
    if err != nil {
        return "", err
    }
    return ref.ID, nil
}

func (f *FirestoreStorage) GetProcessesByCompany(ctx context.Context, company string) ([]Process, error) {
    docs, err := f.client.Collection("processes").Where("company", "==", company).Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get processes by company: %v", err)
	}

	processes := make([]Process, 0)
	for _, doc := range docs {
		var process Process
		if err := doc.DataTo(&process); err != nil {
			return nil, fmt.Errorf("failed to convert document to process: %v", err)
		}
		processes = append(processes, process)
	}
	return processes, nil
}

func (f *FirestoreStorage) GetProcessesByCompanyAndStage(ctx context.Context, company string, stage string) ([]Process, error) {
	docs, err := f.client.Collection("processes").
		Where("company", "==", company).
		Where("stage", "==", stage).
		Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get processes by company and stage: %v", err)
	}

	processes := make([]Process, 0)
	for _, doc := range docs {
		var process Process
		if err := doc.DataTo(&process); err != nil {
			return nil, fmt.Errorf("failed to convert document to process: %v", err)
		}
		processes = append(processes, process)
	}
	return processes, nil
}
