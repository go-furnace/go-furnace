package commands

import (
	"log"
	"os"
	"os/user"
	"path/filepath"
	"sync"

	"github.com/Skarlso/go-furnace/config"
	"github.com/Skarlso/go-furnace/godb"
	"github.com/Skarlso/go-furnace/utils"
	"github.com/Yitsushi/go-commander"
)

var configPath string

// Init is an init command
type Init struct{}

// Execute initializes everything.
func (i *Init) Execute(opts *commander.CommandHelper) {
	configPath = config.Path()
	usr, err := user.Current()
	utils.CheckError(err)

	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			mkdirErr := os.Mkdir(configPath, os.ModePerm)
			utils.CheckError(mkdirErr)
		}
	}

	// // Concurrent for the lulz and profit.
	var wg sync.WaitGroup
	var files = map[string]func() string{
		"ec2_conf.json": i.defaultEC2Config,
		"sg_conf.json":  i.defaultSGConfig,
		"user_data.sh":  i.defaultUserData,
		"minecraft.key": i.dummyMinecraftContent,
	}
	for k, v := range files {
		wg.Add(1)
		go func(filename string, content func() string) {
			defer wg.Done()
			i.makeDefaultConfigurationForFile(filename, content(), usr)
		}(k, v)
	}
	wg.Wait()

	godb.InitDb()
}

func (i *Init) makeDefaultConfigurationForFile(filename, content string, usr *user.User) {
	path := filepath.Join(configPath, filename)
	if i.exists(path) {
		log.Printf("File '%s' already exists. Nothing to do.", path)
		return
	}
	dst, err := os.Create(path)
	utils.CheckError(err)
	defer dst.Close()
	if _, err = dst.WriteString(content); err != nil {
		utils.CheckError(err)
	}
	log.Printf("Configuration created in home. Filename: %s\n", filename)
}

func (i *Init) exists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// NewInit initializes configuration values.
func NewInit(appName string) *commander.CommandWrapper {
	return &commander.CommandWrapper{
		Handler: &Init{},
		Help: &commander.CommandDescriptor{
			Name:             "init",
			ShortDescription: "Initialize configuration values.",
			LongDescription: `Init initializes configurations in the users home/.config folder.
This command is OS agnostic and on Windows it will create a folder under user/.config/go-furnace.`,
			Arguments: "",
			Examples:  []string{},
		},
	}
}

func (i *Init) defaultEC2Config() string {
	return `{
    "dry_run": true,
    "image_id": "ami-ea26ce85",
    "key_name": "minecraft_keys",
    "min_count": 1,
    "max_count": 1,
    "instance_type": "t2.nano",
    "monitoring": {
        "enabled": true
    }
}`
}

func (i *Init) defaultSGConfig() string {
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

func (i *Init) defaultUserData() string {
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

func (i *Init) dummyMinecraftContent() string {
	return "Dummy minecraft.key file created. Don't forget to fill this out!"
}
