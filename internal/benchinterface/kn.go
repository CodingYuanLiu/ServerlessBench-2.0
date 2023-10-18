package benchinterface

import (
	"bytes"
	"context"

	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"

	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/homedir"

	config "serverlessbench2/internal/config"
	util "serverlessbench2/internal/util"
)

func (kn ProviderKnative) CreateWorkflow(flowName string, func_info ...map[string](map[string]string)) error {
	workflowYamlName := config.GetKnWorkflowYaml()
	diaplayerAndSenderYaml := config.GetKnDiaplayerAndSenderYaml()
	goWrapperDir := config.GetKnGoWrapperDir()
	goWrapperFile := config.GetKnGoWrapperFile()
	knServiceYamlName := config.GetKnServiceYaml()
	knPythonDir := config.GetKnPythonDir()
	knWorkflowDockerfile := config.GetKnWorkflowDockerfile()

	userName := config.GetDockerUserName()
	if userName == "" {
		err := errors.New("Docker username isn't set in config.")
		return err
	}

	// formalize workflow name because of possible confliction with broker name.
	rx, _ := regexp.Compile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)
	matched := rx.MatchString(flowName)
	if !matched {
		err := errors.New("flow name must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character")
		return err
	}
	displayerName := flowName + "-display"

	fmt.Printf("Start creating knative workflow %s\n", flowName)

	// copy templates
	util.LoadFile(workflowYamlName, ".")
	util.LoadFile(diaplayerAndSenderYaml, ".")

	//-----------------------------------create sequence components-----------------------------------
	functionNames := []string{}

	// func_info stores the infomation of each function in this workflow.
	// consists of source code path, requirements path, and memory size.
	for _, finfo := range func_info {
		for name := range finfo {
			matched := rx.MatchString(name)
			if !matched {
				err := errors.New("Stage name must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character")
				return err
			}
			stageName := flowName + "-stage-" + name
			fmt.Printf("Flush the existing function stage %s.\n", stageName)
			err := DeleteFunction(stageName)
			if err != nil {
				return err
			}
			var functionFilePath []string
			functionNames = append(functionNames, stageName)
			src_path := finfo[name]["src"]
			req_path := finfo[name]["req"]
			memory, err := strconv.Atoi(finfo[name]["memory"])
			if err != nil {
				return err
			}
			// from src_path get paths of all source code files.
			functionFilePath, err = util.ParseFilesFromDir(src_path)
			if err != nil {
				return err
			}
			functionFilePath = append(functionFilePath, req_path)
			mask := syscall.Umask(0)
			defer syscall.Umask(mask)

			// util.KnTempDir is for creating function wrapped by go.
			err = os.Mkdir(util.KnTempDir, 0777)
			if err != nil {
				return err
			}
			err = os.Mkdir(util.KnTempDir+"/"+goWrapperDir, 0777)
			if err != nil {
				return err
			}
			err = os.Mkdir(util.KnTempDir+"/"+knPythonDir, 0777)
			if err != nil {
				return err
			}

			// copy templates and wrapper files
			util.LoadFile(knWorkflowDockerfile, util.KnTempDir)
			err = os.Rename(util.KnTempDir+"/"+knWorkflowDockerfile, util.KnTempDir+"/Dockerfile")
			if err != nil {
				return err
			}
			util.LoadFile(knServiceYamlName, util.KnTempDir)
			util.LoadFile("go.mod", util.KnTempDir+"/"+goWrapperDir)
			util.LoadFile("go.sum", util.KnTempDir+"/"+goWrapperDir)
			util.LoadFile(goWrapperFile, util.KnTempDir+"/"+goWrapperDir)

			// copy source code files
			for _, file := range functionFilePath[:] {
				fmt.Printf("file:%s\n", file)
				util.Cp(file, util.KnTempDir+"/"+knPythonDir)
			}

			err = createKnService(stageName, userName, memory, util.KnTempDir)
			if err != nil {
				return err
			}
			err = os.RemoveAll(util.KnTempDir)
			if err != nil {
				return err
			}
			fmt.Printf("function %s for workflow %s is created.\n", stageName, flowName)
		}
	}

	// -----------------------------------create sequence----------------------------------

	// construct workflow.yaml for creating workflow, replace keywords in the template, and append step infomation.

	err := util.ReplaceWordInFile(workflowYamlName, []string{"NAME"}, []string{flowName})
	if err != nil {
		return err
	}

	var steps []util.Steps
	for _, funcname := range functionNames {
		ref := util.Ref{
			ApiVersion: "serving.knative.dev/v1",
			Kind:       "Service",
			Name:       funcname,
		}
		step := util.Steps{
			Ref: ref,
		}
		steps = append(steps, step)
	}

	workflowYamlContent := util.Conf{}
	workflowYaml, err := ioutil.ReadFile(workflowYamlName)
	if err != nil {
		return err
	}
	if err = yaml.Unmarshal(workflowYaml, &workflowYamlContent); err != nil {
		return err
	}
	workflowYamlContent.Spec.Steps = steps

	reply_ref := util.Ref{
		ApiVersion: "serving.knative.dev/v1",
		Kind:       "Service",
		Name:       displayerName,
	}
	seq_reply := util.Reply{
		Ref: reply_ref,
	}
	workflowYamlContent.Spec.Reply = seq_reply

	yamlData, err := yaml.Marshal(&workflowYamlContent)
	if err != nil {
		return err
	}

	if err = ioutil.WriteFile(workflowYamlName, yamlData, 0666); err != nil {
		return err
	}

	err = util.KubectlApplyFile(workflowYamlName)
	if err != nil {
		return err
	}

	fmt.Println("Sequence created.")

	// --------------------------------create displayer and the Pod for sending event.--------------------------------

	yamlKeyList := []string{"DISPLAYNAME", "MSGTYPE", "SEQNAME"}
	yamlValueList := []string{displayerName, flowName + "-msg", flowName}

	err = util.ReplaceWordInFile(diaplayerAndSenderYaml, yamlKeyList, yamlValueList)
	if err != nil {
		return err
	}

	err = util.KubectlApplyFile(diaplayerAndSenderYaml)
	if err != nil {
		return err
	}
	fmt.Println("Displayer and event sender created.")

	// ---------------------------------------------getting url---------------------------------------------
	createFlag := false
	for i := 0; i < 12; i++ {
		brokers := getKnComponents(util.TYPE_SEQUENCE)
		val, found := brokers[flowName]
		if !found {
			fmt.Println("Getting Url of the new deployment...")
			time.Sleep(5 * time.Second)
		} else {
			fmt.Printf("Create succeed! You can invoke the sequence by %s\n(only inside the cluster)\n", val)
			createFlag = true
			break
		}
	}
	err = os.Remove(workflowYamlName)
	if err != nil {
		return err
	}
	err = os.Remove(diaplayerAndSenderYaml)
	if err != nil {
		return err
	}
	if !createFlag {
		err := errors.New("Getting Url of the new deployment falied.")
		return err
	}
	return nil
}

