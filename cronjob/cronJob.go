package cronjob

import (
	"time"

	"github.com/robfig/cron/v3"
)

type CronJob struct {
	cron *cron.Cron
}

func NewCronJob() *CronJob {
	return &CronJob{}
}

func (c *CronJob) Start(loc *time.Location, trafficAge int) error {
	c.cron = cron.New(cron.WithLocation(loc), cron.WithSeconds())

	if _, err := c.cron.AddJob("@every 10s", NewStatsJob(trafficAge > 0)); err != nil {
		return err
	}
	if _, err := c.cron.AddJob("@every 1m", NewDepleteJob()); err != nil {
		return err
	}
	if trafficAge > 0 {
		if _, err := c.cron.AddJob("@daily", NewDelStatsJob(trafficAge)); err != nil {
			return err
		}
	}
	if _, err := c.cron.AddJob("@every 5s", NewCheckCoreJob()); err != nil {
		return err
	}
	if _, err := c.cron.AddJob("@every 10m", NewWALCheckpointJob()); err != nil {
		return err
	}

	c.cron.Start()
	return nil
}

func (c *CronJob) Stop() {
	if c.cron != nil {
		c.cron.Stop()
	}
}
