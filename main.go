package main()


func main(){
	switch os.Args[1]{
	case "run":
		run()
	defualt:
		panig("invalid command")
	}
}
