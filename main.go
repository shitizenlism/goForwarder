package main

import (
    "encoding/json"
    "fmt"
    "os"
    "io/ioutil"

    "bufio"
    "context"
    "io"
    "os/exec"
    "sync"
    "time"
)

type StreamConfig struct {
	Interval int
	DvrStreams []string
	CdnStreams []string
	DvrPushApp []string
	CdnPushApp []string
}

func main() {
	jsonFile,err := os.Open("config.json")
	
	if err != nil{
		fmt.Println("Fail to open config.json. err=",err)
		return
	}
	defer jsonFile.Close()

	data,err := ioutil.ReadAll(jsonFile)
	if err != nil{
		fmt.Println("Fail to read config.json in. err=", err)
		return
	}

	var sc StreamConfig
	err = json.Unmarshal(data, &sc)
	if err != nil {
		fmt.Printf("err=%v. exit!\n", err)
		return
	}
	//fmt.Println(sc.DvrStreams)
	//fmt.Println(sc.DvrPushApp)
	//fmt.Println(sc.CdnStreams)
	//fmt.Println(sc.CdnPushApp)

	for {
		for i,s := range(sc.DvrStreams){
		    fmt.Println("checking ",s)
		    ctx, cancel := context.WithCancel(context.Background())
		    go func(cancelFunc context.CancelFunc) {
		        time.Sleep(8 * time.Second)
		        cancelFunc()
		        //fmt.Println("timeout. checking process quit.")
		    }(cancel)
		    Command(ctx, "ffprobe " + s, sc.DvrPushApp[i])
		    time.Sleep(1*time.Second)
		}

		for i,s := range(sc.CdnStreams){
		    fmt.Println("checking ",s)
		    ctx, cancel := context.WithCancel(context.Background())
		    go func(cancelFunc context.CancelFunc) {
		        time.Sleep(8 * time.Second)
		        cancelFunc()
		        //fmt.Println("timeout. checking process quit.")
		    }(cancel)
		    Command(ctx, "ffprobe " + s, sc.CdnPushApp[i])
		    time.Sleep(1*time.Second)
		}

		fmt.Printf("all streams checking is done. next check will be after %d s\n", sc.Interval)
		time.Sleep(time.Second * time.Duration(sc.Interval))
	}
}

func RunShellCmd(appName string){
    cmd := exec.Command("supervisorctl", "restart", appName)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    if err := cmd.Run(); err != nil{
    	fmt.Println(err)
    }
}

func read(ctx context.Context, wg *sync.WaitGroup, std io.ReadCloser, appName string) {
    reader := bufio.NewReader(std)
    defer wg.Done()
    for {
        select {
        case <-ctx.Done():
            return
        default:
            readString, err := reader.ReadString('\n')
            if err != nil || err == io.EOF {
 				fmt.Println("cmd-pipe exit. err=",ctx.Err())
                if ctx.Err() == nil{
                    fmt.Println("check succeed.")
                } else {
                    fmt.Println("stream is gone. check aborted. restart publisher...")
                    RunShellCmd(appName)
                }
                return
            }
            fmt.Print(readString)
        }
    }
}

func Command(ctx context.Context, cmd string, appName string) error {
    //c := exec.CommandContext(ctx, "cmd", "/C", cmd) // windows
    c := exec.CommandContext(ctx, "bash", "-c", cmd) // mac linux
    //stdout, err := c.StdoutPipe()
    //if err != nil {
    //    return err
    //}
     stderr, err := c.StderrPipe()
     if err != nil {
         return err
     }

    var wg sync.WaitGroup

    wg.Add(1)
    //go read(ctx, &wg, stdout, appName)
     go read(ctx, &wg, stderr, appName)

    err = c.Start()

    wg.Wait()
    return err
}
