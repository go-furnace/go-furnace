package commands

import "github.com/Yitsushi/go-commander"

// Push command.
type Push struct {
}

// Execute defines what this command does.
func (c *Push) Execute(opts *commander.CommandHelper) {
	// deployVersion := opts.Arg(0)

	// sess := session.New(&aws.Config{Region: aws.String("eu-central-1")})
	// cfClient := cloudformation.New(sess, nil)
}

// NewPush Creates a new Push command.
func NewPush(appName string) *commander.CommandWrapper {
	return &commander.CommandWrapper{
		Handler: &Push{},
		Help: &commander.CommandDescriptor{
			Name:             "Push",
			ShortDescription: "Push to stack",
			LongDescription:  `Push a version of the application to a stack`,
			Arguments:        "name",
			Examples:         []string{"push", "push version"},
		},
	}
}
