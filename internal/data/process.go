package data

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func InitValidator() {
	validate = validator.New()
	if err := validate.RegisterValidation("process_stage", validateProcessStage); err != nil {
		panic(fmt.Sprintf("Failed to register validation: %v", err))
	}
}

func GetValidator() *validator.Validate {
	if validate == nil {
		InitValidator()
	}
	return validate
}

func validateProcessStage(fl validator.FieldLevel) bool {
    stage := fl.Field().String()
    s := strings.ToLower(stage)
    switch s {
    case "apply", "reject", "oa", "phone", "onsite", "offer":
        return true
    }
    return false
}

// the current stage of a candidates process
type ProcessStage string

const (
	ProcessStageApply  ProcessStage = "Apply"
	ProcessStageReject ProcessStage = "Reject"
	ProcessStageOA     ProcessStage = "OA"
	ProcessStagePhone  ProcessStage = "Phone"
	ProcessStageOnsite ProcessStage = "Onsite"
	ProcessStageOffer  ProcessStage = "Offer"
)

// GetAvailableStages returns all available process stages
func GetAvailableStages() []ProcessStage {
	return []ProcessStage{
		ProcessStageApply,
		ProcessStageReject,
		ProcessStageOA,
		ProcessStagePhone,
		ProcessStageOnsite,
		ProcessStageOffer,
	}
}

// track one company's interview process for a candidate
type Process struct {
	ID        string    `firestore:"id"`
	Company   string    `firestore:"company"`
	Stage     string    `firestore:"stage"`
	CreatedAt time.Time `firestore:"created_at"`
	UpdatedAt time.Time `firestore:"updated_at"`
}
