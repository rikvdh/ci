package web

import (
	"encoding/json"
	"sync"

	"time"

	"github.com/Sirupsen/logrus"
	"github.com/go-iris2/iris2"
	"github.com/go-iris2/iris2/adaptors/websocket"
	"github.com/rikvdh/ci/lib/builder"
	"github.com/rikvdh/ci/lib/indexer"
	"github.com/rikvdh/ci/models"
)

type BuildList struct {
	Running []models.Job `json:"running"`
	Queue   []models.Job `json:"queue"`
}

func getBuildList() []byte {
	var msg BuildList

	models.Handle().Where("status = ?", models.StatusBusy).Order("start DESC").Find(&msg.Running)
	models.Handle().Where("status = ?", models.StatusNew).Order("start DESC").Find(&msg.Queue)

	data, _ := json.Marshal(msg)
	return data
}

type jsonData struct {
	Action          string `json:"action"`
	ID              int    `json:"id,omitempty"`
	CurrentPosition int64  `json:"current_position,omitempty"`
	Data            string `json:"data,omitempty"`
}

func startWs(app *iris2.Framework) {
	ws := websocket.New(websocket.Config{
		Endpoint:         "/ws",
		ClientSourcePath: "/ws.js",
	})

	app.Adapt(ws)

	var mu sync.RWMutex

	conns := make(map[string]websocket.Connection)

	go func() {
		ch := builder.GetEventChannel()
		for {
			<-ch

			data := getBuildList()
			logrus.Infof("Emit new build-list via websocket to %d connections", len(conns))
			mu.RLock()
			for _, con := range conns {
				con.EmitMessage(data)
			}
			mu.RUnlock()
		}
	}()

	ws.OnConnection(func(c websocket.Connection) {
		ctx := c.Context()
		if ctx.Session().GetString("authenticated") != "true" {
			go func() {
				// its a hack, but for some reason disconnecting here fails..
				time.Sleep(time.Second * 2)
				c.Disconnect()
			}()
			return
		}
		mu.Lock()
		conns[c.ID()] = c
		mu.Unlock()

		c.EmitMessage(getBuildList())

		c.OnMessage(func(d []byte) {
			var req jsonData
			json.Unmarshal(d, &req)
			switch req.Action {
			case "build":
				item := models.Branch{}
				models.Handle().Preload("Build").Where("id = ?", req.ID).First(&item)
				if item.ID > 0 {
					indexer.ScheduleJob(item.Build.ID, item.ID, item.LastReference)
				} else {
					logrus.Warnf("error, branch not found: %v", req)
				}
			case "logpos":
				item := models.Job{}
				models.Handle().First(&item, req.ID)
				if item.ID > 0 {
					buf := builder.GetLogFromPos(&item, req.CurrentPosition)
					rep := jsonData{
						Action:          req.Action,
						ID:              req.ID,
						Data:            buf,
						CurrentPosition: int64(len(buf)) + req.CurrentPosition,
					}
					rb, _ := json.Marshal(rep)
					c.EmitMessage(rb)
				} else {
					logrus.Warnf("logreader: %v", req)
				}
			}
		})

		c.OnDisconnect(func() {
			mu.Lock()
			delete(conns, c.ID())
			mu.Unlock()
		})
	})
}
