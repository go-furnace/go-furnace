package commands

import (
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/Skarlso/go_aws_mine/utils"
	"github.com/Yitsushi/go-commander"
)

// Init is an init command
type Init struct{}

// Execute initializes everything.
func (i *Init) Execute(opts *commander.CommandHelper) {
	usr, err := user.Current()
	utils.CheckError(err)

	err = os.Mkdir(filepath.Join(usr.HomeDir, ".config", "go_aws_mine"), os.ModePerm)
	utils.CheckError(err)

	// Copy files from cfg folder to ./config/go_aws_mine.
	dst, err := os.Create(filepath.Join(usr.HomeDir, ".config", "go_aws_mine", "ec2_conf.json"))
	utils.CheckError(err)

	if _, err = dst.WriteString(defaultEC2Config()); err != nil {
		utils.CheckError(err)
	}

	log.Println("Ec2 configuration created in home.")
}

// NewInit initializes configuration values.
func NewInit(appName string) *commander.CommandWrapper {
	return &commander.CommandWrapper{
		Handler: &Init{},
		Help: &commander.CommandDescriptor{
			Name:             "init",
			ShortDescription: "Initialize configuration values.",
			LongDescription: `Init initializes configurations in the users home/.config folder.
This command is OS agnostic and on Windows it will create a folder under user/.config/go_aws_mine.`,
			Arguments: "",
			Examples:  []string{},
		},
	}
}

func defaultEC2Config() string {
	return `{
    "dry_run": true,
    "image_id": "ami-ea26ce85",
    "key_name": "minecraft_keys",
    "min_count": 1,
    "max_count": 1,
    "instance_type": "t2.large",
    "monitoring": {
        "enabled": true
    }
}`
}

func defaultSGConfig() string {
	return `{
  "ip_permissions": [
    {
      "ip_protocol": "tcp",
      "from_port": 22,
      "to_port": 22,
      "ip_ranges": [{
        "cidr_ip": "0.0.0.0/0"
      }]
    },
    {
      "ip_protocol": "tcp",
      "from_port": 25565,
      "to_port": 25565,
      "ip_ranges": [{
        "cidr_ip": "0.0.0.0/0"
      }]
    }
  ]
}`
}

func defaultUserData() string {
	return `#!/bin/bash
yum update -y
yum install git -y
yum install libevent-devel -y
yum install ncurses-devel -y
yum install glibc-static -y
yum install java-1.8.0-openjdk -y
yum groupinstall "Development tools" -y
cd ~
wget https://github.com/downloads/libevent/libevent/libevent-2.0.21-stable.tar.gz
tar xzvf libevent-2.0.21-stable.tar.gz
cd libevent-2.0.21-stable
./configure && make
make install
cd /home/ec2-user
wget https://github.com/tmux/tmux/releases/download/2.2/tmux-2.2.tar.gz
tar xfvz tmux-2.2.tar.gz
cd tmux-2.2
./configure && make
cd /home/ec2-user
chown -R ec2-user:ec2-user tmux-2.2`
}
