package commands

import "github.com/Yitsushi/go-commander"

// Deploy command.
type Deploy struct {
}

// Execute defines what this command does.
func (c *Deploy) Execute(opts *commander.CommandHelper) {
	// deployVersion := opts.Arg(0)

	// sess := session.New(&aws.Config{Region: aws.String("eu-central-1")})
	// cfClient := cloudformation.New(sess, nil)
}

// NewDeploy Creates a new Deploy command.
func NewDeploy(appName string) *commander.CommandWrapper {
	return &commander.CommandWrapper{
		Handler: &Deploy{},
		Help: &commander.CommandDescriptor{
			Name:             "Deploy",
			ShortDescription: "Deploy to stack",
			LongDescription:  `Deploy a version of the application to a stack`,
			Arguments:        "name",
			Examples:         []string{"deploy", "deploy version"},
		},
	}
}
