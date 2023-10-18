package benchinterface

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strconv"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	config "serverlessbench2/internal/config"
	"serverlessbench2/internal/runner"
	util "serverlessbench2/internal/util"
)

func ApplyYaml(yaml_content util.TestYaml) error {
	// get docker username and write config
	provider := yaml_content.Platform
	resultRoot := yaml_content.Resultpath
	if _, err := os.Stat(resultRoot); os.IsNotExist(err) {
		fmt.Println("Error:", resultRoot, "does not exist")
		return nil
	}
	dockerUserName := config.GetDockerUserName()
	if dockerUserName == "" {
		err := errors.New("username of Docker isn't set.")
		return err
	}
	err := config.ChangeProvider(provider)
	if err != nil {
		return err
	}
	err = config.ChangeTestResultDir(resultRoot)
	if err != nil {
		return err
	}
	// var create_res string
	// func_info stores the information of each function in this workflow.
	// consists of source code path, requirements path, and memory size.
	funcPath := map[string](map[string]string){}
	testName := yaml_content.Name

	// create functions with source code files, and store function information: list(map[func_name](map[info_key(src/req/mem)]value))
	fmt.Printf("[test:--%s--]: Start collecting function information and flushing...\n", testName)
	for _, function := range yaml_content.Component.Function {
		name := function.Name
		srcPath := function.DirPath
		reqPath := function.ReqPath
		memorySize := function.Memory
		mem := strconv.Itoa(memorySize)
		_ = DeleteFunction(name)
		funcPath[name] = map[string]string{"src": srcPath, "req": reqPath, "memory": mem}
	}
	fmt.Printf("[test:--%s--]: Function information collected.\n", testName)

	// create apps(sequences of functions), flush first.
	workflowInfomation := map[string]([]map[string](map[string]string)){}
	fmt.Printf("[test:--%s--]: Start collecting workflow information and flushing...\n", testName)
	for _, app := range yaml_content.Component.Workflow {
		name := app.Name
		currentFunc := app.Stage[0].FuncName
		currentFuncPath := funcPath[currentFunc]
		stages := []map[string](map[string]string){{currentFunc: currentFuncPath}}
		_ = DeleteWorkflow(name)
		// build stages map for each function(information acquired from last step)
		for _, function := range app.Stage[1:] {
			funcName := function.FuncName
			nextStage := map[string]map[string]string{}
			subMap := funcPath[funcName]
			nextStage[funcName] = subMap
			stages = append(stages, nextStage)
		}
		workflowInfomation[name] = stages
	}
	fmt.Printf("[test:--%s--]: Workflow information collected.\n", testName)
	testcaseYamlByte, _ := util.Asset("resources/testcase.yaml")
	testcaseRaw := make(map[interface{}]interface{})
	testcaseInfo := make(map[string](map[string]string))
	if err := yaml.Unmarshal(testcaseYamlByte, &testcaseRaw); err != nil {
		return err
	}
	for name, value := range testcaseRaw {
		testcaseInfoValue := make(map[string]string)
		for k, v := range value.(map[interface{}]interface{}) {
			testcaseInfoValue[k.(string)] = v.(string)
		}
		testcaseInfo[name.(string)] = testcaseInfoValue
	}

	fmt.Printf("[test:--%s--]: Start running testcases...\n", testName)
	for _, testcase := range yaml_content.MetricController.Default {
		testcaseName := testcase.Name
		testcaseType := testcaseInfo[testcaseName]["type"]
		testcaseDir := testcaseInfo[testcaseName]["dir"]
		if testcaseDir == "" {
			fmt.Printf("No such default testcase: %s.\n", testcaseName)
			continue
		}
		for _, testApp := range yaml_content.Test {
			appName := testApp.Name
			appType := testApp.Type
			fmt.Printf("[test:--%s--]: Running testcase [%s] on app [%s]...\n", testName, testcaseName, appName)
			if appType != testcaseType {
				fmt.Printf("[test:--%s--]: Types don't match. Testcase [%s] is type [%s] and app [%s] is type [%s].\n", testName, testcaseName, testcaseType, appName, appType)
				continue
			}

			var appInfo interface{}
			if appType == "single" {
				appInfo = funcPath[appName]
			} else {
				appInfo = workflowInfomation[appName]
			}

			appParams := []string{}
			for _, defaultParam := range testApp.Param.Default {
				paramFile := path.Join("./inputs", appName, defaultParam.Name)
				content, err := os.ReadFile(paramFile)
				if err != nil {
					return err
				}
				appParams = append(appParams, string(content))
			}

			for _, otherParam := range testApp.Param.Other {
				appParams = append(appParams, otherParam.Value)
			}
			for _, param := range appParams {
				TCrunner := runner.RunnerFactory(appType)
				if err := TCrunner.TestcaseRunner(testName, testcaseDir, appName, appInfo, param); err != nil {
					return errors.Wrap(err, fmt.Sprintf("error occured for test %s: running testcase %s on app %s.", testName, testcaseName, appName))
				}
			}
			fmt.Printf("[test:--%s--]: Testcase [%s] on app [%s] finished. \n", testName, testcaseName, appName)
		}
	}
	fmt.Printf("[test:--%s--]: All testcases finished.\n", testName)
	return nil
}

