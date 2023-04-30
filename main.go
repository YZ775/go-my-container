package main

import (
	"fmt"
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
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}
	check_err(cmd.Run())

}

func child() {
	fmt.Printf("Running %v as PID %d \n", os.Args[2:], os.Getpid())
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	current_path, _ := os.Getwd()
	new_root := fmt.Sprintf("NEW ROOT: %s/rootfs\n", current_path)
	fmt.Print(new_root)
	// check_err(syscall.Chroot(new_root))
	check_err(syscall.Chroot("/home/ubuntu/Home/tech/go-my-container/rootfs"))

	check_err(os.Chdir("/"))
	syscall.Mount("proc", "proc", "proc", 0, "")
	check_err(cmd.Run())

}

func check_err(err error) {
	if err != nil {
		panic(err)
	}
}
