package progress

import (
	"fmt"
	"io"
	"strings"
	"sync"
)

// ProgressWriter 进度写入器
type ProgressWriter struct {
	io.Writer
	total    int64
	current  int64
	lastPerc int
	mu       sync.Mutex
	prefix   string
	suffix   string
}

// NewProgressWriter 创建新的进度写入器
func NewProgressWriter(w io.Writer, total int64, prefix, suffix string) *ProgressWriter {
	return &ProgressWriter{
		Writer: w,
		total:  total,
		prefix: prefix,
		suffix: suffix,
	}
}

// Write 实现 io.Writer 接口
func (pw *ProgressWriter) Write(p []byte) (int, error) {
	pw.mu.Lock()
	defer pw.mu.Unlock()

	n, err := pw.Writer.Write(p)
	if err != nil {
		return n, err
	}

	pw.current += int64(n)
	perc := int(float64(pw.current) / float64(pw.total) * 100)
	if perc != pw.lastPerc {
		pw.updateProgress(perc)
		pw.lastPerc = perc
	}
	return n, nil
}

// updateProgress 更新进度显示
func (pw *ProgressWriter) updateProgress(perc int) {
	// 清除当前行
	fmt.Print("\r" + strings.Repeat(" ", 80) + "\r")

	// 显示进度条
	barWidth := 30
	completed := int(float64(barWidth) * float64(perc) / 100)
	bar := strings.Repeat("=", completed) + strings.Repeat("-", barWidth-completed)

	// 构建进度信息
	progress := fmt.Sprintf("%s [%s] %d%% %s", pw.prefix, bar, perc, pw.suffix)
	fmt.Print("\r" + progress)
}

// ProgressBar 进度条
type ProgressBar struct {
	total   int64
	current int64
	mu      sync.Mutex
	prefix  string
	suffix  string
}

// NewProgressBar 创建新的进度条
func NewProgressBar(total int64, prefix, suffix string) *ProgressBar {
	return &ProgressBar{
		total:  total,
		prefix: prefix,
		suffix: suffix,
	}
}

// Update 更新进度
func (pb *ProgressBar) Update(n int64) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	pb.current += n
	perc := int(float64(pb.current) / float64(pb.total) * 100)

	// 清除当前行
	fmt.Print("\r" + strings.Repeat(" ", 80) + "\r")

	// 显示进度条
	barWidth := 30
	completed := int(float64(barWidth) * float64(perc) / 100)
	bar := strings.Repeat("=", completed) + strings.Repeat("-", barWidth-completed)

	// 构建进度信息
	progress := fmt.Sprintf("%s [%s] %d%% %s", pb.prefix, bar, perc, pb.suffix)
	fmt.Print("\r" + progress)
}

// Finish 完成进度条
func (pb *ProgressBar) Finish() {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	// 清除当前行
	fmt.Print("\r" + strings.Repeat(" ", 80) + "\r")

	// 显示完成信息
	bar := strings.Repeat("=", 30)
	progress := fmt.Sprintf("%s [%s] 100%% %s", pb.prefix, bar, pb.suffix)
	fmt.Print("\r" + progress + "\n")
}
