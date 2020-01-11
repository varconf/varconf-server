package cmd

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"runtime"
)

func Execute() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	h := flag.Bool("h", false, "this help!")
	d := flag.Bool("d", false, "is daemon?")

	p := flag.String("p", "./", "config `dir` for server!")
	s := flag.String("s", "", "`start|stop` for server!")
	flag.Parse()

	if *h {
		flag.Usage()
	}

	if *s != "" {
		manageServer(*p, *s, *d)
	} else {
		flag.Usage()
	}
}

func manageServer(dir, operate string, daemon bool) {
	pidFile := path.Join(dir, "./pid.lock")
	configFile := path.Join(dir, "./config.json")

	switch operate {
	case "start":
		fmt.Println("Starting!")
		if daemon {
			cmd := exec.Command(os.Args[0], "-p", dir, "-s", "start")
			err := cmd.Start()
			if err == nil {
				fmt.Printf("PID %d is running...\n", cmd.Process.Pid)
			} else {
				fmt.Println("Start failed!", err.Error())
			}
		} else {
			pid := fmt.Sprintf("%d", os.Getpid())
			err := ioutil.WriteFile(pidFile, []byte(pid), 0666)
			if err != nil {
				fmt.Println("Start failed!", err.Error())
				return
			}
			err = Start(configFile)
			if err != nil {
				fmt.Println("Start failed!", err.Error())
				os.Remove(pidFile)
			}
		}
		break

	case "stop":
		fmt.Println("Stopping!")
		pb, err := ioutil.ReadFile(pidFile)
		if err != nil {
			fmt.Println("Read PID error!", err)
			return
		}

		pid := string(pb)
		cmd := new(exec.Cmd)
		if runtime.GOOS == "windows" {
			cmd = exec.Command("taskkill", "/f", "/pid", pid)
		} else {
			cmd = exec.Command("kill", pid)
		}

		err = cmd.Start()
		if err == nil {
			fmt.Printf("PID %s has been stopped!\n", pid)
			os.Remove(pidFile)
		} else {
			fmt.Println("PID "+pid+" stop failed! %s\n", err)
		}
		break

	default:
		fmt.Println("Unknown operation!")
	}
}
