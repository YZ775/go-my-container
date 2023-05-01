package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"
)

type DockerManifest []struct {
	Config   string   `json:"Config"`
	RepoTags []string `json:"RepoTags"`
	Layers   []string `json:Layers`
}

func main() {
	if len(os.Args) < 2 {
		panic("not enough argument")
	}
	var err error
	switch os.Args[1] {
	case "run":
		err = run()
	case "child":
		err = child()
	case "load":
		err = load_image("nginx.tar")
	default:
		panic("invalid command")
	}
	if err != nil {
		panic(err)
	}
}

func run() error {
	fmt.Printf("Running main process as PID %d \n", os.Getpid())
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWIPC | syscall.CLONE_NEWNET | syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func child() error {
	fmt.Printf("[Child]Running %v as PID %d \n", os.Args[2:], os.Getpid())
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := set_cgroup_setting(os.Getegid(), "1000", "10000", "100m")
	if err != nil {
		return err
	}

	err = syscall.Mount("/dev", "/home/ubuntu/Home/tech/go-my-container/rootfs2/dev", "devtmpfs", syscall.MS_BIND, "")
	if err != nil {
		return err
	}
	err = syscall.Chroot("/home/ubuntu/Home/tech/go-my-container/rootfs2")
	if err != nil {
		return err
	}
	err = os.Chdir("/")
	if err != nil {
		return err
	}
	syscall.Mount("proc", "/proc", "proc", 0, "")
	if err != nil {
		return err
	}
	err = syscall.Sethostname([]byte("container"))
	if err != nil {
		return err
	}
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
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

func load_image(filename string) error {
	fmt.Print(fmt.Sprintf("./image/%s\n", filename))
	cmd := exec.Command("tar", "-xvf", fmt.Sprintf("./image/%s", filename), "-C", "./image/archive")
	err := cmd.Run()
	if err != nil {
		return (err)
	}
	byteArray, err := ioutil.ReadFile("./image/archive/manifest.json")
	if err != nil {
		return (err)
	}
	manifest := make(DockerManifest, 0)
	err = json.Unmarshal(byteArray, &manifest)
	if err != nil {
		return (err)
	}
	top_layer := manifest[0].Layers[0]
	cmd = exec.Command("tar", "-xvf", fmt.Sprintf("./image/archive/%s", top_layer), "-C", "rootfs2")
	err = cmd.Run()
	if err != nil {
		return (err)
	}
	return nil
}
