package commands

import dm "google.golang.org/api/deploymentmanager/v2"

// MockDeploymentService is a mock for deployment service.
type MockDeploymentService struct {
	insert *dm.DeploymentsInsertCall
	delete *dm.DeploymentsDeleteCall
	get    *dm.DeploymentsGetCall
	update *dm.DeploymentsUpdateCall
}

// Insert inserts a deployment into a given project.
func (m *MockDeploymentService) Insert(project string, deployment *dm.Deployment) *dm.DeploymentsInsertCall {
	return m.insert
}

// Delete deletes a deployment from a given project.
func (m *MockDeploymentService) Delete(project string, deployment string) *dm.DeploymentsDeleteCall {
	return m.delete
}

// Get retrieves a deployment from a given project.
func (m *MockDeploymentService) Get(project string, deployment string) *dm.DeploymentsGetCall {
	return m.get
}

// Update updates a deployment for a given project.
func (m *MockDeploymentService) Update(project string, deployment string, deployment2 *dm.Deployment) *dm.DeploymentsUpdateCall {
	return m.update
}
