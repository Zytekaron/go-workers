# go-workers
```
go get github.com/Zytekaron/go-workers
```
Support: [Contact Me](https://zytekaron.com/contact/)

## Method Overview

### Creation

`NewPool(size int, run RunFunc)` Create a new WorkerPool with an initial worker count <br>
`NewBufferedPool(size int, bufSize int, run RunFunc)` Create a new WorkerPool with an initial worker count and job buffer size

`RunFunc` = `func(interface{})`

### Basic usage

`Pool#Size()` Get the total number of workers in this WorkerPool <br>
`Pool#Busy()` Get the number of busy workers in this WorkerPool <br>
`Pool#Run(data interface{})` Add a job to this WorkerPool

### Scaling

`Pool#ScaleUp(newSize int)` Scale the WorkerPool up to a new specified size <br>
`Pool#ScaleUp(newSize int)` Scale the WorkerPool down to a new specified size

## Examples

Example usage can be found [here](example/main.go)

## License
<b>go-workers</b> is licensed under the [MIT License](https://github.com/Zytekaron/go-workers/blob/master/LICENSE)
