package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// func TestLoadConfigFileIfExists(t *testing.T) {
// 	yamlData := []byte(`main:
//   stackname: MyStack
//   spinner: 1
// aws:
//   code_deploy_role: CodeDeployServiceRole
//   region: us-east-1
//   enable_plugin_system: false
//   template_name: mystack.template
//   code_deploy:
//     code_deploy_s3_bucket: furnace_code_bucket
//     code_deploy_s3_key: furnace_deploy_app
//     git_account: Skarlso/furnace-codedeploy-app
//     git_revision: b89451234...`)
// 	location := os.TempDir()
// 	err := ioutil.WriteFile(filepath.Join(location, ".testexists.furnace"), []byte("testexists.yaml"), os.ModePerm)
// 	if err != nil {
// 		t.Fatal("failed to create temporary file for testing: ", err)
// 	}
// 	err = ioutil.WriteFile(filepath.Join(location, "testexists.yaml"), yamlData, os.ModePerm)
// 	if err != nil {
// 		t.Fatal("failed to create temporary file for testing: ", err)
// 	}
// 	err = LoadConfigFileIfExists(location, "testexists")
// 	if err != nil {
// 		t.Fatal("failed to load config file: ", err)
// 	}
// }

func TestLoadConfigFileIfNotExists(t *testing.T) {
	err := LoadConfigFileIfExists(os.TempDir(), "testnotexists")
	if err == nil {
		t.Fatal("loading a non existing config file should have errored out")
	}
}

func TestLoadConfigFileIfExistsOutSideTheCurrentDir(t *testing.T) {
	yamlData := []byte(`main:
  stackname: MyStack
  spinner: 1
aws:
  code_deploy_role: CodeDeployServiceRole
  region: us-east-1
  enable_plugin_system: false
  template_name: mystack.template
  code_deploy:
    code_deploy_s3_bucket: furnace_code_bucket
    code_deploy_s3_key: furnace_deploy_app
    git_account: Skarlso/furnace-codedeploy-app
    git_revision: b89451234...`)
	location := os.TempDir()
	location2 := filepath.Join(location, "temp2")
	os.Mkdir(filepath.Join(location2), os.ModeDir)
	err := ioutil.WriteFile(filepath.Join(location, ".testdiffdir.furnace"), []byte("testdiffdir.yaml"), os.ModePerm)
	if err != nil {
		t.Fatal("failed to create temporary file for testing: ", err)
	}
	err = ioutil.WriteFile(filepath.Join(location, "testdiffdir.yaml"), yamlData, os.ModePerm)
	if err != nil {
		t.Fatal("failed to create temporary file for testing: ", err)
	}
	err = LoadConfigFileIfExists(location2, "testdiffdir")
	if err != nil {
		t.Fatal("failed to load config file: ", err)
	}
}
