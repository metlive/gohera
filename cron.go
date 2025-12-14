package gohera

import (
	"context"
	"strings"
	"sync"

	"github.com/robfig/cron/v3"
)

//执行任务格式举例         			 格式
//-----------------					 --------------
//每隔5秒执行一次：					 */5 * * * * ?
//每隔1分钟执行一次：					 0 */1 * * * ?
//每天23点执行一次：					 0 0 23 * * ?
//每天凌晨1点执行一次：					 0 0 1 * * ?
//每月1号凌晨1点执行一次：				 0 0 1 1 * ?
//每周一和周三晚上22:30: 				 00 30 22 * * 1,3
//在26分、29分、33分执行一次：			 0 26,29,33 * * * ?
//每天的0点、13点、18点、21点都执行一次：  0 0 0,13,18,21 * * ?
//每年三月的星期四的下午14:10和14:40:	 00 10,40 14 ? 3 4

//Field name     Mandatory?   Allowed values    Allowed special characters
//----------     ----------   --------------    --------------------------
//Seconds        Yes           0-59             * / , -
//Minutes        Yes          0-59              * / , -
//Hours          Yes          0-23              * / , -
//Day of month   Yes          1-31              * / , - L W
//Month          Yes          1-12 or JAN-DEC   * / , -
//Day of week    Yes          0-6 or SUN-SAT    * / , - L #
//Year           No           1970–2099         * / , -

type HandlerFunc func()

type Manager struct {
	run          *cron.Cron
	sepc         string
	jobName      string
	jobFunc      HandlerFunc
	handlerMutex sync.Mutex
}

// NewJobManager 创建定时任务管理器
func NewJobManager() *Manager {
	return &Manager{
		run: cron.New(cron.WithSeconds()),
	}
}

// Command 设置当前任务的名称和处理函数
func (m *Manager) Command(jobName string, jobFunc func()) *Manager {
	m.handlerMutex.Lock()
	defer m.handlerMutex.Unlock()
	m.jobName = jobName
	m.jobFunc = jobFunc
	return m
}

func (m *Manager) schedule(sepc string, jobName string, jobFunc func()) {
	ctx := context.Background()
	_, err := m.run.AddFunc(sepc, jobFunc)
	if err != nil {
		Errortf(ctx, "schedule %s run error", jobName)
	}
	Infotf(ctx, "schedule job name %v", jobName)
}

// Stop 停止定时任务
func (m *Manager) Stop() {
	m.run.Stop()
}

// Start 启动定时任务
func (m *Manager) Start() {
	m.run.Start()
}

// 每五秒执行
func (m *Manager) EveryFiveSeconds() {
	m.schedule("*/5 * * * * *", m.jobName, m.jobFunc)
}

// 每十秒执行
func (m *Manager) EveryTenSeconds() {
	m.schedule("*/10 * * * * *", m.jobName, m.jobFunc)
}

// 每秒执行
func (m *Manager) EverySeconds() {
	m.schedule("*/1 * * * * *", m.jobName, m.jobFunc)
}

// 每分钟执行
func (m *Manager) EveryMinutes() {
	m.schedule("0 */1 * * * *", m.jobName, m.jobFunc)
}

// EveryFiveMinutes 每五分钟执行
func (m *Manager) EveryFiveMinutes() {
	m.schedule("0 */5 * * * *", m.jobName, m.jobFunc)
}

// EveryTenMinutes 每十分钟执行
func (m *Manager) EveryTenMinutes() {
	m.schedule("0 */10 * * * *", m.jobName, m.jobFunc)
}

// EveryThirtyMinutes 每三十分钟执行
func (m *Manager) EveryThirtyMinutes() {
	m.schedule("0 */30 * * * *", m.jobName, m.jobFunc)
}

// Hourly 每小时执行 (XX:00)
func (m *Manager) Hourly() {
	m.schedule("0 0 * * * *", m.jobName, m.jobFunc)
}

// Daily 每天执行 (00:00)
func (m *Manager) Daily() {
	m.schedule("0 0 0 * * *", m.jobName, m.jobFunc)
}

// Weekly 每周执行 (周日 00:00)
func (m *Manager) Weekly() {
	m.schedule("0 0 0 * * 0", m.jobName, m.jobFunc)
}

// Monthly 每月执行 (1号 00:00)
func (m *Manager) Monthly() {
	m.schedule("0 0 0 1 * *", m.jobName, m.jobFunc)
}

// Yearly 每年执行 (1月1号 00:00)
func (m *Manager) Yearly() {
	m.schedule("0 0 0 1 1 *", m.jobName, m.jobFunc)
}

// Cron 使用自定义 Cron 表达式执行
func (m *Manager) Cron(sepc string) {
	m.schedule(sepc, m.jobName, m.jobFunc)
}

// DailyAt 每天在指定时间执行 (格式 HH:mm)
func (m *Manager) DailyAt(time string) {
	segments := strings.Split(time, ":")
	hour := segments[0]
	minute := "*"
	if len(segments) == 2 {
		minute = segments[1]
	}
	m.schedule("0 "+minute+" "+hour+" * * *", m.jobName, m.jobFunc)
}
