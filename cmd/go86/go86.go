package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"

	glog "github.com/golang/glog"
	bios "go86.org/go86/bios"
	cpu "go86.org/go86/cpu"
	deb "go86.org/go86/debugger"
	dos "go86.org/go86/dos"
)

func doinst(opcodes string) bool {
	fmt.Printf("OpCodes: [%s]\n\n", opcodes)
	d, err := hex.DecodeString(opcodes)
	if err != nil {
		return false
	}

	cs := 0x1000
	c := cpu.NewCpu(1024 * 1024)
	bios.NewBios(c)
	dos.NewDos(c)
	copy(c.Mem.At(cs, 0), d)
	c.Sregs[cpu.SREG_CS] = 0x1000
	c.Sregs[cpu.SREG_DS] = 0x1000
	c.Ip = 0
	c.Run()
	return true
}

func dorun(filename string) bool {
	exe, err := dos.ReadExeFromFile(filename)
	if err != nil {
		fmt.Printf("Failed to read file header from: '%s'; error: %s\n", filename, err)
		return false
	}

	if *runForceType {
		exe.Etype = dos.IMAGE
	}

	c := cpu.NewCpu(1024 * 1024)
	bios.NewBios(c)
	di := dos.NewDos(c)
	_, err = di.Load(exe)
	if err != nil {
		fmt.Println(err)
		return false
	}

	if *dbg == "gdb" || *dbg == "lame" {
		request := make(chan deb.DebuggerRequest)
		response := make(chan deb.DebuggerResponse, 5)
		deb.EnableDebugger(c, 1234, *dbg, request, response)

	}
	c.Run()
	fmt.Println("")
	return true
}

var (
	dbg          = flag.String("dbg", "none", "Enable and specifies the type of debugger.  Values are: none, gdb, lame")
	runcmd       = flag.NewFlagSet("run", flag.ExitOnError)
	instcmd      = flag.NewFlagSet("inst", flag.ExitOnError)
	runForceType = runcmd.Bool("image", false, "load binary as binary image instead of COM or EXE")
)

func showHelp() {
	fmt.Println(`
go86 is a tools for executing x86 Code.

Usage:

	go86 [arguments] <command> [command arguments]

The commands are:

	inst        Execute a string of opcodes
	run         Execute a DOS executable (exe, com, or binary image)
	help        Displays help
		
Use "go86 help <command>" for more information about that command.

Program arguments:
	`)
	flag.PrintDefaults()
}

func main() {
	flag.Parse()

	if len(os.Args) < 2 {
		glog.Warning("expected 'run' or 'inst' subcommands")
		showHelp()
		os.Exit(1)
	}

	args := flag.Args()
	switch args[0] {
	case "run":
		runcmd.Parse(args[1:])
		if runcmd.NArg() < 1 {
			fmt.Print("Go86\n\nUsage: go86 run <DOS EXE>.\n")
			showHelp()
			os.Exit(1)
		}
		if !dorun(runcmd.Arg(0)) {
			os.Exit(1)
		}
		// do cmd
		os.Exit(0)
	case "inst":
		// Example inst (hello world):
		// 8D161500B409CD21B87F00BA010002C2B44CCD21000A0D48656C6C6F20576F726C640D0A0A0A24
		instcmd.Parse(args[1:])
		if instcmd.NArg() < 1 {
			flag.Usage()
			fmt.Print("Go86\n\nUsage: go86 int <HEX format OpCodes>.\n")
			showHelp()
			os.Exit(1)
		}
		if !doinst(instcmd.Arg(0)) {
			os.Exit(1)
		}
		os.Exit(0)
	case "help":
		if flag.NArg() > 1 {
			switch args[1] {
			case "inst":
				fmt.Println("inst [opcodes] - execute opcodes")
				instcmd.PrintDefaults()
				os.Exit(0)
			case "run":
				fmt.Println("run [executable] - execute DOS executable")
				runcmd.PrintDefaults()
				os.Exit(0)
			}
		}
		showHelp()
		os.Exit(0)
	default:
		fmt.Print("ERROR: No command specified.\n\n")
		showHelp()
		os.Exit(0)
	}
}
