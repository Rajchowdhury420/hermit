package task

import (
	"fmt"
	"strings"

	"github.com/hideckies/hermit/pkg/common/meta"
)

func SetTask(task string, agentName string) error {
	// Adjust task arguments
	if strings.HasPrefix(task, "cp ") ||
		strings.HasPrefix(task, "download ") ||
		strings.HasPrefix(task, "upload ") {

		taskSplit := strings.Split(task, " ")
		if len(taskSplit) != 3 {
			return fmt.Errorf("invalid the number of arguments")
		}

		cmd := taskSplit[0]
		src := taskSplit[1]
		dest := taskSplit[2]

		if dest == "." {
			srcSplit := strings.Split(src, "/")
			dest = srcSplit[len(srcSplit)-1]
		}
		if dest[len(dest)-1] == '/' {
			dest = fmt.Sprintf("%s%s", dest, src)
		}
		task = cmd + " " + src + " " + dest
	}
	if task == "ls" {
		task = "ls ."
	}

	err := meta.WriteTask(agentName, task, false)
	if err != nil {
		return err
	}
	return nil
}