func CreateFunction(function_name string, memorySize int, source_code_file ...string) error {
	rx, _ := regexp.Compile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)
	matched := rx.MatchString(function_name)
	if !matched {
		err := errors.New("function name must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character")
		return err
	}
	provider := CheckProvider()
	if provider == "" {
		err := errors.New("provider can't be empty")
		return err
	}
	p := providerFactory()
	if p == nil {
		err := errors.New(fmt.Sprintf("invalid provider: %s", provider))
		return err
	}
	err := p.DeleteFunction(function_name)
	if err != nil {
		return err
	}
	err = p.CreateFunction(function_name, memorySize, source_code_file...)
	if err != nil {
		return err
	}
	return nil
}

func CreateWorkflow(flow_name string, func_info ...map[string](map[string]string)) error {
	rx, _ := regexp.Compile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)
	matched := rx.MatchString(flow_name)
	if !matched {
		err := errors.New("flow name must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character")
		return err
	}
	provider := CheckProvider()
	if provider == "" {
		err := errors.New("provider can't be empty")
		return err
	}
	p := providerFactory()
	if p == nil {
		err := errors.New(fmt.Sprintf("Invalid provider: %s", provider))
		return err
	}
	_ = p.DeleteWorkflow(flow_name)
	err := p.CreateWorkflow(flow_name, func_info...)
	if err != nil {
		return err
	}
	return nil
}

func InvokeWorkflow(flow_name string, func_names ...string) error {
	provider := CheckProvider()
	if provider == "" {
		err := errors.New("provider can't be empty")
		return err
	}
	p := providerFactory()
	if p == nil {
		err := errors.New(fmt.Sprintf("Invalid provider: %s", provider))
		return err
	}
	err := p.InvokeWorkflow(flow_name, func_names...)
	if err != nil {
		return err
	}
	return nil
}

func InvokeFunction(function_name string, params ...string) error {
	provider := CheckProvider()
	if provider == "" {
		err := errors.New("provider can't be empty")
		return err
	}
	p := providerFactory()
	if p == nil {
		err := errors.New(fmt.Sprintf("Invalid provider: %s", provider))
		return err
	}
	err := p.InvokeFunction(function_name, params...)
	if err != nil {
		return err
	}
	return nil
}

func DeleteFunction(function_name string) error {
	provider := CheckProvider()
	if provider == "" {
		err := fmt.Errorf("provider can't be empty")
		return err
	}
	p := providerFactory()
	if p == nil {
		err := fmt.Errorf("Invalid provider: %s", provider)
		return err
	}
	return p.DeleteFunction(function_name)
}

func DeleteWorkflow(flow_name string) error {
	provider := CheckProvider()
	if provider == "" {
		err := errors.New("provider can't be empty")
		return err
	}
	p := providerFactory()
	if p == nil {
		err := errors.New(fmt.Sprintf("Invalid provider: %s", provider))
		return err
	}
	err := p.DeleteWorkflow(flow_name)
	return err
}

func ListFunction() error {
	provider := CheckProvider()
	if provider == "" {
		err := errors.New("provider can't be empty")
		return err
	}
	p := providerFactory()
	if p == nil {
		err := errors.New(fmt.Sprintf("Invalid provider: %s", provider))
		return err
	}
	err := p.ListFunction()
	if err != nil {
		return err
	}
	return nil
}
