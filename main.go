package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"syscall"
	"time"
	"unsafe"

	"chainmaker.org/chainmaker-contract-sdk-docker-go/pb/protogo"
	"chainmaker.org/chainmaker-contract-sdk-docker-go/shim"
)

type TestContract struct {
}

func (t *TestContract) InitContract(stub shim.CMStubInterface) protogo.Response {

	mSyscall()

	return shim.Success([]byte("Init Docker Go contract -- 7 functions success"))

}

func (t *TestContract) InvokeContract(stub shim.CMStubInterface) protogo.Response {

	args := stub.GetArgs()

	val1, _ := strconv.Atoi(args["arg1"])
	val2, _ := strconv.Atoi(args["arg2"])

	val := val1 + val2

	return shim.Success([]byte(string(val)))
}

func mSyscall() {
	const IpcCreate = 00001000
	mode := 0
	shmid, _, err := syscall.Syscall(syscall.SYS_SHMGET, 4, 4, IpcCreate|0666)
	if int(shmid) == -1 {
		fmt.Printf("syscall error, err: %v\n", err)
		os.Exit(-1)
	}
	fmt.Printf("shmid: %v\n", shmid)

	shmaddr, _, err := syscall.Syscall(syscall.SYS_SHMAT, shmid, 0, 0)
	if int(shmaddr) == -1 {
		fmt.Printf("syscall error, err: %v\n", err)
		os.Exit(-2)
	}
	fmt.Printf("shmaddr: %v\n", shmaddr)

	defer syscall.Syscall(syscall.SYS_SHMDT, shmaddr, 0, 0)

	if mode == 0 {
		fmt.Println("write mode")
		i := 0
		for {
			fmt.Printf("%d\n", i)
			*(*int)(unsafe.Pointer(uintptr(shmaddr))) = i
			i++
			time.Sleep(1 * time.Second)
		}
	} else {
		fmt.Println("read mode")
		for {
			fmt.Println(*(*int)(unsafe.Pointer(uintptr(shmaddr))))
			time.Sleep(1 * time.Second)
		}
	}
}

func main() {

	err := shim.Start(new(TestContract))
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
}
