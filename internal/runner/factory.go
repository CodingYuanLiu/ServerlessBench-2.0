package runner

type TCRunner interface {
	TestcaseRunner(testName string, testcaseDir string, appName string, appInfo interface{}, param string) error
}

func RunnerFactory(appType string) TCRunner {
	switch appType {
	case "single":
		return SingleRunner{}
	case "workflow":
		return WorkflowRunner{}
	}
	return nil
}
