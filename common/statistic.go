package common

import (
	"sync"
	"time"
)

type Statistic struct {
	rx, tx, count int
	timestamp     time.Time
}

type aggregatedStatistic struct {
	Statistic
	duration time.Duration
}

func (s *Statistic) add(new Statistic) {
	s.count += new.count
	s.rx += new.rx
	s.tx += new.tx
	if new.timestamp.After(s.timestamp) {
		s.timestamp = new.timestamp
	}
}

type statisticWindow struct {
	aggregatedStatistic
	startTime int64
	endTime   int64
	mux       *sync.Mutex
}

func getStatisticWindow(startTime time.Time, duration time.Duration) statisticWindow {
	win := statisticWindow{
		startTime: startTime.Unix(),
		endTime:   startTime.Add(duration).Unix(),
		mux:       &sync.Mutex{},
	}
	win.rx = 0
	win.tx = 0
	win.count = 0
	win.duration = duration
	win.timestamp = startTime
	return win
}

func (sw *statisticWindow) aggregateStatistic(statistic Statistic) {
	if sw.startTime <= statistic.timestamp.Unix() && statistic.timestamp.Unix() < sw.endTime {
		sw.mux.Lock()
		sw.add(statistic)
		sw.mux.Unlock()
	}
}

func (sw statisticWindow) getFinalResult() aggregatedStatistic {
	statistic := aggregatedStatistic{
		duration: sw.duration,
	}
	statistic.rx = sw.rx
	statistic.tx = sw.tx
	statistic.timestamp = sw.timestamp
	statistic.count = sw.count
	return statistic
}

func (sw statisticWindow) getTotal() (rx, tx int) {
	var crx, ctx int
	sw.mux.Lock()
	crx = sw.rx
	ctx = sw.tx
	sw.mux.Unlock()
	return crx, ctx
}

func (sw statisticWindow) getAverageByTime() (rx, tx int) {
	var arx, atx int
	sw.mux.Lock()
	arx = sw.rx / int(sw.startTime-sw.timestamp.Unix())
	atx = sw.tx / int(sw.startTime-sw.timestamp.Unix())
	sw.mux.Unlock()
	return arx, atx
}

func (sw statisticWindow) getAverageByCount() (rx, tx int) {
	var arx, atx int
	sw.mux.Lock()
	arx = sw.rx / sw.count
	atx = sw.tx / sw.count
	sw.mux.Unlock()
	return arx, atx
}

type TunnelStatistic struct {
	history        map[int64]aggregatedStatistic
	name           string
	windowDuration time.Duration
	currentWin     statisticWindow
	mux            *sync.Mutex
}

func GetTunnelStatistic(name string, duration time.Duration, historyDuration time.Duration) *TunnelStatistic {
	ts := TunnelStatistic{
		history:        make(map[int64]aggregatedStatistic),
		name:           name,
		windowDuration: duration,
		mux:            &sync.Mutex{},
	}
	historySize := int(historyDuration.Seconds() / ts.windowDuration.Seconds())
	ts.currentWin = getStatisticWindow(time.Now(), ts.windowDuration)
	go func() {
		for range time.Tick(ts.windowDuration) {
			ts.mux.Lock()
			c := ts.currentWin
			if len(ts.history) >= historySize {
				var t int64
				for t, _ = range ts.history {
					break
				}
				delete(ts.history, t)
			}
			ts.history[c.startTime] = c.getFinalResult()
			ts.currentWin = getStatisticWindow(time.Now(), ts.windowDuration)
			ts.mux.Unlock()
		}
	}()
	return &ts
}

func (ts *TunnelStatistic) AddStatistic(statistic Statistic) {
	ts.mux.Lock()
	ts.currentWin.add(statistic)
	ts.mux.Unlock()
}

func (ts TunnelStatistic) GetCurrentTotal() (rx, tx int) {
	var crx, ctx int
	ts.mux.Lock()
	crx, ctx = ts.currentWin.getTotal()
	ts.mux.Unlock()
	return crx, ctx
}

func (ts TunnelStatistic) GetCurrentAverageByTime() (rx, tx int) {
	var arx, atx int
	ts.mux.Lock()
	arx, atx = ts.currentWin.getAverageByTime()
	ts.mux.Unlock()
	return arx, atx
}

func (ts TunnelStatistic) GetCurrentAverageByCount() (rx, tx int) {
	var arx, atx int
	ts.mux.Lock()
	arx, atx = ts.currentWin.getAverageByCount()
	ts.mux.Unlock()
	return arx, atx
}

func (ts TunnelStatistic) GetTotal(start, end time.Time) (arx, atx, count int) {
	arx = 0
	atx = 0
	count = 0
	startUnix := start.Unix()
	endUnix := end.Unix()
	for t, s := range ts.history {
		if startUnix <= t && t < endUnix {
			arx += s.rx
			atx += s.tx
			count += s.count
		}
	}
	return arx, atx, count
}

func (ts TunnelStatistic) GetAverageByTime(start, end time.Time) (arx, atx int) {
	arx = 0
	atx = 0
	startUnix := start.Unix()
	endUnix := end.Unix()
	for t, s := range ts.history {
		if startUnix <= t && t < endUnix {
			arx += s.rx
			atx += s.tx
		}
	}
	arx = arx / int((startUnix-endUnix)/int64(ts.windowDuration.Seconds()))
	atx = atx / int((startUnix-endUnix)/int64(ts.windowDuration.Seconds()))
	return arx, atx
}

func (ts TunnelStatistic) GetAverageByCount(start, end time.Time) (arx, atx int) {
	arx = 0
	atx = 0
	count := 0
	startUnix := start.Unix()
	endUnix := end.Unix()
	for t, s := range ts.history {
		if startUnix <= t && t < endUnix {
			arx += s.rx
			atx += s.tx
			count += s.count
		}
	}
	arx = arx / count
	atx = atx / count
	return arx, atx
}
