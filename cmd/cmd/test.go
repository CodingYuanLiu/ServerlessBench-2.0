package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "for test",
	Run: func(cmd *cobra.Command, args []string) {
		cmd1 := exec.Command("kubectl exec -it curl -- curl -v  http://broker-ingress.knative-eventing.svc.cluster.local/default/default -X POST -H \"Ce-Id: say-hello\" -H \"Ce-Specversion: 1.0\" -H \"Ce-Type: greeting\" -H \"Ce-Source: not-sendoff\" -H \"Content-Type: application/json\" -d '{\"text\":\"test\"}'")
		err := cmd1.Run()
		if err != nil {
			fmt.Printf("%+v\n\n", err)
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
