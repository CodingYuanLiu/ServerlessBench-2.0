package cmd

import (
	"fmt"
	"io/ioutil"
	"path"
	"regexp"
	benchinterface "serverlessbench2/internal/benchinterface"
	"strconv"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create func_name sourcecode1 sourcecode2... requirementfile / create -d func_name dirpath(need requirement.txt)/ create -f flow_name -c src1,src2,... -y mem1,mem2,.. -q req1,req2,... -n stagename1,stagename2,...",
	Short: "create function",
	Run: func(cmd *cobra.Command, args []string) {
		isFlow, _ := cmd.Flags().GetBool("flow")
		if isFlow {
			if len(args) != 1 {
				panic("Usage:cli create -f flow_name --flags....")
			} else {
				stages := []map[string](map[string]string){}
				useDeafultMem := false
				flowName := args[0]
				reqs, _ := cmd.Flags().GetStringSlice("flowreq")
				srcs, _ := cmd.Flags().GetStringSlice("flowsrc")
				mems, _ := cmd.Flags().GetStringSlice("flowmemory")
				names, _ := cmd.Flags().GetStringSlice("flowstagename")
				numOfComponents := len(srcs)
				if len(mems) == 0 {
					useDeafultMem = true
				}
				idx := 0
				for {
					if idx == numOfComponents {
						break
					}
					var mem string
					name := names[idx]
					src := srcs[idx]
					req := reqs[idx]
					if useDeafultMem {
						mem = "128"
					} else {
						mem = mems[idx]
						if mem == "-1" {
							mem = "128"
						}
					}
					append_map := map[string]map[string]string{}
					append_map[name] = map[string]string{"src": src, "req": req, "memory": mem}
					stages = append(stages, append_map)
					idx++
				}
				err := benchinterface.CreateWorkflow(flowName, stages[:]...)
				if err != nil {
					fmt.Printf("%+v\n\n", err)
					panic(err)
				}

			}
		} else {
			fromDir, _ := cmd.Flags().GetBool("dir")
			funcName := args[0]
			memory, _ := cmd.Flags().GetString("memory")
			var memorySize int
			if memory == "" {
				memorySize = 128
			} else {
				memorySize, _ = strconv.Atoi(memory)
			}
			if fromDir {
				if len(args) != 2 {
					panic("Usage:cli create -d func_name dirpath")
				}
				var src_files []string
				codeDir := args[1]
				rx := regexp.MustCompile("requirements.txt$")
				reqMatchedFlag := false

				_dir, err := ioutil.ReadDir(codeDir)
				if err != nil {
					fmt.Printf("%+v\n\n", err)
					panic(err)
				}

				for _, _file := range _dir {
					reqMatched := rx.MatchString(_file.Name())
					if reqMatched {
						if reqMatchedFlag {
							panic("Function code dir must contain only one requirements.txt")
						} else {
							reqMatchedFlag = true
						}
					}
					src_files = append(src_files, path.Join(codeDir, _file.Name()))
				}
				if !reqMatchedFlag {
					panic("Function code dir must contain requirements.txt")
				}
				err = benchinterface.CreateFunction(funcName, memorySize, src_files[:]...)
				if err != nil {
					fmt.Printf("%+v\n\n", err)
					panic(err)
				}
			} else {
				if len(args) < 2 {
					panic("Usage:cli create function_name source_code...")
				} else {
					func_source := args[1:]
					err := benchinterface.CreateFunction(funcName, memorySize, func_source...)
					if err != nil {
						fmt.Printf("%+v\n\n", err)
						panic(err)
					}
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.PersistentFlags().StringP("memory", "m", "", "function's memory size")
	createCmd.PersistentFlags().BoolP("dir", "d", false, "create function from a directory")
	createCmd.PersistentFlags().BoolP("flow", "f", false, "create workflow")
	createCmd.PersistentFlags().StringSliceP("flowmemory", "y", nil, "memory of workflow components")
	createCmd.PersistentFlags().StringSliceP("flowsrc", "c", nil, "srcfile dir of workflow components")
	createCmd.PersistentFlags().StringSliceP("flowreq", "q", nil, "requirements dir of workflow components")
	createCmd.PersistentFlags().StringSliceP("flowstagename", "n", nil, "names dir of workflow components")
}
