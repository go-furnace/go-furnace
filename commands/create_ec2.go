package commands

import (
  "log"
)

// Yout Command
type CreateEC2 struct {
}

func (c *CreateEC2) Execute() {
  log.Println("Running CreateEC2 command")
}

func (c *CreateEC2) ArgumentDescription() string {
  return "[create-ec2]"
}

func (c *CreateEC2) Description() string {
  return "Create an EC2 instance to run the server on."
}

func (c *CreateEC2) Help() string {
  return "go_aws_mine create-ec2"
}

func (c *CreateEC2) Examples() []string {
  return []string{"", "test"}
}
