package util

type TestYaml struct {
	Name             string           `yaml:"name,omitempty"`
	MetricController MetricController `yaml:"metric-controller,omitempty"`
	Platform         string           `yaml:"platform,omitempty"`
	Component        Component        `yaml:"component,omitempty"`
	Resultpath       string           `yaml:"resultpath,omitempty"`
	Test             []Test           `yaml:"test,omitempty"`
}

// most outside
type MetricController struct {
	Default []Default `yaml:"default,omitempty"`
	Custom  []Custom  `yaml:"custom,omitempty"`
}

type Component struct {
	Function []Function `yaml:"function,omitempty"`
	Workflow []Workflow `yaml:"workflow,omitempty"`
}

type Test struct {
	Name  string `yaml:"name,omitempty"`
	Type  string `yaml:"type,omitempty"`
	Param Param  `yaml:"param,omitempty"`
}

// inner
type Default struct {
	Name string `yaml:"name,omitempty"`
}

type Custom struct {
	Name    string `yaml:"name,omitempty"`
	Testdir string `yaml:"testdir,omitempty"`
}

type Function struct {
	Name    string `yaml:"name,omitempty"`
	DirPath string `yaml:"dirpath,omitempty"`
	ReqPath string `yaml:"reqpath,omitempty"`
	Memory  int    `yaml:"memory,omitempty"`
}

type Workflow struct {
	Name    string  `yaml:"name,omitempty"`
	DirPath string  `yaml:"dirpath,omitempty"`
	Stage   []Stage `yaml:"stage,omitempty"`
}

type Stage struct {
	FuncName string `yaml:"funcname,omitempty"`
}

type Param struct {
	Default []Default `yaml:"default,omitempty"`
	Other   []Other   `yaml:"other,omitempty"`
}

type Other struct {
	Value string `yaml:"value,omitempty"`
}

// -----------------------for kn----------------------------------
type Conf struct {
	ApiVersion string   `yaml:"apiVersion,omitempty"`
	Kind       string   `yaml:"kind,omitempty"`
	Metadata   Metadata `yaml:"metadata,omitempty"`
	Spec       Spec     `yaml:"spec,omitempty"`
}

type Metadata struct {
	Name string `yaml:"name,omitempty"`
}

type Spec struct {
	ChannelTemplate ChannelTemplate `yaml:"channelTemplate,omitempty"`
	Steps           []Steps         `yaml:"steps,omitempty"`
	Reply           Reply           `yaml:"reply,omitempty"`
}

type Steps struct {
	Ref Ref `yaml:"ref,omitempty"`
}

type ChannelTemplate struct {
	ApiVersion string `yaml:"apiVersion,omitempty"`
	Kind       string `yaml:"kind,omitempty"`
}

type Reply struct {
	Ref Ref `yaml:"ref,omitempty"`
}

type Ref struct {
	ApiVersion string `yaml:"apiVersion,omitempty"`
	Kind       string `yaml:"kind,omitempty"`
	Name       string `yaml:"name,omitempty"`
}