func (kn ProviderKnative) CreateFunction(funName string, MemorySize int, src ...string) error {
	knServiceYamlName := config.GetKnServiceYaml()
	knEntrypointFile := config.GetKnServiceFile()
	//get docker username
	userName := config.GetDockerUserName()
	if userName == "" {
		err := errors.New("Docker username isn't set in config.")
		return err
	}
	rx, _ := regexp.Compile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)
	matched := rx.MatchString(funName)
	if !matched {
		err := errors.New("function name must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character")
		return err
	}
	// create temporary directory
	mask := syscall.Umask(0)
	defer syscall.Umask(mask)

	err := os.Mkdir(util.KnTempDir, 0777)
	if err != nil {
		return err
	}

	// copy templates
	util.LoadFile("Dockerfile", util.KnTempDir)
	util.LoadFile(knEntrypointFile, util.KnTempDir)
	util.LoadFile(knServiceYamlName, util.KnTempDir)

	// copy file to temporary directory
	for _, file := range src {
		util.Cp(file, util.KnTempDir)
	}
	err = createKnService(funName, userName, MemorySize, util.KnTempDir)
	if err != nil {
		return err
	}

	// get the urls, timeout 60s
	for i := 0; i < 12; i++ {
		services := getKnComponents(util.TYPE_SERVICE)
		val, found := services[funName]
		if !found {
			fmt.Println("Getting Url of the new deployment...")
			time.Sleep(5 * time.Second)
		} else {
			fmt.Printf("Create succeed! You can invoke the benchmark by %s\n", val)
			err = os.RemoveAll(util.KnTempDir)
			if err != nil {
				return err
			}
			return nil
		}
	}
	err = errors.New("fail to get url of the new Deployment.")
	return err
}

