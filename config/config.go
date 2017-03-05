package config

import (
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
)

const (
	// PRECREATE Event name for plugins
	PRECREATE = "pre-create"
	// POSTCREATE Event name for plugins
	POSTCREATE = "post-create"
	// PREDELETE Event name for plugins
	PREDELETE = "pre-delete"
	// POSTDELETE Event name for plugins
	POSTDELETE = "post-delete"
)

// CODEDEPLOYROLE is the default name of the codedeploy role.
const CODEDEPLOYROLE = "CodeDeployServiceRole"

// REGION to operate in.
var REGION string

var configPath string

// WAITFREQUENCY global wait frequency default.
var WAITFREQUENCY = 1

// STACKNAME is the default name for a stack.
var STACKNAME = "FurnaceStack"

// SPINNER is the index of which spinner to use. Defaults to 7.
var SPINNER int

// CFClient abstraction for cloudFormation client.
type CFClient struct {
	Client cloudformationiface.CloudFormationAPI
}

// Path retrieves the main configuration path.
func Path() string {
	// Get configuration path
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("Error occurred: %s", err.Error())
	}
	return filepath.Join(usr.HomeDir, ".config", "go-furnace")
}

func init() {
	configPath = Path()
	REGION = os.Getenv("FURNACE_REGION")
	spinner := os.Getenv("FURNACE_SPINNER")
	if len(spinner) < 1 {
		SPINNER = 7
	} else {
		SPINNER, _ = strconv.Atoi(spinner)
	}
	if len(REGION) < 1 {
		log.Fatal("Please define a region to operate in with FURNACE_REGION exp: config.REGION.")
	}
	stackname := os.Getenv("FURNACE_STACKNAME")
	if len(stackname) > 0 {
		STACKNAME = stackname
	}
}

// LoadCFStackConfig Load the CF stack configuration file into a []byte.
func LoadCFStackConfig() []byte {
	dat, err := ioutil.ReadFile(filepath.Join(configPath, "cloud_formation.json"))
	if err != nil {
		log.Fatalf("Error occurred: %s", err.Error())
	}
	return dat
}
