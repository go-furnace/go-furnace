package config

import (
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/Skarlso/go-furnace/utils"
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

var configPath string

// WAITFREQUENCY global wait frequency default.
var WAITFREQUENCY = 1

// STACKNAME is the default name for a stack.
const STACKNAME = "FurnaceStack"

// CODEDEPLOYROLE is the default name of the codedeploy role.
const CODEDEPLOYROLE = "CodeDeployServiceRole"

// GITREVISION is the revision number to deploy.
var GITREVISION string

// GITACCOUNT is the account/project from which to deploy.
var GITACCOUNT string

// CFClient abstraction for cloudFormation client.
type CFClient struct {
	Client cloudformationiface.CloudFormationAPI
}

// Path retrieves the main configuration path.
func Path() string {
	// Get configuration path
	usr, err := user.Current()
	utils.CheckError(err)
	return filepath.Join(usr.HomeDir, ".config", "go-furnace")
}

func init() {
	configPath = Path()
	GITACCOUNT = os.Getenv("GIT_ACCOUNT")
	GITREVISION = os.Getenv("GITREVISION")
	if len(GITACCOUNT) < 1 {
		log.Fatal("Please define a git account and project to deploy from in the form of: account/project under GIT_ACCOUNT.")
	}
	if len(GITREVISION) < 1 {
		log.Fatal("Please define the git commit hash to use for deploying under GIT_REVISION.")
	}
}

// LoadCFStackConfig Load the CF stack configuration file into a []byte.
func LoadCFStackConfig() []byte {
	dat, err := ioutil.ReadFile(filepath.Join(configPath, "cloud_formation.json"))
	utils.CheckError(err)
	return dat
}
