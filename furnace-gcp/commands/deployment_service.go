package commands

import (
	"net/http"

	"golang.org/x/net/context"
	dm "google.golang.org/api/deploymentmanager/v2"
	"google.golang.org/api/option"
)

// DeploymentService defines a service which implement `Insert`. This method
// inserts a deployment into a GCP project.
type DeploymentService interface {
	Insert(project string, deployment *dm.Deployment) *dm.DeploymentsInsertCall
	Delete(project string, deployment string) *dm.DeploymentsDeleteCall
	Get(project string, deployment string) *dm.DeploymentsGetCall
	Update(project string, deployment string, deployment2 *dm.Deployment) *dm.DeploymentsUpdateCall
}

// DeploymentmanagerService defines a struct that we can use to mock GCP's
// deploymentmanager/v2 API.
type DeploymentmanagerService struct {
	*dm.Service
	Deployments DeploymentService
}

// NewDeploymentService will return a deployment manager service that
// can be used as a mock for the GCP deployment manager.
func NewDeploymentService(ctx context.Context, client *http.Client) DeploymentmanagerService {
	d, _ := dm.NewService(ctx, option.WithHTTPClient(client))

	return DeploymentmanagerService{
		Deployments: d.Deployments,
	}
}
