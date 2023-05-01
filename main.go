package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	if len(os.Args) < 2 {
		panic("not enough argument")
	}
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("invalid command")
	}
}

func run() {
	// fmt.Printf("Running %v as PID %d \n", os.Args[2:], os.Getpid())
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWIPC | syscall.CLONE_NEWNET | syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}
	check_err(cmd.Run())

}

func child() {
	fmt.Printf("Running %v as PID %d \n", os.Args[2:], os.Getpid())
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	check_err(set_cgroup_setting(os.Getegid(), "1000", "10000", "100m"))

	check_err(syscall.Mount("/dev", "/home/ubuntu/Home/tech/go-my-container/rootfs/dev", "devtmpfs", syscall.MS_BIND, ""))
	check_err(syscall.Chroot("/home/ubuntu/Home/tech/go-my-container/rootfs"))
	check_err(os.Chdir("/"))
	syscall.Mount("proc", "/proc", "proc", 0, "")

	check_err(syscall.Sethostname([]byte("container")))
	check_err(cmd.Run())

}

func check_err(err error) {
	if err != nil {
		panic(err)
	}
}

func set_cgroup_setting(pid int, cpu_quota string, cpu_period string, memory_max_mb string) error {
	err := os.MkdirAll("/sys/fs/cgroup/my-container/", 700)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("/sys/fs/cgroup/my-container/cgroup.procs", []byte(fmt.Sprintf("%d", pid)), 700)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("/sys/fs/cgroup/my-container/cpu.max", []byte(fmt.Sprintf("%s %s", cpu_quota, cpu_period)), 700)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("/sys/fs/cgroup/my-container/memory.max", []byte(fmt.Sprintf("%s", memory_max_mb)), 700)
	if err != nil {
		return err
	}
	return nil
}
