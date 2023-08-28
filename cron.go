package gohera

import (
	"context"
	"github.com/robfig/cron/v3"
	"strings"
	"sync"
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

func NewJobManager() *Manager {
	return &Manager{
		run: cron.New(cron.WithSeconds()),
	}
}

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

func (m *Manager) Stop() {
	m.run.Stop()
}

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

func (m *Manager) EveryFiveMinutes() {
	m.schedule("0 */5 * * * *", m.jobName, m.jobFunc)
}

func (m *Manager) EveryTenMinutes() {
	m.schedule("0 */10 * * * *", m.jobName, m.jobFunc)
}

func (m *Manager) EveryThirtyMinutes() {
	m.schedule("0 */30 * * * *", m.jobName, m.jobFunc)
}

func (m *Manager) Hourly() {
	m.schedule("0 0 * * * *", m.jobName, m.jobFunc)
}

func (m *Manager) Daily() {
	m.schedule("0 0 0 * * *", m.jobName, m.jobFunc)
}

func (m *Manager) Weekly() {
	m.schedule("0 0 0 * * 0", m.jobName, m.jobFunc)
}

func (m *Manager) Monthly() {
	m.schedule("0 0 0 1 * *", m.jobName, m.jobFunc)
}

func (m *Manager) Yearly() {
	m.schedule("0 0 0 1 1 *", m.jobName, m.jobFunc)
}

func (m *Manager) Cron(sepc string) {
	m.schedule(sepc, m.jobName, m.jobFunc)
}

func (m *Manager) DailyAt(time string) {
	segments := strings.Split(time, ":")
	hour := segments[0]
	var minute = "*"
	if len(segments) == 2 {
		minute = segments[1]
	}
	m.schedule("0 "+minute+" "+hour+" * * *", m.jobName, m.jobFunc)
}
