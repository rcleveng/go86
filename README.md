# go86
Experimental 8086 emulator to learn the Go programming language.


# Running

Run the emulator using the go run command.  

Here's an example:

```
go run cmd/go86/go86.go -alsologtostderr run [path to 8086 executable]
```


# Using the debugger

go86 offers two types of debuggers:

* The lame debugger - this is a text based socket protocol designed for
  interactive human use.  It's available on port 2159 by default.

* GDB remote serial protocol debugger - This will implement the GDB remote
  serial protocol over a socket.  You may connect with the following:

  ```
  set debug remote 1
  target remote [host]:2159
  ```

  For more information see:
  https://sourceware.org/gdb/current/onlinedocs/gdb/Connecting.html