func createKnService(funcName string, userName string, memory_size int, func_path string) error {
	knServiceYamlName := config.GetKnServiceYaml()
	curdir, _ := os.Getwd()
	err := os.Chdir(func_path)
	if err != nil {
		return err
	}

	// configure service.yaml
	yamlKeyList := []string{"USERNAME", "NAME", "MEM", "CPU"}
	yamlValueList := []string{userName, funcName, fmt.Sprintf("%d", memory_size), fmt.Sprintf("%d", int(memory_size*1000/128))}

	err = util.ReplaceWordInFile(knServiceYamlName, yamlKeyList, yamlValueList)
	if err != nil {
		return err
	}

	// create docker image
	fmt.Println("\nStart creating docker image...")
	err = util.DockerBuildPush(userName+"/"+strings.ToLower(funcName), "build")
	if err != nil {
		return err
	}
	// push image
	fmt.Println("Start pushing docker image...")
	err = util.DockerBuildPush(userName+"/"+strings.ToLower(funcName), "push")
	if err != nil {
		return err
	}
	// deploy on knative
	err = util.KubectlApplyFile(knServiceYamlName)
	if err != nil {
		return err
	}

	err = os.Chdir(curdir)
	if err != nil {
		return err
	}

	return nil

}

func (kn ProviderKnative) InvokeWorkflow(flowName string, params ...string) error {
	msgName := flowName + "-msg"
	displayerName := flowName + "-display"
	namespace := "default"
	sequences := getKnComponents(util.TYPE_SEQUENCE)
	flowUrl, found := sequences[flowName]
	if !found {
		err := fmt.Errorf("No such flow: %s", flowName)
		return err
	}

	// build request params
	jsonStr := "{"
	if len(params)%2 != 0 {
		err := fmt.Sprintf("params must be key-value pairs")
		return errors.New(err)
	}
	for i := 0; i < len(params); i += 2 {
		param := fmt.Sprintf(`"%s":"%s",`, params[i], params[i+1])
		jsonStr += param
	}
	jsonStr = "{\"message\":" + jsonStr[:len(jsonStr)-1] + "}}"

	// config for sending k8s request.
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	// send HTTP post request inside the POD.
	command := []string{"curl", "-v", flowUrl, "-X", "POST", "-H", "Ce-Id: curl-message", "-H", "Ce-Specversion: 1.0", "-H", "Ce-Type: " + msgName, "-H", "Ce-Source: pod-curl", "-H", "Content-Type: application/json", "-d", jsonStr}
	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name("curl-serverlessbench").
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Command: command,
			Stdin:   false,
			Stdout:  true,
			Stderr:  true,
			TTY:     false,
		}, scheme.ParameterCodec)
	exec_curl, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return err
	}

	// get return value from log of displayer Pod.
	log_output, _ := viewDisplayerLog(displayerName, "user-container", namespace, clientset)
	old_log_length := len(log_output)

	if err = exec_curl.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: ioutil.Discard,
		Stderr: ioutil.Discard,
		Tty:    false,
	}); err != nil {
		return err
	}

	// wait until the result is avaliable or exceeding time limit.
	before := time.Now().Unix()
	var resForPrint string
	rx := regexp.MustCompile(".*\"message\":")
	for {
		after := time.Now().Unix()
		if after-before > 20 {
			break
		}
		res, _ := viewDisplayerLog(displayerName, "user-container", namespace, clientset)
		idx := rx.FindAllStringIndex(res, -1)
		if len(idx) == 0 {
			continue
		}
		resForPrint = strings.Replace(strings.Replace(res[idx[len(idx)-1][1]:len(res)-2], "\n", "", -1), " ", "", -1)
		if len(res) != old_log_length && strings.HasPrefix(resForPrint, "{") {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Println(resForPrint)
	return nil
}

func viewDisplayerLog(displayer_name string, pod_container_name string, namespace string, clientset *kubernetes.Clientset) (string, error) {
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	var displayerPod corev1.Pod

	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.Name, displayer_name) {
			displayerPod = pod
		}
	}
	podLogOpts := corev1.PodLogOptions{
		Container: pod_container_name,
	}
	req := clientset.CoreV1().Pods(namespace).GetLogs(displayerPod.Name, &podLogOpts)
	podLogs, err := req.Stream(context.TODO())
	if err != nil {
		return "", err
	}
	byteArray, err := ioutil.ReadAll(podLogs)
	if err != nil {
		return "", err
	}
	return string(byteArray), nil
}

