package main

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
)

func main() {
	// 定义要运行的命令，这里是运行一个 Go 程序 "code_user/main.go"
	cmd := exec.Command("go", "run", "code_user/main.go")

	// 定义缓冲区，用于存储命令的标准输出和标准错误
	var out, stdErr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stdErr

	// 创建标准输入管道，以便向子进程的标准输入写入数据
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		fmt.Println(err)
		return
	}

	// 向子进程的标准输入写入数据，这里是字符串 "23 11\n"
	io.WriteString(stdinPipe, "23 11\n")

	// 运行命令并等待其完成
	if err := cmd.Run(); err != nil {
		// 如果运行命令时发生错误，打印错误和标准错误输出
		fmt.Println(err, stdErr.String())
		return
	}

	// 打印命令的标准输出
	fmt.Println(out.String())
}
