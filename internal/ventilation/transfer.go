package ventilation

import (
	"context"
	"fmt"
)

type TransferRequest struct {
	GroupID  string `json:"group_id"`
	TargetID string `json:"target_id"`
	Mode     string `json:"mode"`
}

func (r TransferRequest) Validate() error {
	if r.GroupID == "" {
		return fmt.Errorf("fan group is required")
	}
	if r.Mode != "transfer" && r.Mode != "reverse" {
		return fmt.Errorf("mode must be transfer or reverse")
	}
	if r.Mode == "transfer" && r.TargetID == "" {
		return fmt.Errorf("transfer target is required")
	}
	return nil
}

func Execute(ctx context.Context, service *Service, request TransferRequest) (string, error) {
	if err := request.Validate(); err != nil {
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if request.Mode == "reverse" {
		owner, err := service.Reverse(request.GroupID)
		if err != nil {
			return "", err
		}
		return owner.OperationID, nil
	}
	owner, err := service.Transfer(request.GroupID, request.TargetID)
	if err != nil {
		return "", err
	}
	return owner.OperationID, nil
}
