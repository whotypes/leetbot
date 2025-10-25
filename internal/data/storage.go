package data

import "context"

type Storage interface {
	AddProcess(ctx context.Context, p Process) (string, error)
	GetProcessesByCompany(ctx context.Context, company string) ([]Process, error)
	GetProcessesByCompanyAndStage(ctx context.Context, company string, stage string) ([]Process, error)
}