func (kn ProviderKnative) InvokeFunction(function_name string, params ...string) error {
	// get url of function.
	services := getKnComponents(util.TYPE_SERVICE)
	url, found := services[function_name]
	if !found {
		err := errors.New("No such service")
		return err
	}

	// send request
	jsonStr := "{"
	if len(params)%2 != 0 {
		return errors.New(fmt.Sprintf("params must be key-value pairs"))
	}
	for i := 0; i < len(params); i += 2 {
		param := fmt.Sprintf(`"%s":"%s",`, params[i], params[i+1])
		jsonStr += param
	}
	jsonStr = jsonStr[:len(jsonStr)-1] + "}"

	var req *http.Request
	if len(params) != 0 {
		req, _ = http.NewRequest("GET", url, bytes.NewBuffer([]byte(jsonStr)))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, _ = http.NewRequest("GET", url, nil)
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)
	fmt.Print(string(body))
	return nil
}

func (kn ProviderKnative) DeleteWorkflow(flowName string) error {
	displayer_name := flowName + "-display"
	Sequence := getKnComponents(util.TYPE_SEQUENCE)
	_, found := Sequence[flowName]
	if !found {
		return nil
	}
	services := getKnComponents(util.TYPE_SERVICE)
	flowStagePrefix := flowName + "-stage-"
	for stageName, _ := range services {
		if strings.HasPrefix(stageName, flowStagePrefix) {
			cmd := exec.Command("kn", "service", "delete", stageName)
			out, _ := cmd.CombinedOutput()
			fmt.Print(string(out))
		}
	}
	cmd := exec.Command("kn", "service", "delete", displayer_name)
	out, _ := cmd.CombinedOutput()
	fmt.Print(string(out))
	cmd = exec.Command("kubectl", "delete", "Sequence", flowName)
	out, _ = cmd.CombinedOutput()
	fmt.Print(string(out))
	return nil
}

func (kn ProviderKnative) DeleteFunction(function_name string) error {
	services := getKnComponents(util.TYPE_SERVICE)
	_, found := services[function_name]
	if !found {
		return nil
	}
	cmd := exec.Command("kn", "service", "delete", function_name)
	out, _ := cmd.Output()
	fmt.Print(string(out))
	return nil
}

func (kn ProviderKnative) ListFunction() error {
	services := getKnComponents(util.TYPE_SERVICE)
	sequences := getKnComponents(util.TYPE_SEQUENCE)
	fmt.Println("Single services and their URLs.")
	fmt.Println("NAME\tURL")
	for name, url := range services {
		fmt.Println(name + "\t" + url)
	}
	fmt.Print("\n")
	fmt.Println("Names of knative workflows.")
	fmt.Println("NAME\tURL")
	for name, url := range sequences {
		fmt.Println(name + "\t" + url)
	}
	return nil
}

func getKnComponents(comp_type util.KnComponent) map[string]string {
	var listRes string
	if comp_type == util.TYPE_SERVICE {
		cmd := exec.Command("kn", "service", "list")
		out, err := cmd.Output()
		if err != nil {
			return nil
		}
		listRes = string(out)
	} else if comp_type == util.TYPE_SEQUENCE {
		cmd := exec.Command("kubectl", "get", "Sequence")
		out, err := cmd.Output()
		if err != nil {
			return nil
		}
		listRes = string(out)
	} else {
		fmt.Println("Please assign correct component type!(util.TYPE_SERVICE/util.TYPE_BROKER)")
		return nil
	}
	rx, _ := regexp.Compile(`\S*\s*https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`)
	services := rx.FindAllString(listRes, -1)
	ret := make(map[string]string)
	for _, service := range services {
		words := strings.Fields(service)
		ret[words[0]] = words[1]
	}
	return ret
}
