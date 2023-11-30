package panasonic_irc_cgo

// #include "hello.c"
import "C"

func main() {
	C.myCFunction()
}

// #include <stdio.h>
// #include <stdlib.h>
// void myCFunction() {
//   printf("Hello from C code!\n");
// }
