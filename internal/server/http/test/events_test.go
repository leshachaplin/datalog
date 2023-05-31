package http

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func (i *IntegrationTestSuite) TestLog_SendEvents() {
	ctx, mainCansel := context.WithTimeout(i.ctx, time.Minute*2)
	defer mainCansel()

	cases := map[string]struct {
		connectionsAmount int
		timeoutDuration   time.Duration
		expectedError     error
	}{
		"ok": {
			connectionsAmount: 200,
			timeoutDuration:   time.Second * 10,
		},
	}
	for name, tc := range cases {
		tt := tc
		i.Run(name, func() {
			c, cnsl := context.WithTimeout(ctx, tt.timeoutDuration)
			defer cnsl()

			wg := &sync.WaitGroup{}
			wg.Add(1)
			go func(ctx context.Context, wg *sync.WaitGroup) {
				defer wg.Done()
				tick := time.NewTicker(time.Second)
				defer tick.Stop()
				for {
					select {
					case <-ctx.Done():
						return
					case <-tick.C:
						for k := 0; k < tc.connectionsAmount; k++ {
							numEvents := i.Rand.Intn(100) + 30
							go func(c context.Context, en int) {
								t := http.DefaultTransport.(*http.Transport).Clone()
								t.MaxIdleConns = 100
								t.MaxConnsPerHost = 100
								t.MaxIdleConnsPerHost = 100
								eventCli := NewClient(fmt.Sprintf("http://localhost%s", defaultAddrPublic), &http.Client{
									Transport: t,
									Timeout:   time.Second * 30,
								})

								batch := make([]eventReq, en)
								for j := 0; j < en; j++ {
									payload := eventReq{
										ClientTime: time.Now().Format(time.DateTime),
										DeviceID:   uuid.NewString(),
										DeviceOS:   uuid.NewString(),
										Session:    strconv.Itoa(j),
										Event:      "app_open",
										ParamStr:   "test",
										Sequence:   j,
										ParamInt:   j,
									}
									batch[j] = payload
								}
								err := eventCli.SendEvents(c, batch)
								if err != nil {
									log.Err(err).Send()
								}
							}(ctx, numEvents)
						}
					}
				}
			}(c, wg)
			wg.Wait()
			i.app.Stop()
		})
	}
}
