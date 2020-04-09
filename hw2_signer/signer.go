package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var sortslice = make([]string, 100)
var jobchan = make(chan interface{}, 100)

var result string

func main() {
	inputData := []int{0, 1, 1, 2, 3, 5, 8}
	//inputData := []int{0, 1}

	hashSignJobs := []job{
		job(func(in, out chan interface{}) {
			for _, fibNum := range inputData {
				out <- fibNum
			}
		}),
		job(SingleHash),
		job(MultiHash),
		job(CombineResults),
		/*job(func(in, out chan interface{}) {
		  dataRaw := <-in
		  data, ok := dataRaw.(string)
		  if !ok {
		  fmt.Print("cant convert result data to string")
		  }
		  fmt.Print(data)
		  }),*/
	}
	//timer:=time.NewTicker(time.Second)
	//done:=make(chan struct{})
	go func() {
		for i := 1; ; {
			time.Sleep(time.Second)
			fmt.Println("tick ", i)
			i++
			/*
			   select {
			   case t:=<-timer.C:
			   fmt.Println("tick ",t)
			   case <-done:
			   return
			   }*/
		}
	}()

	ExecutePipeline(hashSignJobs...)
	//done<-struct{}{}
	//timer.Stop()
	dataRaw := <-jobchan
	data, ok := dataRaw.(string)
	if !ok {
		fmt.Print("cant convert result data to string")
	}
	fmt.Print(data)

}
func ExecutePipeline(jobs ...job) {

	jonchanin := jobchan
	jonchanout := jobchan
	for _, worker := range jobs {
		jonchanout = make(chan interface{}, 100)
		go worker(jonchanin, jonchanout)
		jonchanin = jonchanout
	}
	jobchan= jonchanout
}

/*func ExecutePipeline(jobs ...job) {
	if len(jobs) > 0 {
		outchannel:= make(chan interface{}, 100)
		jobchanin := jobchan
		jobchanout := jobchan
		go jobs[0](jobchanin, jobchanout)
		for data := range jobchanout {
			jobchanin = make(chan interface{}, 100)
			jobchanin <- data
			for i, worker := range jobs[1 : len(jobs)-1] {
				jobchanout = make(chan interface{}, 100)
				if i==len(jobs)-1{
					go worker(jobchanin, outchannel)
				} else {
					go worker(jobchanin, jobchanout)
					jobchanin = jobchanout
				}

			}
		}

		go jobs[len(jobs)-1](outchannel, jobchan)
		//jobchan = jobchanout

	}

}*/

/*
type HasMD5 struct {}
type Hash32 struct {}
func (h *Hash32) hash (data string) string {

}
type hasher interface {
hash(string) string
}
*/

//var int interlocked
var mux sync.Mutex

func GoMD5(data string) string {
	mux.Lock()
	md5:=DataSignerMd5(data)
	mux.Unlock()
	return md5
}

func Gofunc(chin chan interface{} ,data string, hasher func(string) string)  {
	go func(data string, hasher func(string) string) {
		chin <- hasher(data)
	}(data, hasher)
}

var SingleHash = func(in, out chan interface{}) {
//LOOP:
	for {
		select {
		case <-time.After(time.Second):
			//break LOOP
		case data := <-in:
			element := strconv.Itoa(data.(int))
			go func(){
				fmt.Println("start SingleHash for:",element)
				chanout:= make(chan interface{}, 2)
				Gofunc(chanout,element,DataSignerCrc32)
				Gofunc(chanout,GoMD5(element),DataSignerCrc32)
				out <-(<-chanout).(string)+(<-chanout).(string)
				close(chanout)
				//out <- (DataSignerCrc32(element) + "~" + DataSignerCrc32(GoMD5(element)))
				//out <- (Gofunc(element,DataSignerCrc32) + "~" + Gofunc(GoMD5(element),DataSignerCrc32))
				fmt.Println("finish SingleHash for:",element)
			}()
		}
	}
	close(out)
}

var MultiHash = func(in, out chan interface{}) {
	for data := range in {
		mapa:=make(map[int]string)
		mux:=&sync.Mutex{}
		var hash string
		chanout:= make(chan interface{}, 6)
		fmt.Println("start MultiHash for:",data)

		for i := 0; i < 6; i++ {
			go func(){
				Gofunc(chanout,(strconv.Itoa(i) + data.(string)),DataSignerCrc32)
			}()
		}
		for {
			select {
			case <-time.After(time.Second):
				//break LOOP
			case data := <-chanout:
				element := data.(string)

				go func(){
					fmt.Println("start SingleHash for:",element)
					chanout:= make(chan interface{}, 2)
					Gofunc(chanout,element,DataSignerCrc32)
					Gofunc(chanout,GoMD5(element),DataSignerCrc32)
					out <-(<-chanout).(string)+(<-chanout).(string)
					close(chanout)
					//out <- (DataSignerCrc32(element) + "~" + DataSignerCrc32(GoMD5(element)))
					//out <- (Gofunc(element,DataSignerCrc32) + "~" + Gofunc(GoMD5(element),DataSignerCrc32))
					fmt.Println("finish SingleHash for:",element)
				}()
			}
		}
		fmt.Println("finish MultiHash for:",data)
		out <- hash
	}
	close(out)
}
var CombineResults = func(in, out chan interface{}) {
	var slicenum int
	for data := range in {
		sortslice[slicenum] = data.(string)
		slicenum++
	}
	sort.StringSlice(sortslice[:slicenum]).Sort()
	out <- strings.Join(sortslice[:slicenum], "_")
}
