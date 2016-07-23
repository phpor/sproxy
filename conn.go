package sproxy


type stats struct {
	CurrentTaskNum int64 //当前正在执行的任务
	Stopping bool
}

var Stats *stats
func init()  {
	Stats = &stats{0, false}
}
